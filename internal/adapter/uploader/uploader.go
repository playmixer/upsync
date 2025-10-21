package uploader

import (
	"fmt"
	"os"
	"upsync/internal/adapter/models"
	"upsync/internal/adapter/uploader/directory"
	"upsync/internal/adapter/uploader/ftploader"
	"upsync/internal/adapter/uploader/immich"

	"go.uber.org/zap"
)

type uploader interface {
	List(path string) ([]*models.File, error)
	Read(filepath string) (*[]byte, error)
	Close() error
}

type loader interface {
	ListStore() ([]*models.File, error)
	Write(f *os.File, name string) error
	Close() error
}

var (
	_ uploader = &ftploader.FTPLoader{}
	_ loader   = &immich.Immich{}
)

// сервисы предоставляющие данны
func NewRemote(typeLoader models.TProtocol, cfg Config, log *zap.Logger) (uploader, error) {
	log.Debug("loader", zap.Any("config", cfg))
	if typeLoader == models.RP_FTP {
		u, err := ftploader.New(ftploader.Config{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Login:    cfg.Login,
			Password: cfg.Password,
			Path:     cfg.Path,
		}, log)
		if err != nil {
			return nil, fmt.Errorf("failed initialize ftploader: %w", err)
		}
		return u, nil
	}
	if typeLoader == models.RP_DIR {
		d, err := directory.New(directory.Config{
			Path:       cfg.Path,
			Extensions: cfg.Extensions,
		}, log)
		if err != nil {
			return nil, fmt.Errorf("failed initialize direcory loader: %w", err)
		}
		return d, nil
	}

	return nil, fmt.Errorf("failed initialize loader `%s`", typeLoader)
}

// сервисы для загруки в них данных
func NewSyncStore(typeLoader models.TProtocol, cfg Config, log *zap.Logger) (loader, error) {
	log.Debug("loader", zap.Any("config", cfg))
	if typeLoader == models.RP_IMMICH {
		u, err := immich.New(immich.Config{
			Address: cfg.Address,
			APIKey:  cfg.APIKey,
			Path:    cfg.Path,
		}, log)
		if err != nil {
			return nil, fmt.Errorf("failed initialize ftploader: %w", err)
		}
		return u, nil
	}

	return nil, fmt.Errorf("failed initialize loader `%s`", typeLoader)
}
