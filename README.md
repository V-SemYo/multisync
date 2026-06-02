# multisync

Многопоточная синхронизация локальных папок с несколькими WebDAV серверами. Следит за изменениями в локальных папках и автоматически синхронизирует файлы.

## Установка

git clone https://github.com/V-SemYo/multisync.git
cd multisync
go build -o multisync .

## Быстрый старт

Запусти тестовый WebDAV сервер в Docker:

docker run -d --name webdav-test -p 8080:80 -e USERNAME=test -e PASSWORD=pass bytemark/webdav

Создай конфиг config.yaml (пример ниже). Создай тестовую папку и файл:

mkdir -p ./local_data && echo "Hello WebDAV" > ./local_data/hello.txt

Запусти синхронизацию:

./multisync --config config.yaml --once

Проверь результат:

curl -u test:pass http://localhost:8080/hello.txt

## Команды

| Команда | Пример |
|---------|--------|
| `run` (демон) | `./multisync --config config.yaml` |
| `run --once` | `./multisync --config config.yaml --once` |
| `validate` | `./multisync validate --config config.yaml` |
| `sync --job` | `./multisync sync --job backup --config config.yaml` |
| `start` | `./multisync start --config config.yaml` |
| `stop` | `./multisync stop --config config.yaml` |
| `status` | `./multisync status --config config.yaml` |

## Режимы синхронизации

both — двусторонняя синхронизация: изменения копируются в обе стороны.
upload-only — только загрузка локальных файлов на сервер.
download-only — только скачивание изменений с сервера.

## Пример конфига (config.yaml)

```yaml
sync_jobs:
  - name: "backup"
    local_path: "/home/user/documents"
    webdav:
      url: "http://localhost:8080"
      user: "test"
      password: "pass"
      remote_path: "/backups"
    interval: "30s"
    direction: "both"
    exclude: ["*.tmp", "*.log"]

global:
  log_level: "info"
  log_file: "/var/log/multisync.log"
  pid_file: "/var/run/multisync.pid"
```

Поле `webdav` может быть одним сервером или списком серверов:

```yaml
webdav:
  - url: "http://server1/webdav"
    user: "user1"
    password: "pass1"
    remote_path: "/shared"
  - url: "http://server2/webdav"
    user: "user2"
    password: "pass2"
    remote_path: "/incoming"
```


Поле exclude — необязательное. Принимает паттерны файлов, которые нужно игнорировать (*.tmp, *.log, .DS_Store).

## Логирование

Уровни: debug, info, warn, error. Настраивается в global.log_level. Логи выводятся в консоль и в файл (global.log_file).

debug — всё подряд, включая отладку.
info — важные события (загружено 5 файлов).
warn — предупреждения (сервер не ответил).
error — только ошибки.

## Тестирование

go test ./... -v

## Makefile

make build   — собрать бинарник multisync
make test    — запустить все тесты
make vet     — проверить код (go vet)
make clean   — удалить бинарник

## Требования

Go 1.21+
WebDAV сервер (можно запустить в Docker: bytemark/webdav)

## Структура проекта

multisync/
├── main.go
— CLI (7 команд)
├── config/
— YAML конфиг (структуры и парсер)
├── webdav/
— WebDAV клиент (PROPFIND, GET, PUT, DELETE, MKCOL)
├── sync/
— движок синхронизации (сканер, сравнение, воркер)
├── logger/
— логгер с уровнями
├── Makefile
└── README.md
