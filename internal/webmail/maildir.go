package webmail

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type MessageSummary struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
	Preview   string    `json:"preview"`
	Mailbox   string    `json:"mailbox"`
	SizeBytes int64     `json:"size_bytes"`
}

type MessageDetail struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Date      time.Time `json:"date"`
	Body      string    `json:"body"`
	Mailbox   string    `json:"mailbox"`
	SizeBytes int64     `json:"size_bytes"`
}

type MaildirStore struct {
	Root string
}

func (s MaildirStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	return s.ListMessages(ctx, email, "inbox", limit)
}

func (s MaildirStore) ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error) {
	if limit <= 0 || limit > 300 {
		limit = 300
	}
	local, domain, ok := strings.Cut(strings.ToLower(email), "@")
	if !ok || local == "" || domain == "" {
		return nil, errors.New("valid email is required")
	}
	root := filepath.Join(s.Root, domain, local, "Maildir")
	var paths []string
	for _, mailbox := range folderMailboxes(folder) {
		dir := filepath.Join(root, mailbox)
		entries, err := os.ReadDir(dir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Slice(paths, func(i, j int) bool {
		left, _ := os.Stat(paths[i])
		right, _ := os.Stat(paths[j])
		if left == nil || right == nil {
			return paths[i] > paths[j]
		}
		return left.ModTime().After(right.ModTime())
	})
	if len(paths) > limit {
		paths = paths[:limit]
	}
	messages := make([]MessageSummary, 0, len(paths))
	for _, path := range paths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		message, err := parseMessageSummary(path, root)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func folderMailboxes(folder string) []string {
	switch strings.ToLower(strings.TrimSpace(folder)) {
	case "spam":
		return []string{filepath.Join(".Spam", "new"), filepath.Join(".Spam", "cur")}
	case "trash":
		return []string{filepath.Join(".Trash", "new"), filepath.Join(".Trash", "cur")}
	case "archive":
		return []string{filepath.Join(".Archive", "new"), filepath.Join(".Archive", "cur")}
	case "all":
		return []string{"new", "cur", filepath.Join(".Spam", "new"), filepath.Join(".Spam", "cur"), filepath.Join(".Trash", "new"), filepath.Join(".Trash", "cur"), filepath.Join(".Archive", "new"), filepath.Join(".Archive", "cur")}
	default:
		return []string{"new", "cur"}
	}
}

func parseMessageSummary(path, maildirRoot string) (MessageSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MessageSummary{}, err
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return MessageSummary{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return MessageSummary{}, err
	}
	date, _ := msg.Header.Date()
	rel, _ := filepath.Rel(maildirRoot, path)
	mailbox := strings.Split(filepath.ToSlash(rel), "/")[0]
	return MessageSummary{
		ID:        filepath.Base(path),
		From:      msg.Header.Get("From"),
		To:        msg.Header.Get("To"),
		Subject:   msg.Header.Get("Subject"),
		Date:      date,
		Preview:   preview(msg.Body),
		Mailbox:   mailbox,
		SizeBytes: info.Size(),
	}, nil
}

func (s MaildirStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return MessageDetail{}, err
	}
	path, err := s.messagePath(ctx, root, id)
	if err != nil {
		return MessageDetail{}, err
	}
	return parseMessageDetail(path, root)
}

func (s MaildirStore) MoveMessage(ctx context.Context, email, id, folder string) error {
	root, err := s.maildirRoot(email)
	if err != nil {
		return err
	}
	source, err := s.messagePath(ctx, root, id)
	if err != nil {
		return err
	}
	destinationMailbox, err := destinationMailbox(folder)
	if err != nil {
		return err
	}
	destinationDir := filepath.Join(root, destinationMailbox, "new")
	if destinationMailbox == "new" {
		destinationDir = filepath.Join(root, "new")
	}
	if err := os.MkdirAll(destinationDir, 0750); err != nil {
		return err
	}
	return os.Rename(source, filepath.Join(destinationDir, filepath.Base(source)))
}

func (s MaildirStore) MessagePath(ctx context.Context, email, id string) (string, error) {
	root, err := s.maildirRoot(email)
	if err != nil {
		return "", err
	}
	return s.messagePath(ctx, root, id)
}

func (s MaildirStore) maildirRoot(email string) (string, error) {
	local, domain, ok := strings.Cut(strings.ToLower(email), "@")
	if !ok || local == "" || domain == "" {
		return "", errors.New("valid email is required")
	}
	return filepath.Join(s.Root, domain, local, "Maildir"), nil
}

func (s MaildirStore) messagePath(ctx context.Context, root, id string) (string, error) {
	if strings.Contains(id, "/") || strings.Contains(id, `\`) || strings.Contains(id, "..") || id == "" {
		return "", errors.New("invalid message id")
	}
	for _, mailbox := range []string{"new", "cur", ".Spam/new", ".Spam/cur", ".Trash/new", ".Trash/cur", ".Archive/new", ".Archive/cur"} {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		path := filepath.Join(root, filepath.FromSlash(mailbox), id)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	return "", os.ErrNotExist
}

func destinationMailbox(folder string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(folder)) {
	case "inbox", "":
		return "new", nil
	case "spam":
		return ".Spam", nil
	case "trash":
		return ".Trash", nil
	case "archive":
		return ".Archive", nil
	default:
		return "", errors.New("unsupported mailbox folder")
	}
}

func parseMessageDetail(path, maildirRoot string) (MessageDetail, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MessageDetail{}, err
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return MessageDetail{}, err
	}
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return MessageDetail{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return MessageDetail{}, err
	}
	date, _ := msg.Header.Date()
	rel, _ := filepath.Rel(maildirRoot, path)
	mailbox := strings.Split(filepath.ToSlash(rel), "/")[0]
	return MessageDetail{
		ID:        filepath.Base(path),
		From:      msg.Header.Get("From"),
		To:        msg.Header.Get("To"),
		Subject:   msg.Header.Get("Subject"),
		Date:      date,
		Body:      string(body),
		Mailbox:   mailbox,
		SizeBytes: info.Size(),
	}, nil
}

func preview(reader io.Reader) string {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 4096), 64*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			if len(line) > 160 {
				return line[:160]
			}
			return line
		}
	}
	return ""
}
