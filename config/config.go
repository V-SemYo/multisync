package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config — корневая структура конфигурации приложения
type Config struct {
	SyncJobs []SyncJob    `yaml:"sync_jobs"`
	Global   GlobalConfig `yaml:"global"`
}

// SyncJob — одна задача синхронизации
type SyncJob struct {
	Name      string       `yaml:"name"`
	LocalPath string       `yaml:"local_path"`
	WebDav    WebDAVTarget `yaml:"webdav"`
	Interval  string       `yaml:"interval"`
	Direction string       `yaml:"direction"`
	Exclude   []string     `yaml:"exclude,omitempty"`
}

// WebDAVServer — параметры одного WebDAV сервера
type WebDAVServer struct {
	URL        string `yaml:"url"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	RemotePath string `yaml:"remote_path"`
}

// WebDAVTarget — обёртка, позволяющая указать один сервер или список
type WebDAVTarget struct {
	Servers []WebDAVServer
}

// GlobalConfig — глобальные настройки приложения
type GlobalConfig struct {
	LogLevel string `yaml:"log_level"`
	LogFile  string `yaml:"log_file"`
	PidFile  string `yaml:"pid_file"`
}

// Позволяет указать в YAML как один сервер (объект), так и список серверов
func (t *WebDAVTarget) UnmarshalYAML(value *yaml.Node) error {
	var servers []WebDAVServer
	if err := value.Decode(&servers); err == nil {
		t.Servers = servers
		return nil
	}

	var single WebDAVServer
	if err := value.Decode(&single); err == nil {
		t.Servers = []WebDAVServer{single}
		return nil
	}
	return fmt.Errorf("webdav must be a single server or a list of servers")
}
