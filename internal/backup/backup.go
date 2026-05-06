package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const manifestPath = "proidentity-backup-manifest.json"

type Options struct {
	OutputPath string
	Sources    []Source
	Hostname   string
}

type Source struct {
	Name     string
	Path     string
	Required bool
}

type Manifest struct {
	Version   int             `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	Hostname  string          `json:"hostname,omitempty"`
	GOOS      string          `json:"goos"`
	Entries   []ManifestEntry `json:"entries"`
}

type ManifestEntry struct {
	Source string      `json:"source"`
	Path   string      `json:"path"`
	Type   string      `json:"type"`
	Size   int64       `json:"size"`
	SHA256 string      `json:"sha256,omitempty"`
	Mode   os.FileMode `json:"mode"`
}

type VerifySummary struct {
	Files   int
	Bytes   int64
	Entries int
}

type RestoreOptions struct {
	Overwrite bool
}

func Create(ctx context.Context, options Options) (Manifest, error) {
	if strings.TrimSpace(options.OutputPath) == "" {
		return Manifest{}, errors.New("output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(options.OutputPath), 0750); err != nil {
		return Manifest{}, err
	}
	output, err := os.Create(options.OutputPath)
	if err != nil {
		return Manifest{}, err
	}
	defer output.Close()
	gzipWriter := gzip.NewWriter(output)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	manifest := Manifest{Version: 1, CreatedAt: time.Now().UTC(), Hostname: options.Hostname, GOOS: runtime.GOOS}
	for _, source := range options.Sources {
		select {
		case <-ctx.Done():
			return Manifest{}, ctx.Err()
		default:
		}
		if err := addSource(ctx, tarWriter, &manifest, source); err != nil {
			return Manifest{}, err
		}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Manifest{}, err
	}
	if err := tarWriter.WriteHeader(&tar.Header{Name: manifestPath, Mode: 0640, Size: int64(len(data)), ModTime: time.Now().UTC()}); err != nil {
		return Manifest{}, err
	}
	if _, err := tarWriter.Write(data); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func Verify(ctx context.Context, archivePath string) (VerifySummary, error) {
	manifest, files, err := readArchive(ctx, archivePath)
	if err != nil {
		return VerifySummary{}, err
	}
	if manifest.Version != 1 {
		return VerifySummary{}, fmt.Errorf("unsupported manifest version %d", manifest.Version)
	}
	fileByPath := make(map[string]archiveFile, len(files))
	for _, file := range files {
		fileByPath[file.Path] = file
	}
	var summary VerifySummary
	for _, entry := range manifest.Entries {
		if entry.Type != "file" {
			summary.Entries++
			continue
		}
		file, ok := fileByPath[entry.Path]
		if !ok {
			return VerifySummary{}, fmt.Errorf("missing archive entry %s", entry.Path)
		}
		if file.Size != entry.Size || file.SHA256 != entry.SHA256 {
			return VerifySummary{}, fmt.Errorf("hash mismatch for %s", entry.Path)
		}
		summary.Files++
		summary.Bytes += entry.Size
		summary.Entries++
	}
	return summary, nil
}

func Restore(ctx context.Context, archivePath, targetRoot string, options RestoreOptions) error {
	if _, err := Verify(ctx, archivePath); err != nil {
		return err
	}
	if strings.TrimSpace(targetRoot) == "" {
		return errors.New("target root is required")
	}
	reader, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header.Name == manifestPath {
			continue
		}
		target, err := safeTarget(targetRoot, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if !options.Overwrite {
				if _, err := os.Stat(target); err == nil {
					return fmt.Errorf("target exists: %s", header.Name)
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}
			if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, tarReader)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
	}
}

type archiveFile struct {
	Path   string
	Size   int64
	SHA256 string
}

func addSource(ctx context.Context, tarWriter *tar.Writer, manifest *Manifest, source Source) error {
	if source.Name == "" {
		return errors.New("source name is required")
	}
	info, err := os.Stat(source.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !source.Required {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return filepath.WalkDir(source.Path, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if path == source.Path {
				return nil
			}
			return addPath(tarWriter, manifest, source, path)
		})
	}
	return addPath(tarWriter, manifest, source, source.Path)
}

func addPath(tarWriter *tar.Writer, manifest *Manifest, source Source, path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	rel, err := filepath.Rel(source.Path, path)
	if err != nil || rel == "." {
		rel = filepath.Base(path)
	}
	name := filepath.ToSlash(filepath.Join(source.Name, rel))
	if info.IsDir() {
		manifest.Entries = append(manifest.Entries, ManifestEntry{Source: source.Name, Path: name, Type: "dir", Mode: info.Mode()})
		return tarWriter.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeDir, Mode: int64(info.Mode().Perm()), ModTime: info.ModTime()})
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	header := &tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: int64(info.Mode().Perm()), Size: info.Size(), ModTime: info.ModTime()}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if _, err := io.Copy(tarWriter, io.TeeReader(file, hash)); err != nil {
		return err
	}
	manifest.Entries = append(manifest.Entries, ManifestEntry{Source: source.Name, Path: name, Type: "file", Size: info.Size(), SHA256: hex.EncodeToString(hash.Sum(nil)), Mode: info.Mode()})
	return nil
}

func readArchive(ctx context.Context, archivePath string) (Manifest, []archiveFile, error) {
	reader, err := os.Open(archivePath)
	if err != nil {
		return Manifest{}, nil, err
	}
	defer reader.Close()
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return Manifest{}, nil, err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	var manifest Manifest
	var manifestFound bool
	files := make([]archiveFile, 0)
	for {
		select {
		case <-ctx.Done():
			return Manifest{}, nil, ctx.Err()
		default:
		}
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Manifest{}, nil, err
		}
		if _, err := safeRelative(header.Name); err != nil {
			return Manifest{}, nil, err
		}
		if header.Name == manifestPath {
			if err := json.NewDecoder(tarReader).Decode(&manifest); err != nil {
				return Manifest{}, nil, err
			}
			manifestFound = true
			continue
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		hash := sha256.New()
		size, err := io.Copy(hash, tarReader)
		if err != nil {
			return Manifest{}, nil, err
		}
		files = append(files, archiveFile{Path: header.Name, Size: size, SHA256: hex.EncodeToString(hash.Sum(nil))})
	}
	if !manifestFound {
		return Manifest{}, nil, errors.New("manifest missing")
	}
	return manifest, files, nil
}

func safeTarget(root, name string) (string, error) {
	rel, err := safeRelative(name)
	if err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	if target != rootAbs && !strings.HasPrefix(target, rootAbs+string(os.PathSeparator)) {
		return "", errors.New("unsafe restore path")
	}
	return target, nil
}

func safeRelative(name string) (string, error) {
	name = filepath.ToSlash(strings.TrimSpace(name))
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "../") || strings.HasPrefix(name, "..") {
		return "", errors.New("unsafe archive path")
	}
	return name, nil
}
