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
	if limit <= 0 || limit > 300 {
		limit = 300
	}
	local, domain, ok := strings.Cut(strings.ToLower(email), "@")
	if !ok || local == "" || domain == "" {
		return nil, errors.New("valid email is required")
	}
	root := filepath.Join(s.Root, domain, local, "Maildir")
	var paths []string
	for _, mailbox := range []string{"new", "cur"} {
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
	if strings.Contains(id, "/") || strings.Contains(id, `\`) || strings.Contains(id, "..") || id == "" {
		return MessageDetail{}, errors.New("invalid message id")
	}
	local, domain, ok := strings.Cut(strings.ToLower(email), "@")
	if !ok || local == "" || domain == "" {
		return MessageDetail{}, errors.New("valid email is required")
	}
	root := filepath.Join(s.Root, domain, local, "Maildir")
	for _, mailbox := range []string{"new", "cur"} {
		select {
		case <-ctx.Done():
			return MessageDetail{}, ctx.Err()
		default:
		}
		path := filepath.Join(root, mailbox, id)
		message, err := parseMessageDetail(path, root)
		if err == nil {
			return message, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return MessageDetail{}, err
		}
	}
	return MessageDetail{}, os.ErrNotExist
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
