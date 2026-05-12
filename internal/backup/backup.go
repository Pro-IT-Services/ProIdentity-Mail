package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
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
const encryptedArchiveMagic = "PROIDENTITY-BACKUP-AESGCM-CHUNKED-V1\n"
const encryptionChunkSize = 64 << 10

type Options struct {
	OutputPath    string
	Sources       []Source
	Hostname      string
	EncryptionKey []byte
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
	if len(options.EncryptionKey) > 0 {
		temp, err := os.CreateTemp(filepath.Dir(options.OutputPath), ".proidentity-backup-plain-*")
		if err != nil {
			return Manifest{}, err
		}
		tempPath := temp.Name()
		_ = temp.Close()
		defer os.Remove(tempPath)
		plainOptions := options
		plainOptions.OutputPath = tempPath
		plainOptions.EncryptionKey = nil
		manifest, err := Create(ctx, plainOptions)
		if err != nil {
			return Manifest{}, err
		}
		if err := EncryptFile(tempPath, options.OutputPath, options.EncryptionKey); err != nil {
			return Manifest{}, err
		}
		return manifest, nil
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
	return VerifyWithKey(ctx, archivePath, nil)
}

func VerifyWithKey(ctx context.Context, archivePath string, encryptionKey []byte) (VerifySummary, error) {
	manifest, files, err := readArchive(ctx, archivePath, encryptionKey)
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
	return RestoreWithKey(ctx, archivePath, targetRoot, options, nil)
}

func RestoreWithKey(ctx context.Context, archivePath, targetRoot string, options RestoreOptions, encryptionKey []byte) error {
	if _, err := VerifyWithKey(ctx, archivePath, encryptionKey); err != nil {
		return err
	}
	if strings.TrimSpace(targetRoot) == "" {
		return errors.New("target root is required")
	}
	reader, cleanup, err := openArchive(ctx, archivePath, encryptionKey)
	if err != nil {
		return err
	}
	defer reader.Close()
	defer cleanup()
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

func readArchive(ctx context.Context, archivePath string, encryptionKey []byte) (Manifest, []archiveFile, error) {
	reader, cleanup, err := openArchive(ctx, archivePath, encryptionKey)
	if err != nil {
		return Manifest{}, nil, err
	}
	defer reader.Close()
	defer cleanup()
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

func EncryptFile(inputPath, outputPath string, encryptionKey []byte) error {
	if len(encryptionKey) != 32 {
		return errors.New("backup encryption key must be 32 bytes")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return err
	}
	output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer output.Close()
	if _, err := output.Write([]byte(encryptedArchiveMagic)); err != nil {
		return err
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	buf := make([]byte, encryptionChunkSize)
	for {
		n, readErr := input.Read(buf)
		if n > 0 {
			nonce := make([]byte, aead.NonceSize())
			if _, err := rand.Read(nonce); err != nil {
				return err
			}
			ciphertext := aead.Seal(nil, nonce, buf[:n], nil)
			if uint64(len(ciphertext)) > uint64(^uint32(0)) {
				return errors.New("encrypted backup chunk too large")
			}
			if _, err := output.Write(nonce); err != nil {
				return err
			}
			var size [4]byte
			binary.BigEndian.PutUint32(size[:], uint32(len(ciphertext)))
			if _, err := output.Write(size[:]); err != nil {
				return err
			}
			if _, err := output.Write(ciphertext); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func openArchive(ctx context.Context, archivePath string, encryptionKey []byte) (*os.File, func(), error) {
	reader, err := os.Open(archivePath)
	if err != nil {
		return nil, func() {}, err
	}
	encrypted, err := archiveIsEncrypted(reader)
	if err != nil {
		_ = reader.Close()
		return nil, func() {}, err
	}
	if !encrypted {
		if _, err := reader.Seek(0, io.SeekStart); err != nil {
			_ = reader.Close()
			return nil, func() {}, err
		}
		return reader, func() {}, nil
	}
	_ = reader.Close()
	if len(encryptionKey) != 32 {
		return nil, func() {}, errors.New("encrypted backup requires a 32-byte encryption key")
	}
	temp, err := os.CreateTemp("", "proidentity-backup-decrypted-*")
	if err != nil {
		return nil, func() {}, err
	}
	tempPath := temp.Name()
	if err := decryptFile(ctx, archivePath, temp, encryptionKey); err != nil {
		_ = temp.Close()
		_ = os.Remove(tempPath)
		return nil, func() {}, err
	}
	if _, err := temp.Seek(0, io.SeekStart); err != nil {
		_ = temp.Close()
		_ = os.Remove(tempPath)
		return nil, func() {}, err
	}
	cleanup := func() {
		_ = os.Remove(tempPath)
	}
	return temp, cleanup, nil
}

func archiveIsEncrypted(reader *os.File) (bool, error) {
	magic := make([]byte, len(encryptedArchiveMagic))
	n, err := io.ReadFull(reader, magic)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return false, err
	}
	return n == len(magic) && string(magic) == encryptedArchiveMagic, nil
}

func decryptFile(ctx context.Context, encryptedPath string, output *os.File, encryptionKey []byte) error {
	input, err := os.Open(encryptedPath)
	if err != nil {
		return err
	}
	defer input.Close()
	magic := make([]byte, len(encryptedArchiveMagic))
	if _, err := io.ReadFull(input, magic); err != nil {
		return err
	}
	if string(magic) != encryptedArchiveMagic {
		return errors.New("backup archive is not encrypted with the expected format")
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		nonce := make([]byte, aead.NonceSize())
		n, err := io.ReadFull(input, nonce)
		if errors.Is(err, io.EOF) && n == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		var sizeBytes [4]byte
		if _, err := io.ReadFull(input, sizeBytes[:]); err != nil {
			return err
		}
		size := binary.BigEndian.Uint32(sizeBytes[:])
		maxChunkSize := uint32(encryptionChunkSize + aead.Overhead())
		if size == 0 || size > maxChunkSize {
			return errors.New("invalid encrypted backup chunk size")
		}
		ciphertext := make([]byte, size)
		if _, err := io.ReadFull(input, ciphertext); err != nil {
			return err
		}
		plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return errors.New("encrypted backup authentication failed")
		}
		if _, err := output.Write(plaintext); err != nil {
			return err
		}
	}
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
