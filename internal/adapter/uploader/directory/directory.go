package directory

import (
	"fmt"
	"io"
	"log"
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
}

func New(cfg Config, log *zap.Logger) (*Directory, error) {
	d := &Directory{
		Path: cfg.Path,
		Ext:  cfg.Extensions,
	}
	return d, nil
}

func (d *Directory) List(p string) ([]*models.File, error) {
	r := []*models.File{}
	entries, err := os.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}

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
			r = append(r, &models.File{
				Name:  e.Name(),
				IsDir: e.IsDir(),
				Size:  uint64(info.Size()),
				Time:  info.ModTime(),
				Path:  path.Join(p, e.Name()),
			})
		}
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
