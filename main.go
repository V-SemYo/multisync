package main

import (
	"context"
	"flag"
	"fmt"
	"multisync/config"
	"multisync/logger"
	"multisync/sync"
	"multisync/webdav"
	"os"
	"strconv"
	"syscall"
)

// validateConfig проверяет доступность всех WebDAV серверов из конфига
func validateConfig(cfg *config.Config, log *logger.Logger) {
	for _, job := range cfg.SyncJobs {
		for _, server := range job.WebDav.Servers {
			client := webdav.NewWebDAVClient(server.URL, server.User, server.Password)

			_, err := client.List(server.RemotePath)
			if err != nil {
				log.Error("[%s] server %s not available: %v\n", job.Name, server.URL, err)
				continue

			}

			log.Info("[%s] server %s available\n", job.Name, server.URL)
		}
	}
}

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	jobName := flag.String("job", "", "job name to sync")

	once := flag.Bool("once", false, "run once and exit")

	ctx := context.Background()

	flag.Parse()
	command := flag.Arg(0)

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config error: %v\n", err)

		os.Exit(1)
	}

	log, err := logger.NewLogger(conf.Global.LogLevel, conf.Global.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	// проверяет доступность всех WebDAV серверов из конфига
	case "validate":
		validateConfig(conf, log)

	// запускает цикл синхронизацию в одиночном формате или пока контекст не отменён
	case "start":
		strPID := strconv.Itoa(os.Getpid())

		err = os.WriteFile(conf.Global.PidFile, []byte(strPID), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writefile error: %v\n", err)
		}
		defer os.Remove(conf.Global.PidFile)
		for _, job := range conf.SyncJobs {
			worker := sync.NewWorker(job, log)
			if *once {
				err = worker.RunOnce(ctx)
			} else {
				err = worker.Run(ctx)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "start error: %v\n", err)
				os.Exit(1)
			}
		}
		log.Info("Sync completed")

	// останавливает синхронизацию сигналом SIGTERM
	case "stop":
		pidFile, err := os.ReadFile(conf.Global.PidFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read file error: %v\n", err)
			os.Exit(1)
		}

		pid, err := strconv.Atoi(string(pidFile))
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			os.Exit(1)
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "find process error: %v\n", err)
			os.Exit(1)
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Fprintf(os.Stderr, "signal error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Sync stopped")
		log.Info("Sync stopped")

	// информирует о процессе синхронизации проверкой сигнала
	case "status":
		pidFile, err := os.ReadFile(conf.Global.PidFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read file error: %v\n", err)
			os.Exit(1)
		}

		pid, err := strconv.Atoi(string(pidFile))
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			os.Exit(1)
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "find process error: %v\n", err)
			os.Exit(1)
		}

		err = process.Signal(syscall.Signal(0))
		if err != nil {
			fmt.Println("not running")
			log.Info("process not running")
		} else {
			fmt.Println("running")
			log.Info("process running")
		}

	// проверяет задачу синхронизации
	case "sync":
		var found bool

		for _, job := range conf.SyncJobs {
			if job.Name == *jobName {
				worker := sync.NewWorker(job, log)
				err := worker.RunOnce(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "run error: %v\n", err)
					os.Exit(1)
				}

				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "job not found")
			log.Info("job not found")
		}

	default:
		for _, job := range conf.SyncJobs {
			worker := sync.NewWorker(job, log)
			if *once {
				err = worker.RunOnce(ctx)
			} else {
				err = worker.Run(ctx)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "run error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println("Sync completed")
		log.Info("Sync completed")
	}
}
