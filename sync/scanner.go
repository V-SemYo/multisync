package sync

import (
	"fmt"
	"io/fs"
	"multisync/webdav"
	"path/filepath"
	"strings"
	"time"
)

// ScanLocal рекурсивно сканирует локальную папку и возвращает список файлов в формате FileInfo, пропускает скрытые файлы (начинающиеся с точки)
func ScanLocal(localPath string) ([]webdav.FileInfo, error) {
	var files []webdav.FileInfo
	err := filepath.WalkDir(localPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("files not found: %w", walkErr)
		}

		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("file info request error: %w", err)
		}

		relPath := strings.TrimPrefix(path, localPath)
		relPath = strings.TrimLeft(relPath, "/")

		fileInfo := webdav.FileInfo{
			Path:     relPath,
			Name:     d.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format(time.RFC3339),
			IsDir:    d.IsDir(),
		}

		files = append(files, fileInfo)
		return nil
	},
	)

	if err != nil {
		return nil, fmt.Errorf("walk error: %w", err)
	}

	return files, nil
}
