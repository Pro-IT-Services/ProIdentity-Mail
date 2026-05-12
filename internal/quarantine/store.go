package quarantine

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStore struct {
	Root         string
	MailRoot     string
	DeliveryAddr string
}

type StoreRequest struct {
	TenantID  uint64
	Recipient string
	MessageID string
	Reader    io.Reader
}

type StoredMessage struct {
	StoragePath string
	SHA256      string
	SizeBytes   int64
}

type ReleaseRequest struct {
	Recipient    string
	MessageID    string
	StoragePath  string
	QuarantineID uint64
}

type ReleasedMessage struct {
	MailboxPath string
}

func (s FileStore) StoreMessage(ctx context.Context, req StoreRequest) (StoredMessage, error) {
	if s.Root == "" {
		return StoredMessage{}, errors.New("quarantine root is required")
	}
	if req.Reader == nil {
		return StoredMessage{}, errors.New("message reader is required")
	}
	if err := validateEmail(req.Recipient); err != nil {
		return StoredMessage{}, err
	}
	now := time.Now().UTC()
	name := safeName(req.MessageID)
	if name == "" {
		name = randomName()
	}
	rel := filepath.Join(fmt.Sprintf("tenant-%d", req.TenantID), now.Format("2006"), now.Format("01"), now.Format("02"), name+".eml")
	target, err := safeJoin(s.Root, rel)
	if err != nil {
		return StoredMessage{}, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0750); err != nil {
		return StoredMessage{}, err
	}
	temp := target + ".tmp-" + randomName()
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return StoredMessage{}, err
	}
	hash := sha256.New()
	written, copyErr := copyWithContext(ctx, io.MultiWriter(file, hash), req.Reader)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(temp)
		return StoredMessage{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(temp)
		return StoredMessage{}, closeErr
	}
	if err := os.Rename(temp, target); err != nil {
		_ = os.Remove(temp)
		return StoredMessage{}, err
	}
	return StoredMessage{StoragePath: filepath.ToSlash(rel), SHA256: hex.EncodeToString(hash.Sum(nil)), SizeBytes: written}, nil
}

func (s FileStore) Release(ctx context.Context, req ReleaseRequest) (ReleasedMessage, error) {
	if s.Root == "" || s.MailRoot == "" {
		if s.Root == "" || s.DeliveryAddr == "" {
			return ReleasedMessage{}, errors.New("quarantine root and release target are required")
		}
	}
	if err := validateEmail(req.Recipient); err != nil {
		return ReleasedMessage{}, err
	}
	source, err := safeJoin(s.Root, filepath.FromSlash(req.StoragePath))
	if err != nil {
		return ReleasedMessage{}, err
	}
	if _, err := os.Stat(source); err != nil {
		return ReleasedMessage{}, err
	}
	if s.DeliveryAddr != "" {
		data, err := os.ReadFile(source)
		if err != nil {
			return ReleasedMessage{}, err
		}
		_, domain, _ := strings.Cut(strings.ToLower(req.Recipient), "@")
		if err := sendLocalSMTP(s.DeliveryAddr, "postmaster@"+domain, strings.ToLower(req.Recipient), data); err != nil {
			return ReleasedMessage{}, err
		}
		return ReleasedMessage{}, nil
	}
	local, domain, _ := strings.Cut(strings.ToLower(req.Recipient), "@")
	inboxDir := filepath.Join(s.MailRoot, domain, local, "Maildir", "new")
	if err := os.MkdirAll(inboxDir, 0750); err != nil {
		return ReleasedMessage{}, err
	}
	name := safeName(req.MessageID)
	if name == "" {
		name = fmt.Sprintf("quarantine-%d-%s", req.QuarantineID, randomName())
	}
	target := filepath.Join(inboxDir, name+".eml")
	input, err := os.Open(source)
	if err != nil {
		return ReleasedMessage{}, err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return ReleasedMessage{}, err
	}
	if _, err := copyWithContext(ctx, output, input); err != nil {
		_ = output.Close()
		_ = os.Remove(target)
		return ReleasedMessage{}, err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(target)
		return ReleasedMessage{}, err
	}
	return ReleasedMessage{MailboxPath: target}, nil
}

func sendLocalSMTP(addr, from, recipient string, data []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()
	if err := client.Hello("localhost"); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (s FileStore) Delete(ctx context.Context, storagePath string) error {
	if s.Root == "" || strings.TrimSpace(storagePath) == "" {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := safeJoin(s.Root, filepath.FromSlash(storagePath))
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func copyWithContext(ctx context.Context, writer io.Writer, reader io.Reader) (int64, error) {
	buf := make([]byte, 64*1024)
	var written int64
	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}
		n, err := reader.Read(buf)
		if n > 0 {
			out, writeErr := writer.Write(buf[:n])
			written += int64(out)
			if writeErr != nil {
				return written, writeErr
			}
			if out != n {
				return written, io.ErrShortWrite
			}
		}
		if err == io.EOF {
			return written, nil
		}
		if err != nil {
			return written, err
		}
	}
}

func safeJoin(root, rel string) (string, error) {
	if filepath.IsAbs(rel) || strings.Contains(filepath.ToSlash(rel), "../") || strings.HasPrefix(filepath.ToSlash(rel), "..") {
		return "", errors.New("unsafe storage path")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(rootAbs, rel))
	if err != nil {
		return "", err
	}
	if targetAbs != rootAbs && !strings.HasPrefix(targetAbs, rootAbs+string(os.PathSeparator)) {
		return "", errors.New("unsafe storage path")
	}
	return targetAbs, nil
}

func validateEmail(email string) error {
	local, domain, ok := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	if !ok || local == "" || domain == "" || strings.Contains(local, "/") || strings.Contains(domain, "/") || strings.Contains(local, `\`) || strings.Contains(domain, `\`) {
		return errors.New("valid recipient is required")
	}
	return nil
}

func safeName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "<>")
	value = strings.ReplaceAll(value, "@", "_")
	value = strings.ReplaceAll(value, ".", "_")
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if len(result) > 120 {
		return result[:120]
	}
	return result
}

func randomName() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
