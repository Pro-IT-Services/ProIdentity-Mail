package backup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LiveMapping struct {
	Source string
	Target string
}

type CommandOptions struct {
	Env       []string
	StdinPath string
}

type CommandRunner interface {
	Run(ctx context.Context, name string, args []string, options CommandOptions) error
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args []string, options CommandOptions) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, options.Env...)
	if options.StdinPath != "" {
		input, err := os.Open(options.StdinPath)
		if err != nil {
			return err
		}
		defer input.Close()
		cmd.Stdin = input
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

type LiveRestoreOptions struct {
	ArchivePath      string
	StagingDir       string
	EncryptionKey    []byte
	Mappings         []LiveMapping
	Permissions      []PermissionRule
	DatabaseName     string
	DatabasePassword string
	ImportDatabase   bool
	ControlServices  bool
	Services         []string
	Runner           CommandRunner
	Overwrite        bool
	KeepStaging      bool
}

type PermissionRule struct {
	Path      string
	Owner     string
	Group     string
	DirMode   os.FileMode
	FileMode  os.FileMode
	Recursive bool
}

type LiveRestoreSummary struct {
	Files            int
	Bytes            int64
	Entries          int
	StagingDir       string
	DatabaseImported bool
	PermissionsFixed bool
	ServicesStopped  bool
	ServicesStarted  bool
}

func RestoreLive(ctx context.Context, options LiveRestoreOptions) (LiveRestoreSummary, error) {
	if strings.TrimSpace(options.ArchivePath) == "" {
		return LiveRestoreSummary{}, errors.New("archive path is required")
	}
	runner := options.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	stagingDir := options.StagingDir
	var cleanup func()
	if strings.TrimSpace(stagingDir) == "" {
		temp, err := os.MkdirTemp("", "proidentity-restore-*")
		if err != nil {
			return LiveRestoreSummary{}, err
		}
		stagingDir = temp
		cleanup = func() {
			if !options.KeepStaging {
				_ = os.RemoveAll(temp)
			}
		}
	} else if err := os.MkdirAll(stagingDir, 0750); err != nil {
		return LiveRestoreSummary{}, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	summary, err := VerifyWithKey(ctx, options.ArchivePath, options.EncryptionKey)
	if err != nil {
		return LiveRestoreSummary{}, err
	}
	if err := RestoreWithKey(ctx, options.ArchivePath, stagingDir, RestoreOptions{Overwrite: true}, options.EncryptionKey); err != nil {
		return LiveRestoreSummary{}, err
	}
	out := LiveRestoreSummary{Files: summary.Files, Bytes: summary.Bytes, Entries: summary.Entries, StagingDir: stagingDir}

	if options.ControlServices && len(options.Services) > 0 {
		if err := runner.Run(ctx, "systemctl", append([]string{"stop"}, options.Services...), CommandOptions{}); err != nil {
			return out, fmt.Errorf("stop services: %w", err)
		}
		out.ServicesStopped = true
	}
	for _, mapping := range options.Mappings {
		if err := restoreMapping(stagingDir, mapping, options.Overwrite); err != nil {
			return out, err
		}
	}
	if options.ImportDatabase {
		if options.DatabaseName == "" {
			return out, errors.New("database name is required")
		}
		dumpPath := filepath.Join(stagingDir, "database", "proidentity.sql")
		if _, err := os.Stat(dumpPath); err != nil {
			return out, fmt.Errorf("database dump missing: %w", err)
		}
		env := []string(nil)
		if options.DatabasePassword != "" {
			env = append(env, "MYSQL_PWD="+options.DatabasePassword)
		}
		if err := runner.Run(ctx, "mariadb", []string{"--database", options.DatabaseName}, CommandOptions{Env: env, StdinPath: dumpPath}); err != nil {
			return out, fmt.Errorf("import database: %w", err)
		}
		out.DatabaseImported = true
	}
	if len(options.Permissions) > 0 {
		if err := applyPermissions(ctx, runner, options.Permissions); err != nil {
			return out, err
		}
		out.PermissionsFixed = true
	}
	if options.ControlServices && len(options.Services) > 0 {
		if err := runner.Run(ctx, "systemctl", append([]string{"restart"}, options.Services...), CommandOptions{}); err != nil {
			return out, fmt.Errorf("restart services: %w", err)
		}
		out.ServicesStarted = true
	}
	return out, nil
}

func restoreMapping(stagingDir string, mapping LiveMapping, overwrite bool) error {
	if strings.TrimSpace(mapping.Source) == "" || strings.TrimSpace(mapping.Target) == "" {
		return errors.New("restore mapping source and target are required")
	}
	sourceRel, err := safeRelative(mapping.Source)
	if err != nil {
		return err
	}
	sourcePath, err := safeTarget(stagingDir, sourceRel)
	if err != nil {
		return err
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !overwrite {
		if _, err := os.Stat(mapping.Target); err == nil {
			return fmt.Errorf("target exists: %s", mapping.Target)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if info.IsDir() {
		return copyDir(sourcePath, mapping.Target, overwrite)
	}
	return copyFile(sourcePath, mapping.Target, info, overwrite)
}

func copyDir(source, target string, overwrite bool) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return copyFile(path, targetPath, info, overwrite)
	})
}

func copyFile(source, target string, info os.FileInfo, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(target); err == nil {
			return fmt.Errorf("target exists: %s", target)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Chmod(target, info.Mode().Perm())
}

func applyPermissions(ctx context.Context, runner CommandRunner, rules []PermissionRule) error {
	for _, rule := range rules {
		if strings.TrimSpace(rule.Path) == "" {
			return errors.New("permission path is required")
		}
		if rule.Owner != "" || rule.Group != "" {
			ownerGroup := rule.Owner
			if rule.Group != "" {
				ownerGroup += ":" + rule.Group
			}
			args := []string(nil)
			if rule.Recursive {
				args = append(args, "-R")
			}
			args = append(args, ownerGroup, rule.Path)
			if err := runner.Run(ctx, "chown", args, CommandOptions{}); err != nil {
				return fmt.Errorf("chown %s: %w", rule.Path, err)
			}
		}
		if rule.DirMode != 0 || rule.FileMode != 0 {
			if err := chmodRule(rule); err != nil {
				return err
			}
		}
	}
	return nil
}

func chmodRule(rule PermissionRule) error {
	info, err := os.Stat(rule.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !rule.Recursive {
		mode := rule.FileMode
		if info.IsDir() {
			mode = rule.DirMode
		}
		if mode == 0 {
			return nil
		}
		return os.Chmod(rule.Path, mode.Perm())
	}
	return filepath.WalkDir(rule.Path, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			if rule.DirMode != 0 {
				return os.Chmod(path, rule.DirMode.Perm())
			}
			return nil
		}
		if info.Mode().IsRegular() && rule.FileMode != 0 {
			return os.Chmod(path, rule.FileMode.Perm())
		}
		return nil
	})
}
