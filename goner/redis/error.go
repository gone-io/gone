package redis

import "github.com/gone-io/gone"

// 错误代码：gone-redis内部错误代码编码空间:1501~1599

const (
	//CacheProviderNeedKey cache provider need a key
	CacheProviderNeedKey = 1501 + iota
	KeyNoExpiration
)

func CacheProviderNeedKeyError() gone.Error {
	return gone.NewInnerError(CacheProviderNeedKey, "redis cache provider need a key")
}

func KeyNoExpirationError() gone.Error {
	return gone.NewInnerError(KeyNoExpiration, "There is no expiration time on redis key")
}
