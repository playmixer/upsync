package directory

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"upsync/internal/adapter/models"

	"go.uber.org/zap"
)

type Directory struct {
	Path string
	Ext  []string
	log  *zap.Logger
}

func New(cfg Config, log *zap.Logger) (*Directory, error) {
	d := &Directory{
		Path: cfg.Path,
		Ext:  cfg.Extensions,
		log:  log,
	}
	return d, nil
}

func (d *Directory) List(p string) ([]*models.File, error) {
	r := []*models.File{}
	entries, err := os.ReadDir(p)
	if err != nil {
		return r, fmt.Errorf("failed read dir: %w", err)
	}
	d.log.Debug("dir", zap.String("path", p), zap.Any("extensions", d.Ext), zap.Int("count", len(entries)))

	for _, e := range entries {
		if e.IsDir() {
			_r, err := d.List(path.Join(p, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed getting list `%s`: %w", e.Name(), err)
			}
			r = append(r, _r...)
			continue
		}
		fExtension := strings.ReplaceAll(filepath.Ext(e.Name()), ".", "")
		if len(d.Ext) == 0 || slices.Contains(d.Ext, fExtension) {
			info, _ := e.Info()
			fileInfo := &models.File{
				Name:  e.Name(),
				IsDir: e.IsDir(),
				Size:  uint64(info.Size()),
				Time:  info.ModTime(),
				Path:  path.Join(p, e.Name()),
			}
			r = append(r, fileInfo)
			d.log.Debug("dir file info", zap.String("name", e.Name()), zap.Any("info", fileInfo))
		}
		d.log.Debug("dir file", zap.String("name", e.Name()), zap.String("extension", fExtension), zap.String("path", path.Join(p, e.Name())))
	}

	return r, nil
}

func (d *Directory) Read(filepath string) (*[]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed open file: %w", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed read file: %w", err)
	}
	return &b, nil
}

func (d *Directory) Close() error {

	return nil
}
