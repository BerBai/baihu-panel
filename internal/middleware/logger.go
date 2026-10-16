package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/engigu/baihu-panel/internal/logger"
	"github.com/engigu/baihu-panel/internal/utils"

	"github.com/gin-gonic/gin"
)

// GinLogger 返回使用 zap 的 Gin 日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		// 格式化延迟时间
		var latencyStr string
		if latency < time.Millisecond {
			latencyStr = fmt.Sprintf("%dµs", latency.Microseconds())
		} else if latency < time.Second {
			latencyStr = fmt.Sprintf("%dms", latency.Milliseconds())
		} else {
			latencyStr = fmt.Sprintf("%.2fs", latency.Seconds())
		}

		msg := fmt.Sprintf("[HTTP] %d %s %s %s [%s]", status, method, path, latencyStr, clientIP)

		if status >= 500 {
			errDetails := ""
			if len(c.Errors) > 0 {
				errDetails = fmt.Sprintf(" | 错误原因: %s", c.Errors.String())
			}
			callerDetails := ""
			if callerVal, exists := c.Get("error_caller"); exists {
				if callerStr, ok := callerVal.(string); ok && callerStr != "" {
					callerDetails = fmt.Sprintf(" | 触发位置: %s", callerStr)
				}
			}
			logger.Errorf("%s%s%s", msg, errDetails, callerDetails)

			// 若发生致命 Panic，则输出完整的系统堆栈追踪
			if stackVal, exists := c.Get("error_stack"); exists {
				if stackStr, ok := stackVal.(string); ok && stackStr != "" {
					logger.Errorf("[HTTP] Panic 异常堆栈追踪 (%s):\n%s", path, stackStr)
				}
			}
		} else if status >= 400 {
			if len(c.Errors) > 0 {
				logger.Warnf("%s | 警告: %s", msg, c.Errors.String())
			} else {
				logger.Warn(msg)
			}
		} else {
			logger.Info(msg)
		}
	}
}

// GinRecovery 返回使用 zap 的 Gin 恢复中间件
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				logger.Errorf("[HTTP] Panic 致命异常恢复: %v | 路径: %s %s\n[HTTP] 异常调用栈:\n%s",
					err, c.Request.Method, c.Request.URL.RequestURI(), stack)
				c.Set("error_stack", stack)
				_ = c.Error(fmt.Errorf("panic: %v", err))
				if !c.Writer.Written() {
					// 返回真实 500 + JSON 信封，避免空 body 导致前端 res.json() 抛异常
					utils.ServerError(c, fmt.Sprintf("服务器内部错误: %v", err))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
