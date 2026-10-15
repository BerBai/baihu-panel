package mcp

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/engigu/baihu-panel/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// MaxActiveSSESessions 限制同时保持的 SSE 客户端长连接会话数上限（每个连接占用 1~2 个协程）
	MaxActiveSSESessions = 3
)

var (
	activeSSESessions int32
)

// GetActiveSSESessions 获取当前活跃的 SSE 连接会话数
func GetActiveSSESessions() int32 {
	return atomic.LoadInt32(&activeSSESessions)
}

// RegisterOpenAPIMCPRoutes 在 Gin 路由组上注册 MCP 服务端点
// 同时支持：
// 1. 现行主流标准 Streamable HTTP 单端点协议 (POST/GET/DELETE /open2api/v1/mcp)
// 2. 传统 HTTP + SSE 双端点协议 (GET /open2api/v1/mcp/sse + POST /open2api/v1/mcp/messages)
func RegisterOpenAPIMCPRoutes(g *gin.RouterGroup, deps *Deps) {
	mcpServer := NewBaihuMCPServer(deps)

	// 1. Streamable HTTP (2025-03-26 现行主流标准规范，单端点通信)
	streamableServer := server.NewStreamableHTTPServer(
		mcpServer,
		// 允许在反向代理（如 Nginx、Docker 网络环境或使用自定义域名）下访问，避免本地回环 DNS 重绑定防护误杀
		server.WithDisableLocalhostProtection(true),
	)

	// 2. HTTP + SSE (2024-11-05 早期双端点规范，兼容旧客户端)
	cfg := services.GetConfig()
	prefix := ""
	if cfg != nil && cfg.Server.URLPrefix != "" {
		prefix = strings.TrimSuffix(cfg.Server.URLPrefix, "/")
	}

	sseEndpoint := prefix + "/open2api/v1/mcp/sse"
	msgEndpoint := prefix + "/open2api/v1/mcp/messages"

	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithSSEEndpoint(sseEndpoint),
		server.WithMessageEndpoint(msgEndpoint),
		server.WithAppendQueryToMessageEndpoint(),
	)

	mcpGroup := g.Group("/mcp")
	mcpGroup.Use(demoModeLimiter())
	{
		// SSE 兼容端点 (带活跃连接数限制)
		mcpGroup.GET("/sse", sseSessionLimiter(), gin.WrapH(sseServer.SSEHandler()))
		mcpGroup.POST("/messages", gin.WrapH(sseServer.MessageHandler()))

		// Streamable HTTP 单端点 (处理 /mcp 及 /mcp/ 的全部请求)
		streamableHandler := gin.WrapH(streamableServer)
		mcpGroup.Any("", streamableHandler)
		mcpGroup.Any("/", streamableHandler)
	}
}

// demoModeLimiter 演示模式下禁止使用 MCP 协议及接口能力
func demoModeLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if constant.DemoMode {
			c.JSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"msg":  "演示模式下已禁用 MCP 协议与远程连接能力",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// sseSessionLimiter 限制同时保持的 SSE 挂起连接协程数
func sseSessionLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		current := atomic.AddInt32(&activeSSESessions, 1)
		if current > MaxActiveSSESessions {
			atomic.AddInt32(&activeSSESessions, -1)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  fmt.Sprintf("当前活跃的 MCP SSE 连接数已达上限 (%d)，已拒绝新连接以保护系统协程资源", MaxActiveSSESessions),
			})
			c.Abort()
			return
		}

		defer func() {
			atomic.AddInt32(&activeSSESessions, -1)
		}()

		c.Next()
	}
}
