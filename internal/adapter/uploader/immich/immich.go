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
	handleAssetsDevice string = "%s/api/assets/device/%s"
	handleAssets       string = "%s/api/assets"
	httpTimeout               = 30 * time.Second
)

var httpClient = &http.Client{
	Timeout: httpTimeout,
}

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

	response, err := httpClient.Do(resp)
	if err != nil {
		return nil, fmt.Errorf("failed do response: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", response.StatusCode, string(body))
	}

	sFiles := []string{}
	err = json.Unmarshal(body, &sFiles)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal body: %w", err)
	}
	for _, s := range sFiles {
		// deviceAssetId format: {deviceId}-{filename}-{unix_timestamp}
		// Extract filename by removing known prefix "{deviceId}-" and suffix "-{timestamp}"
		name := s
		prefix := i.cfg.Path + "-"
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			// Remove trailing "-{timestamp}" where timestamp is digits
			if lastDash := strings.LastIndex(name, "-"); lastDash > 0 {
				// Check if the part after last dash is a timestamp (all digits)
				ts := name[lastDash+1:]
				if len(ts) > 0 {
					isTimestamp := true
					for _, c := range ts {
						if c < '0' || c > '9' {
							isTimestamp = false
							break
						}
					}
					if isTimestamp {
						name = name[:lastDash]
					}
				}
			}
			r = append(r, &models.File{
				Name: name,
			})
			i.log.Debug("liststore: parsed with prefix",
				zap.String("deviceAssetId", s),
				zap.String("extracted_name", name),
			)
		} else {
			// Файл имеет другой deviceId — не можем извлечь имя стандартным способом
			i.log.Debug("liststore: skipped - no matching prefix",
				zap.String("deviceAssetId", s),
				zap.String("expected_prefix", prefix),
			)
		}
	}

	return r, nil
}

func (i *Immich) Write(f *os.File, name string) error {
	f.Seek(0, 0)

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

	response, err := httpClient.Do(resp)
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

	// Некоторые версии Immich возвращают 500 при попытке загрузить дубликат
	// (duplicate key value violates unique constraint "UQ_assets_owner_checksum").
	// В таком случае считаем файл уже существующим и не возвращаем ошибку.
	if response.StatusCode == http.StatusInternalServerError {
		bodyStr := string(body)
		if strings.Contains(bodyStr, "Internal Server Error") {
			i.log.Debug("possible duplicate file (server returned 500)", zap.String("name", name), zap.String("response", bodyStr))
			return nil
		}
	}

	return fmt.Errorf("failed store file: status=%d, body=%s", response.StatusCode, string(body))
}

func (i *Immich) Close() error {
	return nil
}
