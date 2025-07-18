# grpc_user-manager

необходимые приложения для работы с grpc

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
# swagger документация
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
# миграции для Postgres
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# мокирование
go install github.com/vektra/mockery/v3@latest
```

комманда генерации протофайлов

```bash
protoc --go_out=. --go_opt=module=github.com/Vy4cheSlave/grpc_user-manager \
       --go-grpc_out=. --go-grpc_opt=module=github.com/Vy4cheSlave/grpc_user-manager \
       protos/auth.proto
```

комманды поднятия миграций

```bash
docker-compose up
migrate --path migrations -database "postgres://dbuser:dbpassword@dbhost:dbport/dbname?sslmode=disable" up
```

комманда создания миграции

```bash
migrate create -ext sql -dir migrations -seq "name_migration"
```

комманда создания моков

```bash]
mockery
```