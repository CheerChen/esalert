package main

import (
	"os"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/koding/multiconfig"
	"go.uber.org/zap"

	"github.com/CheerChen/esalert/controllers"
	"github.com/CheerChen/esalert/models"
	"github.com/CheerChen/esalert/logger"
	"github.com/CheerChen/esalert/actions"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(logHandler())

	// 引入配置
	path := os.Getenv("CONF_PATH")
	conf := multiconfig.NewWithPath(path)

	err := models.InitDB(conf)
	if err != nil {
		logger.Fatal("initializing db failed", zap.String("err", err.Error()))
	}
	actions.Load(conf)

	// 恢复现场：从 MYSQL 获取有效配置
	jobCtrl := new(controllers.JobController)
	jobCtrl.Recover()

	// 同步配置：通过 REST-API 启动停止
	watcher := r.Group("/watcher")
	{

		// POST watcher/:id
		watcher.POST("/:id", jobCtrl.Trigger)

		// DELETE watcher/:id
		watcher.DELETE("/:id", jobCtrl.Stop)

		// GET watcher list
		watcher.GET("/", jobCtrl.List)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, string("Service Available"))
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, string("Not Found"))
	})

	r.Run(":9000")

}

func logHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			//zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			//zap.String("time", end.Format(timeFormat)),
			zap.Duration("latency-ms", latency*1000),
			zap.String("error", c.Errors.String()),
		)
	}
}
