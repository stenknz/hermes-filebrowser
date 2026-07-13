### Task 4: File Service

**Files:**
- Create: `backend/internal/fs/service.go`

**Interfaces:**
- Consumes: `config.Config.Root`
- Produces: `NewService(root string) *Service` with methods: `List(path string) ([]FileInfo, error)`, `Read(path string) ([]byte, error)`, `Write(path string, data []byte) error`, `Delete(path string) error`, `Rename(oldPath, newPath string) error`, `Copy(src, dst string) error`, `Mkdir(path string) error`, `Thumbnail(path string) ([]byte, error)`, `SafePath(path string) (string, error)`

- [ ] **Step 1: Create file service with path traversal prevention**

File: `backend/internal/fs/service.go`
```go
package fs

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

type FileInfo struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	Size    int64       `json:"size"`
	IsDir   bool        `json:"isDir"`
	ModTime time.Time   `json:"modTime"`
	Mode    os.FileMode `json:"mode"`
}

type Service struct {
	root string
}

func NewService(root string) *Service {
	absRoot, _ := filepath.Abs(root)
	return &Service{root: absRoot}
}

func (s *Service) SafePath(path string) (string, error) {
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "..") || strings.Contains(clean, "..") {
		return "", os.ErrPermission
	}
	fullPath := filepath.Join(s.root, clean)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, s.root) {
		return "", os.ErrPermission
	}
	return absPath, nil
}

func (s *Service) List(dirPath string) ([]FileInfo, error) {
	fullPath, err := s.SafePath(dirPath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	var infos []FileInfo
	for _, e := range entries {
		info, _ := e.Info()
		infos = append(infos, FileInfo{
			Name:    e.Name(),
			Path:    filepath.Join(dirPath, e.Name()),
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
	}
	return infos, nil
}

func (s *Service) Read(filePath string) ([]byte, error) {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fullPath)
}

func (s *Service) Write(filePath string, data []byte) error {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *Service) Delete(filePath string) error {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullPath)
}

func (s *Service) Rename(oldPath, newPath string) error {
	fullOld, err := s.SafePath(oldPath)
	if err != nil {
		return err
	}
	fullNew, err := s.SafePath(newPath)
	if err != nil {
		return err
	}
	return os.Rename(fullOld, fullNew)
}

func (s *Service) Copy(src, dst string) error {
	fullSrc, err := s.SafePath(src)
	if err != nil {
		return err
	}
	fullDst, err := s.SafePath(dst)
	if err != nil {
		return err
	}
	srcFile, err := os.Open(fullSrc)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(fullDst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *Service) Mkdir(dirPath string) error {
	fullPath, err := s.SafePath(dirPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullPath, 0755)
}

func (s *Service) Thumbnail(filePath string) ([]byte, error) {
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return nil, err
	}
	src, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	const maxSize = 200
	newW, newH := bounds.Dx(), bounds.Dy()
	if newW > maxSize || newH > maxSize {
		ratio := float64(maxSize) / float64(max(newW, newH))
		newW = int(float64(newW) * ratio)
		newH = int(float64(newH) * ratio)
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 2: Add thumbnail dependency**

```bash
cd backend && go get golang.org/x/image && go mod tidy && go build ./cmd/hermes
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/fs/
git commit -m "feat: add file service with path traversal prevention"
```
