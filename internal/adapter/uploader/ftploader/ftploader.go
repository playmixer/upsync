package ftploader

import (
	"fmt"
	"io"
	"path"
	"time"
	"upsync/internal/adapter/models"

	"github.com/jlaffaye/ftp"
	"go.uber.org/zap"
)

var (
	lazyCloseDuration int64 = 60 // second
)

type FTPLoader struct {
	cfg               Config
	ftp               *ftp.ServerConn
	log               *zap.Logger
	connected         bool
	connectedTime     int64
	lazyCloseDuration int64
}

func New(cfg Config, log *zap.Logger) (*FTPLoader, error) {
	f := &FTPLoader{
		cfg:               cfg,
		log:               log,
		lazyCloseDuration: lazyCloseDuration,
		connected:         false,
	}

	return f, nil
}

func (l *FTPLoader) List(p string) ([]*models.File, error) {
	r := []*models.File{}
	err := l.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed connect: %w", err)
	}
	// defer func() { _ = l.Close() }()

	list, err := l.ftp.List(p)
	if err != nil {
		return nil, fmt.Errorf("failed getting list: %w", err)
	}

	for _, f := range list {
		r = append(r, &models.File{
			Name:  f.Name,
			IsDir: f.Type == ftp.EntryTypeFolder,
			Size:  f.Size,
			Time:  f.Time,
			Path:  f.Name,
		})
	}

	return r, nil
}

func (l *FTPLoader) Connect() error {
	if l.connected {
		return nil
	}
	var err error
	address := fmt.Sprintf("%s:%v", l.cfg.Host, l.cfg.Port)
	l.ftp, err = ftp.Dial(address, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed connect to ftp: %w", err)
	}
	err = l.ftp.Login(l.cfg.Login, l.cfg.Password)
	if err != nil {
		return fmt.Errorf("failed login to ftp: %w", err)
	}
	l.connected = true
	l.connectedTime = time.Now().Unix()
	return nil
}

func (l *FTPLoader) Close() error {
	if !l.connected {
		return nil
	}
	if err := l.ftp.Quit(); err != nil {
		return fmt.Errorf("failed close connect: %w", err)
	}
	l.connected = false
	return nil
}

func (l *FTPLoader) Read(filepath string) (*[]byte, error) {
	fr, err := l.ftp.Retr(path.Join(l.cfg.Path, filepath))
	if err != nil {
		return nil, fmt.Errorf("failed getting file: %w", err)
	}
	defer fr.Close()

	data, err := io.ReadAll(fr)
	if err != nil {
		return nil, fmt.Errorf("failed reading file: %w", err)
	}

	return &data, nil
}
