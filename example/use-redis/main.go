package main

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"github.com/gone-io/gone/goner/redis"
	"time"
)

func priest(cemetery gone.Cemetery) error {

	//使用 goner.RedisPriest 函数，将 redis 相关的Goner 埋葬到 Cemetery 中
	_ = goner.RedisPriest(cemetery)

	cemetery.Bury(&redisUser{})
	return nil
}

type redisUser struct {
	gone.Flag

	cache  redis.Cache  `gone:"gone-redis-cache"`
	locker redis.Locker `gone:"gone-redis-locker"`
}

func (r *redisUser) UseCache() {
	key := "gone-address"
	value := "https://github.com/gone-io/gone"

	//设置缓存
	err := r.cache.Put(
		key,            //第一个参数为 缓存的key，类型为 `string`
		value,          // 第二参数为 需要缓存的值，类型为any，可以是任意类型；传入的值会被编码为 `[]byte` 发往redis
		10*time.Second, // 第三个参数为 过期时间，类型为 `time.Duration`;省略，表示不设置过期时间
	)

	if err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	//获取缓存
	var getValue string
	err = r.cache.Get(
		key,       //第一个参数为 缓存的key，类型为 `string`
		&getValue, //第二参数为指针，接收获取缓存的值，类型为any，可以是任意类型；从redis获取的值会被解码为传入的指针类型
	)
	if err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	fmt.Printf("getValue:%v", getValue)
}

func (r *redisUser) LockTime() {
	lockKey := "gone-lock-key"

	//尝试获取锁 并 锁定一段时间
	//返回的第一个参数为一个解锁的函数
	unlock, err := r.locker.TryLock(
		lockKey,        //第一个参数为 锁的key，类型为 `string`
		10*time.Second, //第二参数为 锁的过期时间，类型为 `time.Duration`
	)
	if err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	//操作完后，需要解锁
	defer unlock()

	//获取锁成功后，可以进行业务操作
	//todo
}

func (r *redisUser) LockFunc() {
	lockKey := "gone-lock-key"
	err := r.locker.LockAndDo(
		lockKey, //第一个参数为 锁的key，类型为 `string`
		func() { //第二个参数为 需要执行的函数，类型为 `func()`，代表一个操作
			//获取锁成功后，可以进行业务操作
			//todo
			println("do some options")
		},
		100*time.Second, //第三个参数为 锁的过期时间，类型为 `time.Duration`;第一次加锁和后续锁续期都将使用该值
		5*time.Second,   //第四个参数为 锁续期的间隔时间，类型为 `time.Duration`;周期性检查所是否将过期，如果在下个周期内会过期则对锁续期
	)
	if err != nil {
		fmt.Printf("err:%v", err)
	}
}

func main() {
	gone.Prepare(priest).AfterStart(func(in struct {
		r *redisUser `gone:"*"`
	}) {
		in.r.UseCache()
		in.r.LockTime()
	}).Run()
}
