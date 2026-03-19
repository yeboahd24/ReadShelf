package cloudinary

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dominic/readshelf/internal/core/port/outbound"
)

type fileStore struct {
	cloudName string
	apiKey    string
	apiSecret string
}

func NewFileStore(cloudName, apiKey, apiSecret string) outbound.FileStore {
	return &fileStore{
		cloudName: cloudName,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (f *fileStore) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"public_id":     key,
		"resource_type": "raw",
		"timestamp":     timestamp,
	}

	signature := f.sign(params)

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		writer.WriteField("public_id", key)
		writer.WriteField("timestamp", timestamp)
		writer.WriteField("api_key", f.apiKey)
		writer.WriteField("signature", signature)

		part, err := writer.CreateFormFile("file", "upload.pdf")
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, reader); err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/raw/upload", f.cloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, pr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudinary upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudinary upload failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (f *fileStore) SignedURL(_ context.Context, key string) (string, error) {
	// Cloudinary raw files are accessible via a predictable URL.
	url := fmt.Sprintf("https://res.cloudinary.com/%s/raw/upload/%s", f.cloudName, key)
	return url, nil
}

func (f *fileStore) Delete(ctx context.Context, key string) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"public_id":     key,
		"resource_type": "raw",
		"timestamp":     timestamp,
	}

	signature := f.sign(params)

	form := strings.NewReader(fmt.Sprintf(
		"public_id=%s&timestamp=%s&api_key=%s&signature=%s",
		key, timestamp, f.apiKey, signature,
	))

	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/raw/destroy", f.cloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, form)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudinary delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudinary delete failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result string `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return nil
}

// sign generates a Cloudinary API signature.
func (f *fileStore) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "resource_type" && k != "api_key" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	toSign := strings.Join(parts, "&") + f.apiSecret
	hash := sha1.Sum([]byte(toSign))
	return hex.EncodeToString(hash[:])
}
