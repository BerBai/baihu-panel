package mcp

import (
	"net/http"

	"github.com/engigu/baihu-panel/internal/constant"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterOpenAPIMCPRoutes 在 Gin 路由组上注册 MCP 服务端点（现行标准 Streamable HTTP 单端点协议）
func RegisterOpenAPIMCPRoutes(g *gin.RouterGroup, deps *Deps) {
	mcpServer := NewBaihuMCPServer(deps)

	// Streamable HTTP (2025-03-26 现行标准规范，单端点支持)
	streamableServer := server.NewStreamableHTTPServer(
		mcpServer,
		// 允许在反向代理（如 Nginx、Docker 网络环境或使用自定义域名）下访问，避免本地回环 DNS 重绑定防护误杀
		server.WithDisableLocalhostProtection(true),
	)

	mcpGroup := g.Group("/mcp")
	mcpGroup.Use(demoModeLimiter())
	{
		// 现行主流标准：单端点 Streamable HTTP (支持 POST/GET/DELETE 等所有协议交互)
		mcpHandler := gin.WrapH(streamableServer)
		mcpGroup.Any("", mcpHandler)
		mcpGroup.Any("/", mcpHandler)
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
