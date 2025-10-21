package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
	"upsync/internal/adapter/models"

	"go.uber.org/zap"
)

const (
	handleAssetsDevice string = "%s/api/assets/device/%s" //получить список файлов ...
	handleAssets       string = "%s/api/assets"           //получить список файлов ...
)

type Immich struct {
	cfg Config
	log *zap.Logger
}

func New(cfg Config, log *zap.Logger) (*Immich, error) {
	i := &Immich{
		cfg: cfg,
		log: log,
	}

	return i, nil
}

func (i *Immich) ListStore() ([]*models.File, error) {
	r := []*models.File{}

	url := fmt.Sprintf(handleAssetsDevice, i.cfg.Address, i.cfg.Path)
	resp, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed create request: %w", err)
	}
	resp.Header.Add("x-api-key", i.cfg.APIKey)

	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		return nil, fmt.Errorf("failed do response: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read body: %w", err)
	}

	sFiles := []string{}
	err = json.Unmarshal(body, &sFiles)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal body: %w", err)
	}
	for _, s := range sFiles {
		arS := strings.Split(s, "-")
		if len(arS) > 1 {
			s = strings.Replace(strings.Replace(s, arS[0]+"-", "", 1), "-"+arS[len(arS)-1], "", 1)
		}
		r = append(r, &models.File{
			Name: s,
		})
	}

	return r, nil
}

func (i *Immich) Write(f *os.File, name string) error {
	f, err := os.Open(f.Name())
	if err != nil {
		return fmt.Errorf("failed open temp file: %w", err)
	}
	defer f.Close()
	// _, err = f.Write(*data)
	// if err != nil {
	// 	return fmt.Errorf("failed write temp file: %w", err)
	// }

	fInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed getting info of file: %w", err)
	}

	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	fWriter, err := w.CreateFormFile("assetData", name)
	if err != nil {
		return fmt.Errorf("failed asset data: %w", err)
	}
	_, err = io.Copy(fWriter, f)
	if err != nil {
		return fmt.Errorf("failed write file to form: %w", err)
	}

	values := map[string]string{
		"deviceAssetId":  fmt.Sprintf("%s-%s-%v", i.cfg.Path, name, time.Now().Unix()),
		"deviceId":       i.cfg.Path,
		"fileCreatedAt":  fInfo.ModTime().Format(time.DateTime),
		"fileModifiedAt": fInfo.ModTime().Format(time.DateTime),
	}

	for k, v := range values {
		nameWriter, err := w.CreateFormField(k)
		if err != nil {
			return fmt.Errorf("failed create field form: %w", err)
		}
		_, err = nameWriter.Write([]byte(v))
		if err != nil {
			return fmt.Errorf("failed write data field form: %w", err)
		}
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed close ")
	}

	url := fmt.Sprintf(handleAssets, i.cfg.Address)
	resp, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return fmt.Errorf("failed create request: %w", err)
	}
	resp.Header.Add("x-api-key", i.cfg.APIKey)
	resp.Header.Add("Content-Type", w.FormDataContentType())

	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		return fmt.Errorf("failed do response: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed read body: %w", err)
	}

	if response.StatusCode == http.StatusOK {
		i.log.Debug("duplicate file", zap.String("response", string(body)), zap.String("name", name))
		return nil
	}

	if response.StatusCode == http.StatusCreated {
		i.log.Debug("file saved", zap.String("response", string(body)))
		return nil
	}

	i.log.Debug("faile store file", zap.String("response", string(body)))
	return nil
}

func (i *Immich) Close() error {
	return nil
}
