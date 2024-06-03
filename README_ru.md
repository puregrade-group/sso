# Single Sign-On (SSO)

### langs
- [en](https://github.com/puregrade-group/sso/blob/master/README.md)

### Описание:

[Single Sign-On wiki](https://ru.wikipedia.org/wiki/%D0%A2%D0%B5%D1%85%D0%BD%D0%BE%D0%BB%D0%BE%D0%B3%D0%B8%D1%8F_%D0%B5%D0%B4%D0%B8%D0%BD%D0%BE%D0%B3%D0%BE_%D0%B2%D1%85%D0%BE%D0%B4%D0%B0) технологии.

#### Стек: Go, gRPC, Postgres, Docker

### Структура проекта:
```
├───cmd
├───config // конфиги
├───internal
│   ├───app // Сборка компонентов приложения
│   │   └───grpc // код приложения gRPC сервера
│   ├───config // Структура конфига приложения
│   ├───domain
│   │   └───models // Общие структуры
│   ├───service // Сервисный слой
│   │   ├───acs
│   ├───storage // Слой хранения данных
│   │   └───postgres
│   ├───transport // Слой хранения данных
│   │   └───grpc
│   │       └───auth // Файлы для регистрации/логина юзеров
├───migrations // Файлы миграций
├───storage // Файлы БД
└───tests // Тесты
    ├───migrations
    └───suite
```

### Установка, сборка, запуск:

#### Используемые компоненты:
1. go compiler 1.21.5
2. git 
3. Docker
4. make
5. Postman

#### Порядок действий:

1. Клонируем репозиторий: `git clone https://github.com/puregrade-group/sso ./my/favorite/dir`
2. Устанавливаем зависимости `go mod download` 
3. Создаем базу и наполняем таблицами `make mgrs-up` или `go run ./cmd/migrator/main.go --storage-path=storage/sso.db  --migrations-path=migrations`
4. Для тестов наполняем необходимыми тестовыми данными `make test-mgrs-up` или `go run ./cmd/migrator/main.go --storage-path="./storage/sso.db"  --migrations-path="./tests/migrations" --migrations-table="test"`
5. Запускаем приложение `make run` или `go run ./cmd/main.go --config=./config/config.yaml`
6. Посмотреть пример работы приложения можно запустив тесты `go test`; отправив запросы через [Postman](https://www.postman.com/) или написав свой собственный клиент для этого приложения (Для этого пригодятся файлы .proto. С их помощью можно сгенерировать основу клиентского приложения на любом языке)

или

5. Билдим Docker образ `docker build --tag image-name .`
6. Запускаем контейнер `docker run -p 50051:50051/tcp --name container-name <image_id>`

##### Примеры:

Логи при запуске приложения:
<p align="left"><img width="400px" src="https://github.com/puregrade-group/sso/raw/master/example/execute_log.png" alt="execute_log.png"/></p>

Аутпут из Postman:
<p align="left"><img width="400px" src="https://github.com/puregrade-group/sso/raw/master/example/postman_output.png" alt="postman_output.png"/></p>

Аутпут из тестов:
<p align="left"><img width="400px" src="https://github.com/puregrade-group/sso/raw/master/example/test_output.png" alt="test_output.png"/></p>

