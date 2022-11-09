# gone-redis

This lib integrated the basic operation of redis with [redigo](github.com/gomodule/redigo/redis).

## How to Use

### 0. Redis Server Config

The Lib use the [gone-config](../config), so we can config in config files(config/default.properties,
config/${env}.properties).

- redis.server: Redis server address, example: `localhost:6379`.
- redis.password: Redis server password.
- redis.db: Redis db index, which you want to use.
- redis.max-idle: Idle connection count in the redis connection pool.
- redis.max-active: Max active connection count in the redis connection pool.
- redis.cache.prefix: A prefix string use to isolate different applications. It's recommended, if Your redis is used by
  multiple applications. if `redis.cache.prefix=app-x`, `Cache.Set("the-module-cache-key", value)` will set value on
  redis key: `app-x#the-module-cache-key` .

### 1. Distributed Cache with Redis

```go
package demo

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
)

func NewService() gone.Goner {
	return &service{}
}

type service struct {
	gone.Flag
	cache redis.Cache `gone:"gone-redis-cache"` //Tag label
}

func (s *service) Use() {

	type ValueStruct struct {
		X int
		Y int
		//...
	}

	var v ValueStruct

	// set cache to redis
	err := s.cache.Set("cache-key", v)
	if err != nil {
		//deal err
	}

	//get value from cache
	err = s.cache.Get("cache-key", &v)
	if err != nil {
		//deal err
	}

	//...
}
```

### 2. Distributed Locks with Redis

```go
package demo

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
	"time"
)

func NewService() gone.Goner {
	return &service{}
}

type service struct {
	gone.Flag
	locker redis.Locker `gone:"gone-redis-locker"`
}

//UseTryLock use Locker.TryLock
func (s *service) UseTryLock() {
	unlock, err := s.locker.TryLock("a-lock-key-in-redis", 10*time.Second)
	if err != nil {
		// deal err
	}
	defer unlock()
	// ... other operations
}

//UseLockAndDo use Locker.LockAndDo
func (s *service) UseLockAndDo() {
	err := s.locker.LockAndDo("a-lock-key-in-redis", func() {
		//do your business
		//...
		//It's automatically unlocked at the end of the function.
		//Otherwise, the lock will be automatically renewed until the function is finished.

	}, 10*time.Second, 2*time.Second)

	if err != nil {
		//deal err
	}
}
```

### 3. Operations on Key

```go
package demo

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
	"time"
)

func NewService() gone.Goner {
	return &service{}
}

type service struct {
	gone.Flag
	key redis.Key `gone:"gone-redis-locker"`
}

func (s *service) UseTryLock() {

	// set the expiry time 
	s.key.Expire("the-key-in-redis", 2*time.Second)
	s.key.ExpireAt("the-key-in-redis", time.Now().Add(10*time.Minute))

	// get key ttl
	s.key.Ttl("the-key-in-redis")

	// and so on
}
```

### 4. Redis hashes

```go
package demo

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
)

func NewService() gone.Goner {
	return &service{}
}

type service struct {
	gone.Flag
	h redis.Hash `gone:"gone-redis-provider,key-in-redis"` //use gone-redis-provider tag provide a redis.Hash to operate Hashes on `key-in-redis`
}

func (s *service) Use() {
	s.h.Set("a-field", "some thing")
	var str string
	s.h.Get("a-field", &str)

	//...
}
```

### 5. Provider

> Provider can isolate key namespace in app again. For Example, you want use `app-x#module-a` as redis prefix for module
> A, and use `app-x#module-b` as redis prefix for module B. You can use it like below.

```go
package A

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
)

//in module A
//...
type service struct {
	gone.Flag
	cache  redis.Cache  `gone:"gone-redis-provider,module-a"` //use cache 
	key    redis.Key    `gone:"gone-redis-provider,module-a"` //use key
	locker redis.Locker `gone:"gone-redis-provider,module-a"` //use locker
}
```

```go
package B

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
)

//in module B
//...
type service struct {
	gone.Flag
	cache  redis.Cache  `gone:"gone-redis-provider,module-b"` //use cache 
	key    redis.Key    `gone:"gone-redis-provider,module-b"` //use key
	locker redis.Locker `gone:"gone-redis-provider,module-b"` //use locker
}
```

If the key value is in config files, you can use `gone:"gone-redis-provider,config=config-file-key,default=default-val"`
.

```go
package A

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
)

//in module B
//...
type service struct {
	gone.Flag
	cache redis.Cache `gone:"gone-redis-provider,config=app.module-a.redis.prefix"` //use cache
}
```

### 5. Redis Pool

You can use `redis.Pool` directly to read/write redis, which provided by [redigo](github.com/gomodule/redigo/redis).

```go
package demo

import (
	"github.com/gone-io/gone/goner/redis"
	"github.com/gone-io/gone"
)

type service struct {
	gone.Flag
	pool redis.Pool `gone:"gone-redis-pool"`
}

func (s *service) Use() {
	conn := s.pool.Get()
	defer s.pool.Close(conn)

	//do some operation
	//conn.Do(/*...*/)

	//send a command
	//conn.Send(/*...*/)
}
```

## Test

> The test script below depend on [Make](https://cmake.org/download/) and [Docker](https://www.docker.com/get-started/)
> which is used to run redis.

```shell
make test
```
