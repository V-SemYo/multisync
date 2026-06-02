package sync

import (
	"fmt"
	"multisync/webdav"
	"time"
)

// Compare сравнивает локальный и удалённый списки файлов
func Compare(local, remote []webdav.FileInfo) (toUpload, toDownload []webdav.FileInfo, toDeleteLocal, toDeleteRemote []webdav.FileInfo) {

	localFiles := make(map[string]webdav.FileInfo)
	remoteFiles := make(map[string]webdav.FileInfo)

	for _, file := range local {
		localFiles[file.Path] = file
	}

	for _, file := range remote {
		remoteFiles[file.Path] = file
	}

	for _, localFile := range local {
		remoteFile, ok := remoteFiles[localFile.Path]

		if !ok {
			toUpload = append(toUpload, localFile)

		} else {
			newer, err := IsNewer(localFile.Modified, remoteFile.Modified)

			if err != nil {
				continue
			}
			if newer {
				toUpload = append(toUpload, localFile)

			} else {
				toDownload = append(toDownload, remoteFile)
			}
		}
	}

	for _, remoteFile := range remote {
		_, ok := localFiles[remoteFile.Path]

		if !ok {
			toDownload = append(toDownload, remoteFile)
		}
	}

	return toUpload, toDownload, nil, nil
}

// IsNewer возвращает true, если data1 новее data2. Даты в формате RFC3339
func IsNewer(data1, data2 string) (bool, error) {
	t1, err := time.Parse(time.RFC3339, data1)
	if err != nil {
		return false, fmt.Errorf("data parse error: %w", err)
	}

	t2, err := time.Parse(time.RFC3339, data2)
	if err != nil {
		return false, fmt.Errorf("data parse error: %w", err)
	}

	return t1.After(t2), nil
}
