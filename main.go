// main.go
package main

import (
	adminRouter "eli/app/admin/router"
	frontendRouter "eli/app/frontend/router"
	"eli/config"
	"fmt"

	"github.com/gin-gonic/gin"
)

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o eli-serv main.go
func main() {
	r := gin.Default()

	// gin.SetMode(gin.ReleaseMode)

	// 加载前台用户模块路由
	frontendRouter.Load(r.Group("/api/v1"))

	// 加载后台用户模块路由
	adminRouter.Load(r.Group("/api/admin/v1"))

	// 启动 Gin 服务
	r.Run(fmt.Sprintf(":%s", config.Get().Http.Port))
}
