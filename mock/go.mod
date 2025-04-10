module github.com/gone-io/gone/v2/mock

go 1.24.1

require (
	github.com/gone-io/gone/v2 v2.0.10
	go.uber.org/mock v0.5.1
)


replace (
	github.com/gone-io/gone/v2 => ../
)