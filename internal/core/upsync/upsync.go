package upsync

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"upsync/internal/adapter/models"
	"upsync/internal/adapter/uploader"

	"go.uber.org/zap"
)

type Store interface {
	GetSynces() ([]*models.SyncItem, error)
}

type UpSync struct {
	store   Store
	log     *zap.Logger
	wC      chan func() error
	wg      *sync.WaitGroup
	tempDir string
	done    chan struct{}
	cfg     Config
}

func New(ctx context.Context, cfg Config, store Store, log *zap.Logger) (*UpSync, error) {
	if cfg.WorkerPoolCount < 1 {
		cfg.WorkerPoolCount = 1
	}

	u := &UpSync{
		store:   store,
		log:     log,
		wC:      make(chan func() error, 10),
		wg:      &sync.WaitGroup{},
		tempDir: "./temp",
		done:    make(chan struct{}),
		cfg:     cfg,
	}

	for i := range cfg.WorkerPoolCount {
		go u.worker(ctx, i)
	}

	_ = os.Mkdir(u.tempDir, os.ModePerm)

	return u, nil
}

func (u *UpSync) Sync(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	synces, err := u.store.GetSynces()
	if err != nil {
		return fmt.Errorf("failed getting remotes: %w", err)
	}

	u.log.Info("sync started", zap.Int("tasks_count", len(synces)))
	var successCount, errorCount int
	for _, item := range synces {
		select {
		case <-ctx.Done():
			u.log.Info("stopped range remotes from context")
			return nil
		default:
			u.log.Info("starting sync", zap.String("name", item.Title))
			err = u.syncDo(ctx, item)
			if err != nil {
				u.log.Error("failed sync", zap.String("name", item.Title), zap.Error(err))
				errorCount++
			} else {
				successCount++
			}
			u.log.Info("stop sync", zap.String("name", item.Title))
		}
	}

	u.log.Info("sync finished",
		zap.Int("total", len(synces)),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

func (u *UpSync) syncDo(ctx context.Context, item *models.SyncItem) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	upl, err := uploader.NewRemote(
		item.Remote.Protocol,
		uploader.Config{
			Host:       item.Remote.Host,
			Port:       item.Remote.Port,
			Login:      item.Remote.Login,
			Password:   item.Remote.Password,
			Path:       item.Remote.Path,
			Extensions: item.Remote.Extensions.Slice(),
		},
		u.log,
	)
	if err != nil {
		return fmt.Errorf("failed initialize remote store: %w", err)
	}
	defer upl.Close()

	loa, err := uploader.NewSyncStore(
		item.Store.Protocol,
		uploader.Config{
			Address: item.Store.Address,
			APIKey:  item.Store.APIKey,
			Path:    item.Store.Path,
		},
		u.log,
	)
	if err != nil {
		return fmt.Errorf("failed initialize sync store: %w", err)
	}
	// получаем список фйлов из хранилища
	storeList, err := loa.ListStore()
	if err != nil {
		return fmt.Errorf("failed getting store list: %w", err)
	}
	u.log.Info("store count files", zap.String("name", item.Title), zap.Int("count", len(storeList)))
	mStoreList := make(map[string]*models.File)
	for _, f := range storeList {
		mStoreList[f.Name] = f
	}
	u.log.Debug("store list map", zap.Any("mStoreList_keys", func() []string {
		keys := make([]string, 0, len(mStoreList))
		for k := range mStoreList {
			keys = append(keys, k)
		}
		return keys
	}()))

	remoteList, err := upl.List(item.Remote.Path)
	if err != nil {
		return fmt.Errorf("failed getting list: %w", err)
	}
	u.log.Info("remote count files", zap.String("name", item.Title), zap.Int("count", len(remoteList)))

	genFilePath := func(ctx context.Context) <-chan string {
		c := make(chan string, 100)
		go func(ctx context.Context) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			defer close(c)
			for _, remoteFile := range remoteList {
				select {
				case <-ctx.Done():
					u.log.Info("stopped range list from context")
					return
				default:
					if !remoteFile.IsDir {
						if _, ok := mStoreList[remoteFile.Name]; ok {
							u.log.Debug("file exists", zap.String("file", remoteFile.Name))
							continue
						}
						c <- remoteFile.Path
					}
				}
			}
		}(ctx)

		return c
	}(ctx)

loopGenFilePath:
	for {
		select {
		case <-ctx.Done():
			return nil
		case p := <-genFilePath:
			if p == "" {
				break loopGenFilePath
			}
			// подготваливаем файл для последующего сохранения
			bFile, err := upl.Read(p)
			if err != nil {
				return fmt.Errorf("failed read file: %w", err)
			}

			filename := filepath.Base(p)
			f, err := os.Create(path.Join(u.tempDir, filename))
			if err != nil {
				return fmt.Errorf("failed create file: %w", err)
			}
			_, err = f.Write(*bFile)
			if err != nil {
				f.Close()
				return fmt.Errorf("failed write file: %w", err)
			}
			u.log.Debug("upload temp file complete", zap.String("name", filename))

			type fR func() error
			u.wC <- func(f *os.File, fname string) fR {
				_f := f
				_fname := fname
				return func() error {
					defer func() {
						err := os.Remove(_f.Name())
						if err != nil {
							u.log.Error("failed remove file", zap.String("name", _f.Name()), zap.Error(err))
						}
					}()
					defer _f.Close()
					err = loa.Write(_f, _fname)
					if err != nil {
						return fmt.Errorf("failed write file: %w", err)
					}
					u.log.Info("upload file complete", zap.String("file", _fname))

					return nil
				}
			}(f, filename)
		}
	}

	return nil
}

func (u *UpSync) worker(ctx context.Context, idx int) {
	u.wg.Add(1)
	defer u.wg.Done()
	u.log.Info("start worker", zap.Int("id", idx))

workerLoop:
	for {
		select {
		case <-u.done:
			u.log.Info("stopping worker* ...", zap.Int("id", idx))
			break workerLoop
		case <-ctx.Done():
			u.log.Info("stopping worker ...", zap.Int("id", idx))
			break workerLoop
		case f := <-u.wC:
			if f != nil {
				if err := f(); err != nil {
					u.log.Error("failed complete work", zap.Error(err))
				}
			}
			u.log.Debug("work func complete", zap.Int("id", idx))
		}
	}
	u.log.Info("stopped worker", zap.Int("id", idx))
}

func (u *UpSync) Close() {
	u.log.Info("close upsync ...")
	close(u.done)
	u.log.Info("wait stopping all workers ...")
	u.wg.Wait()
	close(u.wC)
	u.log.Info("cleanup temp directory ...")
	if err := os.RemoveAll(u.tempDir); err != nil {
		u.log.Error("failed remove temp directory", zap.String("path", u.tempDir), zap.Error(err))
	}
}
