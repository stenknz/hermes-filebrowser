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

func NewService(root string) (*Service, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Service{root: absRoot}, nil
}

var ErrIsRoot = os.ErrInvalid // operating on the root directory is not allowed

func (s *Service) SafePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", os.ErrPermission
	}
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
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return absPath, nil
		}
		return "", err
	}
	if !strings.HasPrefix(resolvedPath, s.root) {
		return "", os.ErrPermission
	}
	return resolvedPath, nil
}

var hiddenPrefixes = []string{"filebrowser.db", "filebrowser.db-"}
var hiddenSuffixes = []string{".db-shm", ".db-wal"}

func isHidden(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	for _, prefix := range hiddenPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	for _, suffix := range hiddenSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
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
		if isHidden(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
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
	if isHidden(filepath.Base(filePath)) {
		return nil, os.ErrPermission
	}
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fullPath)
}

func (s *Service) Write(filePath string, data []byte) error {
	if filePath == "" {
		return ErrIsRoot
	}
	if isHidden(filepath.Base(filePath)) {
		return os.ErrPermission
	}
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *Service) Delete(filePath string) error {
	if filePath == "" {
		return ErrIsRoot
	}
	if isHidden(filepath.Base(filePath)) {
		return os.ErrPermission
	}
	fullPath, err := s.SafePath(filePath)
	if err != nil {
		return err
	}
	return os.RemoveAll(fullPath)
}

func (s *Service) Rename(oldPath, newPath string) error {
	if oldPath == "" || newPath == "" {
		return ErrIsRoot
	}
	if isHidden(filepath.Base(oldPath)) || isHidden(filepath.Base(newPath)) {
		return os.ErrPermission
	}
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
	if src == "" || dst == "" {
		return ErrIsRoot
	}
	if isHidden(filepath.Base(src)) || isHidden(filepath.Base(dst)) {
		return os.ErrPermission
	}
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
	_, err = io.Copy(dstFile, srcFile)
	if closeErr := dstFile.Close(); closeErr != nil {
		return closeErr
	}
	return err
}

func (s *Service) Mkdir(dirPath string) error {
	if dirPath == "" {
		return ErrIsRoot
	}
	fullPath, err := s.SafePath(dirPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullPath, 0755)
}

func (s *Service) Thumbnail(filePath string) ([]byte, error) {
	if isHidden(filepath.Base(filePath)) {
		return nil, os.ErrPermission
	}
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
