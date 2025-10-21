package models

import (
	"strings"
	"time"
)

type TProtocol string

const (
	RP_FTP    TProtocol = "ftp"
	RP_IMMICH TProtocol = "immich"
	RP_DIR    TProtocol = "dir"
)

type arrString string

type Remote struct {
	Protocol   TProtocol `json:"protocol"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Login      string    `json:"login"`
	Password   string    `json:"password"`
	Path       string    `json:"path"`
	Extensions arrString `json:"extensions"`
}

func (s arrString) Slice() []string {
	return strings.Split(string(s), ",")
}

type Storage struct {
	Protocol TProtocol `json:"protocol"`
	Address  string    `json:"address"`
	Host     string    `json:"host"`
	Port     int       `json:"port"`
	APIKey   string    `json:"apiKey"`
	Path     string    `json:"path"`
}

type SyncItem struct {
	Title  string  `json:"title"`
	Remote Remote  `json:"remote"`
	Store  Storage `json:"store"`
}

type File struct {
	Name  string
	IsDir bool
	Size  uint64
	Time  time.Time
	Path  string
}
