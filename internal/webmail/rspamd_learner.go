package webmail

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

type RspamdLearner struct {
	Command string
}

func (l RspamdLearner) Learn(ctx context.Context, path, verdict string) error {
	command := l.Command
	if command == "" {
		command = "rspamc"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp("", "proidentity-rspamc-*.eml")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	args, err := rspamdLearnArgs(tempPath, verdict)
	if err != nil {
		return err
	}
	return exec.CommandContext(ctx, command, args...).Run()
}

func rspamdLearnArgs(path, verdict string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(verdict)) {
	case "spam":
		return []string{"learn_spam", path}, nil
	case "ham":
		return []string{"learn_ham", path}, nil
	default:
		return nil, errors.New("unsupported rspamd learning verdict")
	}
}
