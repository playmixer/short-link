# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).


# profiler
http://127.0.0.1:8080/debug/pprof/

# покрытие тестами
```cmd
go test -v -coverpkg=./... -coverprofile=profile.cov ./...
go tool cover -func profile.cov
```

# анализатор
```cmd
go build -o ./staticlint.exe .\cmd\staticlint\main.go
staticlint.exe .\...
```
или
```cmd
go run .\cmd\staticlint\main.go .\...
```

# build
```bash
go build -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git show --oneline -s)'" ./cmd/shortener/main.go
```


# generat TLS
[https://github.com/denji/golang-tls](https://github.com/denji/golang-tls)

#### Generate private key (.key)
```
# Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out server.key 2048

# Key considerations for algorithm "ECDSA" ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl ecparam -genkey -name secp384r1 -out server.key
```
#### Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)
```
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

## Установка PROTO Generate
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
```
go get github.com/golang/protobuf/protoc-gen-go
```
## Обновить GRPC
go get -u google.golang.org/grpc
### windows
https://github.com/protocolbuffers/protobuf/releases

## Генерация сервиса GRPC
выполнить в папке internal\adapters\api\grpch
```
protoc --go_out=./proto --go_opt=paths=source_relative --go-grpc_out=./proto --go-grpc_opt=paths=source_relative shorten.proto
```


# GRPC API
## Auth
### Login 
*Request*:
```
{
    "id": "identificator"
}
```
*Response*:
```
{
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1bmlxdWVJRCI6ImxhYm9ycGlzaWNpbmcifQ.mUisWUcaIUwo0bFBOdonZVj1Hm-7O6xBAJRocX-moPg"
}
```

Для работы с api передать в метаданных "token" со значением *access_token*
терубуют авторизацию: **NewShort, NewShorts, GetURLByShort, GetUserURLs, DeleteUserURLs**

### GetStatus
в матаданных "x-real-ip" передать ip из доверенной зоны