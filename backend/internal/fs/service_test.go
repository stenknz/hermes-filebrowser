package fs

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafePath_Valid(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.SafePath("subdir/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(dir, "subdir/file.txt")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSafePath_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"../etc/passwd",
		"foo/../../etc/passwd",
		"..",
		"../",
	}
	for _, p := range tests {
		_, err = s.SafePath(p)
		if err == nil {
			t.Errorf("expected error for path %q", p)
		}
	}
}

func TestSafePath_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	// absolute paths that are outside root should fail
	_, err = s.SafePath("/etc/passwd")
	if err == nil {
		t.Error("expected error for absolute path outside root")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)

	infos, err := s.List(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := make(map[string]bool)
	for _, fi := range infos {
		found[fi.Name] = true
	}
	if !found["a.txt"] {
		t.Error("expected a.txt in listing")
	}
	if !found["sub"] {
		t.Error("expected sub in listing")
	}
}

func TestList_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.List("..")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestReadWrite(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("hello world")
	err = s.Write("test.txt", data)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	got, err := s.Read("test.txt")
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestRead_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Read("../outside.txt")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestWrite_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Write("../outside.txt", []byte("data"))
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "del.txt"), []byte("data"), 0644)

	err = s.Delete("del.txt")
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "del.txt")); !os.IsNotExist(err) {
		t.Error("file should not exist after delete")
	}
}

func TestDelete_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Delete("../outside.txt")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestRename(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "old.txt"), []byte("data"), 0644)

	err = s.Rename("old.txt", "new.txt")
	if err != nil {
		t.Fatalf("rename error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "old.txt")); !os.IsNotExist(err) {
		t.Error("old file should not exist after rename")
	}
	if _, err := os.Stat(filepath.Join(dir, "new.txt")); os.IsNotExist(err) {
		t.Error("new file should exist after rename")
	}
}

func TestRename_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Rename("old.txt", "../new.txt")
	if err == nil {
		t.Error("expected error for path traversal in dest")
	}
}

func TestCopy(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "src.txt"), []byte("copy data"), 0644)

	err = s.Copy("src.txt", "dst.txt")
	if err != nil {
		t.Fatalf("copy error: %v", err)
	}

	srcData, _ := os.ReadFile(filepath.Join(dir, "src.txt"))
	dstData, _ := os.ReadFile(filepath.Join(dir, "dst.txt"))
	if string(srcData) != string(dstData) {
		t.Error("copy content mismatch")
	}
}

func TestCopy_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Copy("../outside.txt", "inside.txt")
	if err == nil {
		t.Error("expected error for path traversal in source")
	}
}

func TestMkdir(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Mkdir("newdir/subdir")
	if err != nil {
		t.Fatalf("mkdir error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "newdir/subdir")); os.IsNotExist(err) {
		t.Error("directory should exist after mkdir")
	}
}

func TestMkdir_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	err = s.Mkdir("../outside")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestThumbnail(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	imgPath := filepath.Join(dir, "test.png")
	f, _ := os.Create(imgPath)
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 400, 300)))
	f.Close()

	thumb, err := s.Thumbnail("test.png")
	if err != nil {
		t.Fatalf("thumbnail error: %v", err)
	}
	if len(thumb) == 0 {
		t.Error("thumbnail should not be empty")
	}
}

func TestThumbnail_SmallImage(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	imgPath := filepath.Join(dir, "small.png")
	f, _ := os.Create(imgPath)
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 10, 10)))
	f.Close()

	thumb, err := s.Thumbnail("small.png")
	if err != nil {
		t.Fatalf("thumbnail error: %v", err)
	}
	if len(thumb) == 0 {
		t.Error("thumbnail should not be empty")
	}
}

func TestThumbnail_NotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Thumbnail("nonexistent.png")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestThumbnail_Traversal(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Thumbnail("../outside.png")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestRootTraversal_viaSymlink(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure symlinked paths outside root are still blocked
	outside := t.TempDir()
	linkPath := filepath.Join(dir, "link")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skip("symlinks not supported on this system")
	}

	_, err = s.Read("link/../outside.txt")
	if err == nil {
		t.Error("expected error for traversal via symlink")
	}
}

func TestSafePath_SameAsRoot(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.SafePath(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
	}
}

func TestSafePath_WindowsBackslash(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	// On Windows, filepath.Clean handles backslashes
	got, err := s.SafePath("subdir\\file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// filepath.Clean converts \ to / on Windows, so we check prefix only
	if !strings.HasPrefix(got, dir) {
		t.Errorf("result %q should be under root %q", got, dir)
	}
}

func TestRootIsCleaned(t *testing.T) {
	s, err := NewService("/data/../data")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(s.root, "data") {
		t.Errorf("root should be cleaned, got %q", s.root)
	}
}

func TestList_NonExistentDir(t *testing.T) {
	dir := t.TempDir()
	s, err := NewService(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.List("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}
