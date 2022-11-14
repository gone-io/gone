build:
	go build -ldflags="-w -s" -tags musl -o bin/gone main.go

test:
	go install github.com/golang/mock/mockgen@latest
	make build
	mockgen -package=mock -source=mock/testdata/interface.go|./bin/gone mock -o mock/testdata/mock/interface.go

