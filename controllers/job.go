package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/CheerChen/esalert/alert"
	"github.com/CheerChen/esalert/models"
	"github.com/CheerChen/esalert/logger"
)

type JobController struct{}

var chMap map[string]chan int

func init() {
	chMap = make(map[string]chan int)
}

func (ctrl JobController) Recover() {
	jobs, err := models.GetJobs()
	if err != nil {
		logger.Fatal("failed to access db", zap.String("err", err.Error()))
	}

	logger.Info("recovering alerts",
		zap.Int("count", len(jobs)),
	)

	for _, job := range jobs {
		var a alert.Alert
		if err := yaml.Unmarshal([]byte(job.Value), &a); err != nil {
			logger.Error("failed to parse yaml",
				zap.Int64("id", job.Id),
				zap.String("err", err.Error()),
				zap.String("value", job.Value),
			)
		} else {
			ctrl.initJob(a)
		}
	}
}

func (ctrl JobController) Reload(c *gin.Context) {
	id := c.Param("id")
	job, err := models.GetJobById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"msg":   "job id not found",
			"error": err.Error(),
		})
		return
	}

	// reload/start job
	var a alert.Alert
	if err := yaml.Unmarshal([]byte(job.Value), &a); err != nil {
		logger.Error("failed to parse yaml",
			zap.Int64("id", job.Id),
			zap.String("err", err.Error()),
			zap.String("value", job.Value),
		)
		c.JSON(http.StatusNotAcceptable, gin.H{
			"msg":   "failed to parse yaml",
			"error": err.Error(),
		})
		return
	} else {
		if _, ok := chMap[a.Name]; !ok {
			ctrl.initJob(a)
		} else {
			go ctrl.reloadJob(a)
		}
		c.JSON(http.StatusOK, gin.H{
			"msg": "reload ok",
		})
		return
	}

}

func (ctrl JobController) Stop(c *gin.Context) {
	id := c.Param("id")
	job, err := models.GetJobById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"msg":   "job id not found",
			"error": err.Error(),
		})
		return
	}
	// stop job
	if job.Status == 1 && job.IsDeleted == 0 {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"msg": "job still available",
		})
		return
	}
	jobName := strconv.FormatInt(job.Id, 10)
	if _, ok := chMap[jobName]; !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "job no running",
		})
		return
	}

	go ctrl.stopJob(strconv.FormatInt(job.Id, 10))
	c.JSON(http.StatusOK, gin.H{
		"msg": "stop ok",
	})
	return
}

func (ctrl JobController) List(c *gin.Context) {
	keys := make([]string, 0, len(chMap))
	for k := range chMap {
		keys = append(keys, k)
	}
	c.JSON(http.StatusOK, gin.H{
		"list": keys,
	})
	return
}

func (ctrl JobController) initJob(a alert.Alert) {
	if err := a.Init(); err != nil {
		logger.Error("failed to initialize alert",
			zap.String("id", a.Name),
			zap.String("err", err.Error()),
		)
	} else {
		chMap[a.Name] = make(chan int)
		go ctrl.jobSpin(a)
		logger.Info("initialized alert",
			zap.String("id", a.Name),
		)
	}
}

func (ctrl JobController) reloadJob(a alert.Alert) {
	ctrl.stopJob(a.Name)

	if err := a.Init(); err != nil {
		logger.Error("failed to initialize alert",
			zap.String("id", a.Name),
			zap.String("err", err.Error()),
		)
	} else {
		chMap[a.Name] = make(chan int)
		go ctrl.jobSpin(a)
		logger.Info("reloaded alert",
			zap.String("id", a.Name),
		)
	}
}

func (ctrl JobController) stopJob(name string) {
	logger.Info("stopping alert",
		zap.String("id", name),
	)

	close(chMap[name])
	time.Sleep(time.Second)
	delete(chMap, name)
	logger.Info("delete alert channel map", zap.String("id", name))
}

func (ctrl JobController) jobSpin(a alert.Alert) {
	for {
		select {
		case <-chMap[a.Name]:
			logger.Info("stop alert spin", zap.String("id", a.Name))
			return
		default:
			now := time.Now()
			next := a.Timer.Next(now)
			if now == next {
				logger.Info("start alert spin", zap.String("id", a.Name))
				go a.Run()
			}
			time.Sleep(time.Second)
		}
	}
}
