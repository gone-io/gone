test:
	docker-compose -f testdata/docker-compose.yaml up -d
	make run-test
	docker-compose -f testdata/docker-compose.yaml down


run-test:
	go test ./... -v