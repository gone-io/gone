package main

import (
	"github.com/gone-io/gone"
	log "github.com/sirupsen/logrus"
	"web-mysql/internal"
)

func main() {
	////给启动流程增加traceId
	//tracer.SetTraceId("", func() {
	//	gone.Serve(internal.MasterPriest)
	//})

	// 直接启动服务
	//gone.Serve(internal.MasterPriest)

	// 启动前注册Hook函数；更多内容参考文档：https://goner.fun/zh/guide/hooks.html
	//Hook可以注册多次；支持的Hook函数有：BeforeStart、AfterStart、BeforeStop、AfterStop;
	// 注意：BeforeStart、BeforeStop **后注册先执行**; AfterStart、AfterStop **先注册先执行**;
	gone.
		Prepare(internal.MasterPriest).
		BeforeStart(func(in struct {
			// 在 BeforeStart、AfterStart、BeforeStop、AfterStop Hook中，可以注入任何需要的依赖；
			cemetery gone.Cemetery `gone:"gone-cemetery"`
			log      gone.Logger   `gone:"gone-logger"`
		}) {
			log.Info("before start")
			// 启动前执行的代码
		}).
		AfterStart(func() {
			// 启动后执行的代码
		}).
		BeforeStop(func() {
			// 停止前执行的代码
		}).
		AfterStop(func() {
			// 停止后执行的代码
		}).
		Serve()

}
