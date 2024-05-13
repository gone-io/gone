gone:
	make install-gone
	go mod tidy
	go generate ./...

run:
	make gone
	go run cmd/server/main.go

build:
	make gone
	go build -ldflags="-w -s" -tags musl -o bin/server cmd/server/main.go


build-docker:
	docker-compose build

run-in-docker:
	docker-comopose up


# for test
test:
	make install-mockgen
	make gone
	go test ./... -v

install-gone:
	go install github.com/gone-io/gone/tools/gone@latest

install-mockgen:
	go install github.com/golang/mock/mockgen@latest