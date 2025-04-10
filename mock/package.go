package mock

//go:generate mockgen -source=../interafce.go -destination=./mock.go -package=mock

//go:generate mockgen -source=../config.go -destination=./config.go -package=mock

//go:generate mockgen -source=../logger.go -destination=./logger.go -package=mock
