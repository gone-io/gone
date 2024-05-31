package gin

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"io"
	"strings"
	"time"
)

// NewGinProcessor 创建系统级的`Processor`，该处理器将在路由上挂载多个中间件
// `trace`，负责提供健康检查响应(需要在配置文件中设置健康检查的路径`server.health-check`) 和 为请求绑定唯一的`traceID`
// `recovery`，负责恢复请求内发生panic
// `statRequestTime`，用于在日志中打印统计的请求耗时，可以通过设置配置项(`server.log.show-request-time=false`)来关闭
// `accessLog`，用于在日志中打印请求、响应信息
func NewGinProcessor() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &sysProcessor{}, gone.IdGoneGinProcessor, gone.IsDefault(true)
}

type sysProcessor struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`

	tracer     gone.Tracer `gone:"gone-tracer"`
	router     IRouter     `gone:"gone-gin-router"`
	resHandler Responser   `gone:"gone-gin-responser"`

	// HealthCheckUrl 健康检查路劲
	// 对应配置项为: `server.health-check`
	// 默认为空，不开启；
	// 配置后，能够在该路劲提供一个http-status等于200的空响应
	HealthCheckUrl string `gone:"config,server.health-check"`

	// ShowAccessLog 展示access日志
	// 对应配置项为：`server.log.show-access-log`
	// 默认为`true`;
	// 开启后，日志中将使用`Info`级别打印请求的request和response信息
	ShowAccessLog bool `gone:"config,server.log.show-access-log,default=true"`

	// ShowRequestTime 展示请求时间
	// 对应配置项为：`server.log.show-request-time`
	// 默认为`true`;
	// 开启后，日志中将使用`Info`级别打印请求的 耗时
	ShowRequestTime bool `gone:"config,server.log.show-request-time,default=true"`

	logDataMaxLength int `gone:"config,server.log.data-max-length,default=0"`
}

func (p *sysProcessor) AfterRevive() gone.AfterReviveError {
	m := []HandlerFunc{p.trace, p.recovery}
	if p.ShowRequestTime {
		m = append(m, p.statRequestTime)
	}
	if p.ShowAccessLog {
		m = append(m, p.accessLog)
	}
	p.router.Use(m...)
	return nil
}

var RequestIdHeaderKey = "X-Request-ID"
var TraceIdHeaderKey = "X-Trace-ID"

func (p *sysProcessor) trace(context *Context) (any, error) {
	if p.HealthCheckUrl != "" && context.Request.URL.Path == p.HealthCheckUrl {
		context.AbortWithStatus(200)
		return nil, nil
	}

	traceId := context.GetHeader(TraceIdHeaderKey)
	p.tracer.SetTraceId(traceId, func() {
		requestID := context.GetHeader(RequestIdHeaderKey)
		p.Infof("bind requestId:%s", requestID)
		context.Next()
	})
	return nil, nil
}

func (p *sysProcessor) recovery(context *Context) (any, error) {
	defer p.recover(context)
	context.Next()
	return nil, nil
}

func (p *sysProcessor) recover(context *Context) {
	if r := recover(); r != nil {
		traceID := p.tracer.GetTraceId()
		p.Errorf("[%s] handle panic: %v, %s", traceID, r, gone.PanicTrace(2))
		err := gone.ToError(r)
		p.resHandler.Failed(context.Context, err)
		context.Abort()
	}
}

func (p *sysProcessor) statRequestTime(c *Context) (any, error) {
	beginTime := time.Now()
	defer func() {
		p.Infof("request(%s %s) use time: %v", c.Request.Method, c.Request.URL.Path, time.Now().Sub(beginTime))
	}()
	c.Next()
	return nil, nil
}

func (p *sysProcessor) accessLog(c *Context) (any, error) {
	remoteIP := c.GetHeader("X-Forwarded-For")
	if remoteIP == "" {
		remoteIP = c.RemoteIP()
	}

	data, err := cloneRequestBody(c)
	if err != nil {
		p.Error("accessLog - cloneRequestBody error:", err)
	}

	if p.logDataMaxLength > 0 && len(data) > p.logDataMaxLength {
		buf := make([]byte, 0)
		buf = append(buf, data[0:p.logDataMaxLength]...)
		buf = append(buf, []byte("...")...)
		data = buf
	}

	p.Infof("api-request|%s %s %s %s %s %s\n",
		remoteIP,
		c.Request.Method,
		c.Request.RequestURI,
		c.Request.UserAgent(),
		c.GetHeader("Referer"),
		data,
	)

	blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()

	contentType := c.Writer.Header().Get("Content-Type")
	if strings.Contains(contentType, "json") {
		p.Infof("api-response|%v %s\n", c.Writer.Status(), blw.body.String())
	} else {
		p.Infof("api-response|%v %s\n", c.Writer.Status(), contentType)
	}
	return nil, nil
}

func cloneRequestBody(c *Context) ([]byte, error) {
	data, err := c.GetRawData()
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	return data, nil
}

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
