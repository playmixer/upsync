package storage

import (
	"fmt"
	"upsync/internal/adapter/models"
	"upsync/internal/adapter/storage/jstore"
)

type Store interface {
	GetSynces() ([]*models.SyncItem, error)
}

var (
	_ Store = &jstore.JStore{}
)

func New() (Store, error) {
	s, err := jstore.New()
	if err != nil {
		return nil, fmt.Errorf("failed initialize jstore: %w", err)
	}

	return s, nil
}
