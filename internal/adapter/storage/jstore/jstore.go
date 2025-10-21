package jstore

import (
	"encoding/json"
	"fmt"
	"os"
	"upsync/internal/adapter/models"
)

type JStore struct {
}

func New() (*JStore, error) {
	s := &JStore{}

	return s, nil
}

func (s *JStore) GetSynces() ([]*models.SyncItem, error) {
	r := []*models.SyncItem{}

	bData, err := os.ReadFile("./data.remotes.json")
	if err != nil {
		return nil, fmt.Errorf("failed read file: %w", err)
	}

	err = json.Unmarshal(bData, &r)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal json doc: %w", err)
	}

	return r, nil
}
