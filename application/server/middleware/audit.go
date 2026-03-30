package middleware

import (
	"application/pkg/audit"
	"bytes"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计中间件
// 拦截所有 /api/v2/* 请求，自动记录审计日志
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过审计接口本身，避免循环
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v2/audit/") {
			c.Next()
			return
		}

		// 获取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器来捕获响应内容
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 请求处理完成后，记录审计日志（异步，避免阻塞响应）
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Audit] Recovered from panic: %v", r)
				}
			}()

			// 构建审计上下文
			auditCtx := &audit.AuditContext{
				EventType:     audit.GetEventTypeByPath(path, c.Request.Method),
				RequestPath:   path,
				RequestMethod: c.Request.Method,
				ClientIP:      c.ClientIP(),
				UserAgent:     c.Request.UserAgent(),
			}

			// 设置事件级别
			if auditCtx.EventType != "" {
				auditCtx.EventLevel = audit.GetEventLevelByType(auditCtx.EventType)
			} else {
				auditCtx.EventLevel = audit.EventLevelL1
			}

			// 设置链码函数
			auditCtx.ChaincodeFunc = audit.GetChaincodeFuncByPath(path, c.Request.Method)

			// 从请求体中提取患者ID和病历ID
			auditCtx.TargetPatientID = audit.ExtractPatientIDFromRequest(requestBody)
			auditCtx.TargetRecordID = audit.ExtractRecordIDFromRequest(requestBody)

			// 设置操作者ID（从session或header中获取，这里默认用admin）
			auditCtx.ActorID = "0feceb66ffc1"

			// 设置详情摘要
			if len(requestBody) > 0 {
				auditCtx.DetailJSON = string(requestBody)
				if len(auditCtx.DetailJSON) > 2048 {
					auditCtx.DetailJSON = auditCtx.DetailJSON[:2048] + "..."
				}
			}

			// 根据响应状态码判断是否成功
			statusCode := c.Writer.Status()
			if statusCode >= 400 {
				auditCtx.FailReason = getFailReasonByStatus(statusCode)
				auditCtx.EventType = audit.EventTypeAPIError
				auditCtx.EventLevel = audit.EventLevelL3

				// 尝试从响应体中获取错误信息
				if blw.body.Len() > 0 {
					responseBody := blw.body.String()
					if len(responseBody) > 500 {
						auditCtx.FailReason = responseBody[:500]
					} else {
						auditCtx.FailReason = responseBody
					}
				}

				if err := audit.WriteAuditEventWithFail(auditCtx); err != nil {
					log.Printf("[Audit] Failed to write audit event (fail): %v", err)
				}
			} else {
				if err := audit.WriteAuditEvent(auditCtx); err != nil {
					log.Printf("[Audit] Failed to write audit event (success): %v", err)
				}
			}
		}()
	}
}

// bodyLogWriter 自定义响应写入器，用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// getFailReasonByStatus 根据HTTP状态码获取失败原因
func getFailReasonByStatus(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "SERVER_ERROR"
	case statusCode == 403:
		return "FORBIDDEN"
	case statusCode == 401:
		return "UNAUTHORIZED"
	case statusCode == 404:
		return "NOT_FOUND"
	case statusCode == 400:
		return "BAD_REQUEST"
	default:
		return "CLIENT_ERROR"
	}
}
