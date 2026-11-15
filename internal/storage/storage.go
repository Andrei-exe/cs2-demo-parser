// Package storage keeps uploaded demos and their parse results on disk,
// addressed by the SHA-256 of the demo content.
//
// Layout under root:
//
//	demos/<hash>.dem              the uploaded demo
//	results/v<version>/<hash>.json the cached MatchStats for that demo
//	tmp/                          files being written, renamed into place when complete
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cs2-parser/internal/models"
)

// ErrNotFound is returned by LoadResult when there is no cached result for a hash.
var ErrNotFound = errors.New("result not found")

// ErrInvalidHash is returned when a hash is not 64 lowercase hex characters.
var ErrInvalidHash = errors.New("invalid hash")

type Storage struct {
	demosDir   string
	resultsDir string
	tmpDir     string
}

// New creates the directory layout under root. resultsVersion separates
// results produced by different parser versions, so bumping it invalidates
// the whole cache without deleting anything.
func New(root, resultsVersion string) (*Storage, error) {
	s := &Storage{
		demosDir:   filepath.Join(root, "demos"),
		resultsDir: filepath.Join(root, "results", "v"+resultsVersion),
		tmpDir:     filepath.Join(root, "tmp"),
	}

	for _, dir := range []string{s.demosDir, s.resultsDir, s.tmpDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", dir, err)
		}
	}

	return s, nil
}

// DemoPath returns where the demo with the given hash is stored.
func (s *Storage) DemoPath(hash string) string {
	return filepath.Join(s.demosDir, hash+".dem")
}

// ResultPath returns where the cached result for the given hash is stored.
func (s *Storage) ResultPath(hash string) string {
	return filepath.Join(s.resultsDir, hash+".json")
}

// ClearTmp removes leftovers from writes interrupted by a crash or kill.
// Only call it at startup, while nothing else is writing.
func (s *Storage) ClearTmp() error {
	if err := os.RemoveAll(s.tmpDir); err != nil {
		return fmt.Errorf("clear tmp: %w", err)
	}
	return os.MkdirAll(s.tmpDir, 0o755)
}

// SaveDemo copies r to disk and hashes it in the same pass, then moves the
// file to DemoPath(hash). If that demo is already stored, the new copy is
// dropped. Returns the lowercase hex SHA-256 of the content.
func (s *Storage) SaveDemo(r io.Reader) (string, error) {
	tmp, err := os.CreateTemp(s.tmpDir, "upload-*.dem")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	// After a successful rename the temp path no longer exists and this is a no-op
	defer func() { _ = os.Remove(tmp.Name()) }()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), r); err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("write demo: %w", err)
	}
	// Close before rename: Windows refuses to move a file that is still open
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close demo: %w", err)
	}

	hash := hex.EncodeToString(h.Sum(nil))
	dst := s.DemoPath(hash)

	// Same content already on disk: keep the existing file. Renaming over it
	// would fail on Windows if a worker is reading it right now
	if exists(dst) {
		return hash, nil
	}

	if err := os.Rename(tmp.Name(), dst); err != nil {
		// A concurrent upload of the same demo may have won the race
		if exists(dst) {
			return hash, nil
		}
		return "", fmt.Errorf("store demo: %w", err)
	}

	return hash, nil
}

// LoadResult reads the cached result for hash. Returns ErrNotFound if the
// demo was never parsed by this parser version.
func (s *Storage) LoadResult(hash string) (*models.MatchStats, error) {
	if !ValidHash(hash) {
		return nil, ErrInvalidHash
	}

	data, err := os.ReadFile(s.ResultPath(hash))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read result: %w", err)
	}

	var res models.MatchStats
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("decode result: %w", err)
	}

	return &res, nil
}

// SaveResult writes res as the cached result for hash. The file is written
// to tmp first and renamed, so readers never see a half-written result.
func (s *Storage) SaveResult(hash string, res *models.MatchStats) error {
	if !ValidHash(hash) {
		return ErrInvalidHash
	}

	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}

	tmp, err := os.CreateTemp(s.tmpDir, "result-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write result: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close result: %w", err)
	}

	if err := os.Rename(tmp.Name(), s.ResultPath(hash)); err != nil {
		return fmt.Errorf("store result: %w", err)
	}

	return nil
}

// ValidHash reports whether h looks like a hash produced by SaveDemo:
// exactly 64 lowercase hex characters. Any hash coming from outside
// (URL, request body) must pass this before it is used in a path.
func ValidHash(h string) bool {
	if len(h) != sha256.Size*2 {
		return false
	}
	for i := 0; i < len(h); i++ {
		c := h[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
