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

## Postman
You can access postman collection for tests [on this url](https://www.postman.com/avionics-operator-63001856/workspace/genesis-weather).