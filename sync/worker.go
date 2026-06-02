package sync

import (
	"context"
	"fmt"
	"multisync/config"
	"multisync/logger"
	"multisync/webdav"
	"path/filepath"
	"time"
)

// Worker — воркер для одной задачи синхронизации, хранит задачу из конфига, параметры сервера и готовый WebDAV-клиент
type Worker struct {
	Job    config.SyncJob
	Server config.WebDAVServer
	Client *webdav.WebDAVClient
	Logger *logger.Logger
}

// NewWorker — создаёт воркер для задачи
func NewWorker(job config.SyncJob, log *logger.Logger) *Worker {
	server := job.WebDav.Servers[0]

	client := webdav.NewWebDAVClient(server.URL, server.User, server.Password)

	return &Worker{
		Job:    job,
		Client: client,
		Server: server,
		Logger: log,
	}

}

// Run запускает цикл синхронизации. Работает, пока контекст не отменён
func (w *Worker) Run(ctx context.Context) error {
	w.Logger.Info("sync started")

	interval, err := time.ParseDuration(w.Job.Interval)
	if err != nil {
		return fmt.Errorf("time parse error: %w", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return nil

		case <-ticker.C:
			localFiles, err := ScanLocal(w.Job.LocalPath)
			if err != nil {
				w.Logger.Error("scan files error: %v", err)
				continue
			}

			remoteFiles, err := w.Client.List(w.Server.RemotePath)
			if err != nil {
				w.Logger.Error("get server files error: %v", err)
				continue
			}

			toUpload, toDownload, _, _ := Compare(localFiles, remoteFiles)

			for _, localFile := range toUpload {
				if localFile.IsDir {
					continue
				}
				localPath := filepath.Join(w.Job.LocalPath, localFile.Path)
				remotePath := w.Server.RemotePath + "/" + localFile.Path
				err := w.Client.Upload(localPath, remotePath)
				if err != nil {
					w.Logger.Error("upload file %s: %v", localFile.Path, err)
					continue
				}
				w.Logger.Info("uploaded %s", localFile.Path)
			}

			for _, remoteFile := range toDownload {
				if remoteFile.IsDir {
					continue
				}
				localPath := filepath.Join(w.Job.LocalPath, remoteFile.Path)
				remotePath := w.Server.RemotePath + "/" + remoteFile.Path
				err := w.Client.Download(remotePath, localPath)
				if err != nil {
					w.Logger.Error("download file %s: %v", remoteFile.Path, err)
					continue
				}
				w.Logger.Info("downloaded %s", remoteFile.Path)
			}
		}
	}
}

// RunOnce выполняет один полный цикл синхронизации и завершается
func (w *Worker) RunOnce(ctx context.Context) error {
	w.Logger.Info("sync started")

	localFiles, err := ScanLocal(w.Job.LocalPath)
	if err != nil {
		w.Logger.Error("scan files error: %v", err)
		return err
	}

	remoteFiles, err := w.Client.List(w.Server.RemotePath)
	if err != nil {
		w.Logger.Error("get server files error: %v", err)
		return err
	}

	toUpload, toDownload, _, _ := Compare(localFiles, remoteFiles)

	for _, localFile := range toUpload {
		if localFile.IsDir {
			continue
		}
		localPath := filepath.Join(w.Job.LocalPath, localFile.Path)
		remotePath := w.Server.RemotePath + "/" + localFile.Path
		err := w.Client.Upload(localPath, remotePath)
		if err != nil {
			w.Logger.Error("upload file %s: %v", localFile.Path, err)
			continue
		}
		w.Logger.Info("uploaded %s", localFile.Path)
	}

	for _, remoteFile := range toDownload {
		if remoteFile.IsDir {
			continue
		}
		localPath := filepath.Join(w.Job.LocalPath, remoteFile.Path)
		remotePath := w.Server.RemotePath + "/" + remoteFile.Path
		err := w.Client.Download(remotePath, localPath)
		if err != nil {
			w.Logger.Error("download file %s: %v", remoteFile.Path, err)
			continue
		}
		w.Logger.Info("downloaded %s", remoteFile.Path)
	}
	return nil
}
