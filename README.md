# Some kind of weather notification service

## Setup
_if you want to use linter you should install it before_
```cmd
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
```
[more info see here](https://golangci-lint.run/welcome/install/)

## To run
_dont forget running docker before_
```cmd
make up
```

## About
I used Gin, SQL and migrate.

Also i have added postman collection for tests.

## File structure

```
.
├── Dockerfile
├── Makefile
├── README.md
├── bin
│   └── weather-service
├── cmd
│   └── main.go
├── docker-compose.yaml
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   ├── api.go
│   │   ├── handlers
│   │   │   ├── handlers.go
│   │   │   ├── subscription.go
│   │   │   └── weather.go
│   │   └── middleware
│   │       └── middleware.go
│   ├── application
│   │   └── application.go
│   ├── config
│   │   └── config.go
│   ├── database
│   │   ├── database.go
│   │   └── migrations
│   │       ├── 000001_create_subscription_table.down.sql
│   │       └── 000001_create_subscription_table.up.sql
│   ├── env
│   │   └── env.go
│   ├── mailer
│   │   └── mailer.go
│   ├── models
│   │   ├── subscription.go
│   │   └── weather.go
│   ├── store
│   │   ├── storage.go
│   │   └── subscription.go
│   └── weather
│       ├── adapter.go
│       └── weather.go
└── postman
    └── weather.postman_collection.json

17 directories, 27 files
```
