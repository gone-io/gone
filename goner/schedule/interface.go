package schedule

type JobName string

// RunFuncOnceAt 定时跑
// @Param fn 要调用的函数
// @Param spec 调用时间 cron tab 格式
// ┌───────────── second (0 - 59)
// │ ┌───────────── minute (0 - 59)
// │ │ ┌───────────── hour (0 - 23)
// │ │ │ ┌───────────── day of the month (1 - 31)
// │ │ │ │ ┌───────────── month (1 - 12)
// │ │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday;
// │ │ │ │ │ │                                   7 is also Sunday on some systems)
// │ │ │ │ │ │
// │ │ │ │ │ │
// * * * * * *
// @Param lockKey 分布式锁的key
// @Param lockTtl 锁定时长
type RunFuncOnceAt func(spec string, jobName JobName, fn func())

type Scheduler interface {

	//Cron use: Cron(run facility.RunFuncOnceAt)
	Cron(run RunFuncOnceAt)
}

type Schedule interface {
	Start() error
	Stop() error
}
