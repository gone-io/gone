package gin

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/internal/json"
	"io"
	"strings"
	"time"
)

// NewSysMiddleware create a new sys middleware
func NewSysMiddleware() (gone.Goner, gone.GonerOption, gone.GonerOption) {
	return &SysMiddleware{}, gone.Order0, gone.IdGoneGinSysMiddleware
}

// SysMiddleware system middleware
type SysMiddleware struct {
	gone.Flag

	logger     gone.Logger `gone:"*"`
	tracer     gone.Tracer `gone:"*"`
	router     IRouter     `gone:"*"`
	resHandler Responser   `gone:"*"`

	// healthCheckUrl 健康检查路劲
	// 对应配置项为: `server.health-check`
	// 默认为空，不开启；
	// 配置后，能够在该路劲提供一个http-status等于200的空响应
	healthCheckUrl string `gone:"config,server.health-check"`

	logFormat string `gone:"server.log.format,default=console"`

	// showRequestTime 展示请求时间
	// 对应配置项为：`server.log.show-request-time`
	// 默认为`true`;
	// 开启后，日志中将使用`Info`级别打印请求的 耗时
	showRequestTime bool `gone:"config,server.log.show-request-time,default=true"`

	showRequestLog   bool `gone:"config,server.log.show-request-log,default=true"`
	logDataMaxLength int  `gone:"config,server.log.data-max-length,default=0"`
	logRequestId     bool `gone:"config,server.log.request-id,default=true"`
	logRemoteIp      bool `gone:"config,server.log.remote-ip,default=true"`
	logRequestBody   bool `gone:"config,server.log.request-body,default=true"`
	logUserAgent     bool `gone:"config,server.log.user-agent,default=true"`
	logReferer       bool `gone:"config,server.log.referer,default=true"`

	requestBodyLogContentTypes string `gone:"config,server.log.show-request-body-for-content-types,default=application/json,application/xml,application/x-www-form-urlencoded"`

	showResponseLog bool `gone:"config,server.log.show-response-log,default=true"`

	responseBodyLogContentTypes string `gone:"config,server.log.show-response-body-for-content-types,default=application/json,application/xml,application/x-www-form-urlencoded"`

	useTracer bool `gone:"config,server.use-tracer,default=true"`

	isAfterProxy bool `gone:"config,server.is-after-proxy,default=false"`
}

func (m *SysMiddleware) Process(context *gin.Context) {
	if m.healthCheckUrl != "" && context.Request.URL.Path == m.healthCheckUrl {
		context.AbortWithStatus(200)
		return
	}

	traceId := context.GetHeader(gone.TraceIdHeaderKey)
	if m.useTracer {
		m.tracer.SetTraceId(traceId, func() {
			m.process(context)
		})
	} else {
		m.process(context)
	}
}

func (m *SysMiddleware) process(context *gin.Context) {
	defer m.stat(context, time.Now())
	defer m.recover(context)

	if m.showRequestLog {
		logMap := make(map[string]any)

		if m.logRequestId {
			requestID := context.GetHeader(gone.RequestIdHeaderKey)
			logMap["request-id"] = requestID
		}

		if m.logRemoteIp {
			var remoteIP string
			if m.isAfterProxy {
				remoteIP = context.GetHeader("X-Forwarded-For")
			} else {
				remoteIP = context.RemoteIP()
			}
			logMap["remote-ip"] = remoteIP
		}

		logMap["method"] = context.Request.Method
		logMap["path"] = context.Request.URL.Path

		if m.logUserAgent {
			logMap["user-agent"] = context.Request.UserAgent()
		}

		if m.logReferer {
			logMap["referer"] = context.Request.Referer()
		}

		if m.logRequestBody && strings.Contains(m.requestBodyLogContentTypes, context.ContentType()) {
			data, err := cloneRequestBody(context)
			if err != nil {
				m.logger.Error("accessLog - cloneRequestBody error:", err)
			}

			if m.logDataMaxLength > 0 && len(data) > m.logDataMaxLength {
				buf := make([]byte, 0, m.logDataMaxLength+3)
				buf = append(buf, data[0:m.logDataMaxLength]...)
				buf = append(buf, []byte("...")...)
				data = buf
			}
			logMap["body"] = string(data)
		}

		m.log("request", logMap)
	}

	if m.showResponseLog {
		crw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: context.Writer}
		context.Writer = crw

		context.Next()

		logMap := make(map[string]any)
		logMap["method"] = context.Request.Method
		logMap["path"] = context.Request.URL.Path
		logMap["status"] = crw.Status()

		contentType := context.Writer.Header().Get("Content-Type")
		logMap["content-type"] = contentType

		if strings.Contains(m.responseBodyLogContentTypes, contentType) {
			data := crw.body.String()
			if m.logDataMaxLength > 0 && len(data) > m.logDataMaxLength {
				buf := make([]byte, 0, m.logDataMaxLength+3)
				buf = append(buf, data[0:m.logDataMaxLength]...)
				buf = append(buf, []byte("...")...)
				data = string(buf)
			}
			logMap["body"] = data
		}
		m.log("response", logMap)
	} else {
		context.Next()
	}
}

func (m *SysMiddleware) recover(context *gin.Context) {
	if r := recover(); r != nil {
		m.logger.Errorf("request(%s %s) panic: %v, %s",
			context.Request.Method,
			context.Request.URL.Path,
			r,
			gone.PanicTrace(2, 1),
		)

		err := gone.ToError(r)
		m.resHandler.Failed(context, err)
		context.Abort()
	}
}

func (m *SysMiddleware) stat(c *gin.Context, begin time.Time) {
	if m.showRequestTime {
		m.log("request-use-time", map[string]any{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"use-time": time.Since(begin),
		})
	}
}

func (m *SysMiddleware) log(t string, info map[string]any) {
	switch m.logFormat {
	case "json":
		info["type"] = t
		jsonLog, _ := json.Marshal(info)
		m.logger.Infof("%s", jsonLog)
	default:
		arr := make([]string, 0, len(info))
		for k, v := range info {
			arr = append(arr, fmt.Sprintf("%s=%v", k, v))
		}
		m.logger.Infof("[%s] %s", t, strings.Join(arr, "|"))
	}
}

//-------------------------------

func cloneRequestBody(c *gin.Context) ([]byte, error) {
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
