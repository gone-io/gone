run:
	make gone
	go run cmd/server/main.go

gone:
	make install-gone
	gone -s internal -p internal -f Priest -o internal/priest.go

install-gone:
	go install github.com/gone-io/gone/tools/gone@v0.0.4