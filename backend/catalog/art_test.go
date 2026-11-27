package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoverFor(t *testing.T) {
	url, isArt := CoverFor("Neon Drift 2089", "neon-drift-2089-steam-key-global")
	if !isArt || url != "/covers/neon-drift-2089.webp?v="+artVersion {
		t.Errorf("a title with art: got %q, %v", url, isArt)
	}

	// A title nobody drew falls back to the SVG the API draws itself.
	url, isArt = CoverFor("Some New Game", "some-new-game-steam-key-global")
	if isArt || url != CoverURL("some-new-game-steam-key-global") {
		t.Errorf("a title without art: got %q, %v", url, isArt)
	}

	// A title is never trusted as a path. Even an odd one can only ever
	// select a name from the fixed list, or the fallback.
	url, isArt = CoverFor("../../etc/passwd", "x")
	if isArt || strings.Contains(url, "..") {
		t.Errorf("an odd title: got %q, %v", url, isArt)
	}
}

// The list in art.go and the files in frontend/public/covers must agree.
// A name on the list without a file would be a broken picture on the site.
// A file missing from the list would sit in the repository unused.
func TestArtTitlesMatchTheFiles(t *testing.T) {
	// Tests run with the package's own folder as the working directory.
	folder := filepath.Join("..", "..", "frontend", "public", "covers")
	entries, err := os.ReadDir(folder)
	if err != nil {
		t.Fatalf("reading %s: %v", folder, err)
	}

	onDisk := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".webp") {
			t.Errorf("%s: only .webp files belong in the covers folder", name)
			continue
		}
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		// Covers are 450 x 600. Anything much larger than 60 KB was probably
		// saved at the wrong size or quality.
		if info.Size() == 0 || info.Size() > 80*1024 {
			t.Errorf("%s is %d bytes, want more than 0 and at most 80 KB", name, info.Size())
		}
		onDisk[strings.TrimSuffix(name, ".webp")] = true
	}

	for title := range artTitles {
		if !onDisk[title] {
			t.Errorf("%s is on the list in art.go but has no file in %s", title, folder)
		}
	}
	for title := range onDisk {
		if !artTitles[title] {
			t.Errorf("%s.webp is in %s but not on the list in art.go", title, folder)
		}
	}
}
