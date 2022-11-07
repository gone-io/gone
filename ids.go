package gone

// Gone框架中的内置组件ID
const (
	// IdGoneHeaven 天堂组件的ID，代码了程序本身，Gone程序启动时默认注入
	IdGoneHeaven = "gone-heaven"

	// IdGoneCemetery 坟墓组件的ID，是完成依赖注入的关键组件，Gone程序启动时默认注入
	IdGoneCemetery = "gone-cemetery"

	//IdGoneTestKit 测试箱，调用 gone.Test 或者 gone.TestAt 时，会将测试箱注入到程序；非测试代码中不应该注入该组件
	IdGoneTestKit = "gone-test-kit"

	// 配置、日志、Tracer 一起构成Gone框架的基础Goner，可以使用 [goner.BasePriest](goner#BasePriest) 牧师函数批量安葬

	//IdConfig 配置 Goner 的ID，提过能配置能力
	IdConfig = "config"
	//IdGoneConfigure 配置器 Goner 的ID
	IdGoneConfigure = "gone-configure"
	// IdGoneTracer Tracer Goner 的ID，提供日志追踪能力
	IdGoneTracer = "gone-tracer"
	// IdGoneLogger 日志 Goner 的ID，用于日志打印
	IdGoneLogger = "gone-logger"

	//IdGoneCumx [cmux Goner](/goner/cmux#Server) ID
	IdGoneCumx = "gone-cumx"

	//IdGoneGin Gin相关的组件ID，可以使用 [goner.GinPriest](goner#GinPriest) 牧师函数批量安葬
	IdGoneGin          = "gone-gin"
	IdGoneGinRouter    = "gone-gin-router"
	IdGoneGinProcessor = "gone-gin-processor"
	IdGoneGinProxy     = "gone-gin-proxy"
	IdGoneGinResponser = "gone-gin-responser"

	//IdGoneXorm Xorm Goner 的ID，封装了xorm，用于操作数据库；使用 [goner.XormPriest](goner#XormPriest) 牧师函数安葬
	IdGoneXorm = "gone-xorm"

	//IdGoneRedisPool redis pool goner; redis 相关 Goner，使用 [goner.RedisPriest](goner#RedisPriest) 牧师函数安葬
	IdGoneRedisPool   = "gone-redis-pool"
	IdGoneRedisCache  = "gone-redis-cache"
	IdGoneRedisLocker = "gone-redis-locker"

	// IdGoneSchedule 定时器Goner；使用 [goner.SchedulePriest](goner#SchedulePriest) 牧师函数安葬
	IdGoneSchedule = "gone-schedule"
)
