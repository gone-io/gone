package main

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/schedule"
)

func priest(cemetery gone.Cemetery) error {

	//使用 goner.SchedulePriest 函数，将 定时任务 相关的Goner 埋葬到 Cemetery 中
	_ = goner.SchedulePriest(cemetery)

	//1.将配置文件支持的相关Goner 埋葬到 Cemetery 中
	_ = config.Priest(cemetery)

	cemetery.Bury(&sch{})
	return nil
}

type sch struct {
	gone.Flag

	cron string `gone:"config,cron.job1,default=*/5 * * * * *"` //2. 注入放到配置文件的定时任务配置
}

func (sch *sch) job1() {
	//todo 定时任务逻辑
	fmt.Println("job1 execute")
}

func (sch *sch) Cron(run schedule.RunFuncOnceAt) {

	//使用 run `RunFuncOnceAt`设置定时任务，
	run(
		sch.cron, // 3. 使用从配置文件注入的定时配置
		"job1",   //需要设置一个唯一标识，用于 分布式锁加锁
		sch.job1, // 定时任务逻辑
	)
}

func main() {
	gone.Serve(priest)
}
