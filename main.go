package main

import (
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/koding/multiconfig"
	"go.uber.org/zap"

	"./models"
	"./alert"
	"./controllers"
)

var logger *zap.Logger

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	logger, _ = zap.NewProduction()
	r.Use(logHandler(logger))

	// load config
	//path := os.Getenv("CONF_PATH")
	conf := multiconfig.NewWithPath("config/dev.toml")

	alert.InitLogger(logger)
	jobCtrl := new(controllers.JobController)
	jobCtrl.Init()

	// recover alerts
	err := models.InitDB(conf)
	if err != nil {
		logger.Fatal("initializing db failed", zap.String("err", err.Error()))
	}
	jobCtrl.Recover()

	// rest-api trigger
	watcher := r.Group("/watcher")
	{

		// POST watcher/:id
		watcher.POST("/:id", jobCtrl.Reload)

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

func logHandler(logger *zap.Logger) gin.HandlerFunc {
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
