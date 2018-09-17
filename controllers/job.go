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
)

type JobController struct {
	runningJobs map[string]int
	logger      *zap.Logger
	signCh      chan string
}

func (ctrl *JobController) Init() {
	ctrl.logger = alert.GetLogger()
	ctrl.runningJobs = make(map[string]int)
	ctrl.signCh = make(chan string, 10)
}

func (ctrl JobController) Recover() {
	jobs, err := models.GetJobs()
	if err != nil {
		ctrl.logger.Fatal("failed to access db", zap.String("err", err.Error()))
	}

	ctrl.logger.Info("recovering alerts",
		zap.Int("count", len(jobs)),
	)

	for _, job := range jobs {
		var a alert.Alert
		if err := yaml.Unmarshal([]byte(job.Value), &a); err != nil {
			ctrl.logger.Error("failed to parse yaml",
				zap.Int64("id", job.Id),
				zap.String("err", err.Error()),
				zap.String("value", job.Value),
			)
		} else {
			ctrl.initJob(a)
			ctrl.runningJobs[a.Name] = 1
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
		ctrl.logger.Error("failed to parse yaml",
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
		if _, isRunning := ctrl.runningJobs[a.Name]; isRunning {
			go ctrl.reloadJob(a)
		} else {
			ctrl.initJob(a)
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
	if _, isRunning := ctrl.runningJobs[jobName]; !isRunning {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "job not runningJobs",
		})
		return
	}

	ctrl.stopJob(strconv.FormatInt(job.Id, 10), false)
	c.JSON(http.StatusOK, gin.H{
		"msg": "stop ok",
	})
	return
}

func (ctrl JobController) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"list": ctrl.runningJobs,
	})
	return
}

func (ctrl JobController) initJob(a alert.Alert) {
	if err := a.Init(); err != nil {
		ctrl.logger.Error("failed to initialize alert",
			zap.String("id", a.Name),
			zap.String("err", err.Error()),
		)
	} else {
		go ctrl.jobSpin(a)
		ctrl.logger.Info("initialized alert",
			zap.String("id", a.Name),
		)
	}
}

func (ctrl JobController) reloadJob(a alert.Alert) {
	ctrl.stopJob(a.Name, true)

	if err := a.Init(); err != nil {
		ctrl.logger.Error("failed to initialize alert",
			zap.String("id", a.Name),
			zap.String("err", err.Error()),
		)
	} else {
		go ctrl.jobSpin(a)
		ctrl.logger.Info("reloaded alert",
			zap.String("id", a.Name),
		)
	}
}

func (ctrl JobController) stopJob(name string, isReload bool) {
	ctrl.logger.Info("stopping alert",
		zap.String("id", name),
	)

	ctrl.signCh <- name

	if isReload {
		// 需要延迟启动，否则新的goroutine会抢先收到退出信号
		time.Sleep(time.Second)
	}
}

func (ctrl JobController) jobSpin(a alert.Alert) {
	for {
		select {
		case sig := <-ctrl.signCh:
			if sig == a.Name {
				ctrl.logger.Info("stop alert spin", zap.String("id", a.Name))
				return
			}
		default:
			now := time.Now()
			next := a.Timer.Next(now)
			if now == next {
				ctrl.logger.Info("start alert spin", zap.String("id", a.Name))
				go a.Run()
			}
			time.Sleep(time.Second)
		}
	}
}
