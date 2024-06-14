package gone

// GonerIds for Gone framework inner Goners
const (
	// IdGoneHeaven , The GonerId of Heaven Goner, which represents the program itself, and which is injected by default when it starts.
	IdGoneHeaven GonerId = "gone-heaven"

	// IdGoneCemetery , The GonerId of Cemetery Goner, which is Dependence Injection Key Goner, and which is injected by default.
	IdGoneCemetery GonerId = "gone-cemetery"

	// IdGoneTestKit , The GonerId of TestKit Goner, which is injected by default when using gone.Test or gone.TestAt to run test code.
	IdGoneTestKit GonerId = "gone-test-kit"

	//IdConfig , The GonerId of Config Goner, which can be used for Injecting Configs from files or envs.
	IdConfig GonerId = "config"

	//IdGoneConfigure , The GonerId of Configure Goner, which is used to read configs from devices.
	IdGoneConfigure GonerId = "gone-configure"

	// IdGoneTracer ,The GonerId of Tracer
	IdGoneTracer GonerId = "gone-tracer"

	// IdGoneLogger , The GonerId of Logger
	IdGoneLogger GonerId = "gone-logger"

	// IdGoneCMux , The GonerId of CMuxServer
	IdGoneCMux GonerId = "gone-cmux"

	// IdGoneGin , IdGoneGinRouter , IdGoneGinProcessor, IdGoneGinProxy, IdGoneGinResponser, IdHttpInjector;
	// The GonerIds of Goners in goner/gin, which integrates gin framework for web request.
	IdGoneGin          GonerId = "gone-gin"
	IdGoneGinRouter    GonerId = "gone-gin-router"
	IdGoneGinProcessor GonerId = "gone-gin-processor"
	IdGoneGinProxy     GonerId = "gone-gin-proxy"
	IdGoneGinResponser GonerId = "gone-gin-responser"
	IdHttpInjector     GonerId = "http"

	// IdGoneXorm , The GonerId of XormEngine Goner, which is for xorm engine.
	IdGoneXorm GonerId = "gone-xorm"

	// IdGoneRedisPool ,IdGoneRedisCache, IdGoneRedisKey, IdGoneRedisLocker, IdGoneRedisProvider
	// The GonerIds of Goners in goner/redis, which integrates redis framework for cache and locker.
	IdGoneRedisPool     GonerId = "gone-redis-pool"
	IdGoneRedisCache    GonerId = "gone-redis-cache"
	IdGoneRedisKey      GonerId = "gone-redis-key"
	IdGoneRedisLocker   GonerId = "gone-redis-locker"
	IdGoneRedisProvider GonerId = "gone-redis-provider"

	// IdGoneSchedule , The GonerId of Schedule Goner, which is for schedule in goner/schedule.
	IdGoneSchedule GonerId = "gone-schedule"

	// IdGoneReq , The GonerId of urllib.Client Goner, which is for request in goner/urllib.
	IdGoneReq GonerId = "gone-urllib"
)
