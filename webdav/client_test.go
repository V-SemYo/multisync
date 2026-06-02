package webdav

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestList проверяет, что метод List правильно отправляет PROPFIND-запрос и корректно парсит XML-ответ от сервера в структуру []FileInfo
func TestList(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PROPFIND" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		xmlResponse := `<?xml version="1.0" encoding="utf-8"?>
<D:multistatus xmlns:D="DAV:">
  <D:response>
    <D:href>/test/</D:href>
    <D:propstat>
      <D:prop>
        <D:displayname>test</D:displayname>
        <D:getcontentlength>0</D:getcontentlength>
        <D:getlastmodified>Mon, 12 Apr 2025 15:04:05 GMT</D:getlastmodified>
        <D:resourcetype><D:collection/></D:resourcetype>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusMultiStatus)
		w.Write([]byte(xmlResponse))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewWebDAVClient(server.URL, "user", "password")

	files, err := client.List("/test")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if !files[0].IsDir {
		t.Error("expected IsDir = true, got false")
	}

	if files[0].Path != "/test/" {
		t.Errorf("expected path /test/, got %s", files[0].Path)
	}
}

// TestDownload проверяет, что метод Download скачивает файл с сервера и сохраняет его содержимое в указанный локальный путь
func TestDownload(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/hello.txt" {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello WebDAV"))
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewWebDAVClient(server.URL, "user", "password")

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "hello.txt")

	err := client.Download("/hello.txt", localPath)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	file, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}

	expected := "Hello WebDAV"
	got := string(file)

	if expected != got {
		t.Errorf("expected: %q\n got: %q", expected, got)
	}

}

// TestUpload проверяет, что метод Upload читает локальный файл и отправляет его содержимое на сервер через PUT-запрос
func TestUpload(t *testing.T) {
	var uploadedData []byte
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read file error", http.StatusInternalServerError)
			return
		}
		uploadedData = data

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write(uploadedData)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewWebDAVClient(server.URL, "user", "password")

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "upload.txt")

	err := os.WriteFile(tmpFile, []byte("Hello Upload"), 0644)
	if err != nil {
		t.Errorf("write file error: %v", err)
	}

	err = client.Upload(tmpFile, "/upload.txt")
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	expected := "Hello Upload"

	if expected != string(uploadedData) {
		t.Errorf("expected: %q\n got: %q", expected, uploadedData)
	}

}
