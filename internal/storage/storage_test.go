package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cs2-parser/internal/models"
)

func newTestStorage(t *testing.T) (*Storage, string) {
	t.Helper()
	root := t.TempDir()
	s, err := New(root, "1")
	if err != nil {
		t.Fatal(err)
	}
	return s, root
}

func TestValidHash(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid", strings.Repeat("a", 64), true},
		{"valid mixed", strings.Repeat("0123456789abcdef", 4), true},
		{"empty", "", false},
		{"too short", strings.Repeat("a", 63), false},
		{"too long", strings.Repeat("a", 65), false},
		{"uppercase", strings.Repeat("A", 64), false},
		{"non hex", strings.Repeat("g", 64), false},
		{"path traversal", "../" + strings.Repeat("a", 61), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidHash(tc.in); got != tc.want {
				t.Errorf("ValidHash(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestNewCreatesDirs(t *testing.T) {
	_, root := newTestStorage(t)

	for _, dir := range []string{"demos", filepath.Join("results", "v1"), "tmp"} {
		info, err := os.Stat(filepath.Join(root, dir))
		if err != nil {
			t.Fatalf("stat %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestSaveDemoHashesContent(t *testing.T) {
	s, _ := newTestStorage(t)
	content := []byte("not really a demo")

	hash, err := s.SaveDemo(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(content)
	if want := hex.EncodeToString(sum[:]); hash != want {
		t.Errorf("hash = %s, want %s", hash, want)
	}

	got, err := os.ReadFile(s.DemoPath(hash))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("stored content = %q, want %q", got, content)
	}
}

func TestSaveDemoTwiceKeepsOneFile(t *testing.T) {
	s, root := newTestStorage(t)
	content := []byte("same demo")

	first, err := s.SaveDemo(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.SaveDemo(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Errorf("same content gave different hashes: %s vs %s", first, second)
	}

	demos, _ := os.ReadDir(filepath.Join(root, "demos"))
	if len(demos) != 1 {
		t.Errorf("demos dir has %d files, want 1", len(demos))
	}
	leftovers, _ := os.ReadDir(filepath.Join(root, "tmp"))
	if len(leftovers) != 0 {
		t.Errorf("tmp dir has %d leftover files, want 0", len(leftovers))
	}
}

func TestResultRoundTrip(t *testing.T) {
	s, _ := newTestStorage(t)
	hash := strings.Repeat("ab", 32)

	if _, err := s.LoadResult(hash); !errors.Is(err, ErrNotFound) {
		t.Fatalf("LoadResult before save: err = %v, want ErrNotFound", err)
	}

	want := &models.MatchStats{
		MapName: "de_mirage",
		CT:      models.TeamStats{Name: "Team A", Score: 13},
		Players: map[uint64]*models.PlayerStats{
			76561198000000001: {SteamID: 76561198000000001, Name: "player", Kills: 20},
		},
	}
	if err := s.SaveResult(hash, want); err != nil {
		t.Fatal(err)
	}

	got, err := s.LoadResult(hash)
	if err != nil {
		t.Fatal(err)
	}
	if got.MapName != want.MapName || got.CT != want.CT {
		t.Errorf("got %+v, want %+v", got, want)
	}
	p := got.Players[76561198000000001]
	if p == nil || p.Kills != 20 {
		t.Errorf("player not restored correctly: %+v", p)
	}
}

func TestResultsAreVersioned(t *testing.T) {
	root := t.TempDir()
	hash := strings.Repeat("cd", 32)

	v1, err := New(root, "1")
	if err != nil {
		t.Fatal(err)
	}
	if err := v1.SaveResult(hash, &models.MatchStats{MapName: "de_dust2"}); err != nil {
		t.Fatal(err)
	}

	v2, err := New(root, "2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v2.LoadResult(hash); !errors.Is(err, ErrNotFound) {
		t.Errorf("v2 sees v1 result: err = %v, want ErrNotFound", err)
	}
}

func TestResultRejectsInvalidHash(t *testing.T) {
	s, _ := newTestStorage(t)
	bad := "../../etc/passwd"

	if _, err := s.LoadResult(bad); !errors.Is(err, ErrInvalidHash) {
		t.Errorf("LoadResult: err = %v, want ErrInvalidHash", err)
	}
	if err := s.SaveResult(bad, &models.MatchStats{}); !errors.Is(err, ErrInvalidHash) {
		t.Errorf("SaveResult: err = %v, want ErrInvalidHash", err)
	}
}
