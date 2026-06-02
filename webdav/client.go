package webdav

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// propfindBod - XML-тело запроса PROPFIND. Указывает серверу, какие свойства нужны
const propfindBody = `<?xml version="1.0" encoding="utf-8"?>
<D:propfind xmlns:D="DAV:">
  <D:prop>
    <D:displayname/>
    <D:getcontentlength/>
    <D:getlastmodified/>
    <D:resourcetype/>
  </D:prop>
</D:propfind>`

// WebDAVClient — клиент для взаимодействия с WebDAV-сервером
type WebDAVClient struct {
	BaseURL  string
	User     string
	Password string
	Client   *http.Client
}

// NewWebDAVClient создаёт готовый к работе экземпляр с настроенным HTTP-клиентом
func NewWebDAVClient(baseURL, user, password string) *WebDAVClient {
	return &WebDAVClient{
		BaseURL:  baseURL,
		User:     user,
		Password: password,
		Client:   &http.Client{},
	}
}

// List получает список всех файлов и папок с сервера по указанному пути
func (c *WebDAVClient) List(remotePath string) ([]FileInfo, error) {
	fullURL, err := url.JoinPath(c.BaseURL, remotePath)
	if err != nil {
		return nil, fmt.Errorf("join url error: %w", err)
	}

	body := strings.NewReader(propfindBody)
	req, err := http.NewRequest("PROPFIND", fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(c.User, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 207 {
		return nil, fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response read error: %w", err)
	}

	var ms MultiStatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, fmt.Errorf("xml parse error: %w", err)
	}

	var files []FileInfo
	for _, resp := range ms.Responses {
		files = append(files, resp.ToFileInfo())
	}

	return files, nil
}

// Download скачивает файл с WebDAV-сервера и сохраняет его на локальный диск.
func (c *WebDAVClient) Download(remotePath, localPath string) error {
	fullURL, err := url.JoinPath(c.BaseURL, remotePath)
	if err != nil {
		return fmt.Errorf("join url error: %w", err)
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.SetBasicAuth(c.User, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	localP := filepath.Dir(localPath)
	err = os.MkdirAll(localP, 0755)
	if err != nil {
		return fmt.Errorf("dir error: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("file create error: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	return nil
}

// Upload загружает файл с WebDAV-сервера, если файл на сервере уже существует, он будет перезаписан
func (c *WebDAVClient) Upload(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer file.Close()

	fullURL, err := url.JoinPath(c.BaseURL, remotePath)
	if err != nil {
		return fmt.Errorf("join url error: %w", err)
	}

	req, err := http.NewRequest("PUT", fullURL, file)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.SetBasicAuth(c.User, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// Delete удаляет файл или папку на сервере
func (c *WebDAVClient) Delete(remotePath string) error {
	fullURL, err := url.JoinPath(c.BaseURL, remotePath)
	if err != nil {
		return fmt.Errorf("join url error: %w", err)
	}

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.SetBasicAuth(c.User, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// Mkdir создаёт папку на сервере, если директория уже существует, возвращает ошибку
func (c *WebDAVClient) Mkdir(remotePath string) error {
	fullURL, err := url.JoinPath(c.BaseURL, remotePath)
	if err != nil {
		return fmt.Errorf("join url error: %w", err)
	}

	req, err := http.NewRequest("MKCOL", fullURL, nil)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.SetBasicAuth(c.User, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("response error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusMethodNotAllowed {
		return fmt.Errorf("directory already exists")
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}
