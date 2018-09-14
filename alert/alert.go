package alert

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"go.uber.org/zap"
)

type Alert struct {
	Name      string    `yaml:"name"`
	Interval  string    `yaml:"interval"`
	Search    Dict      `yaml:"search"`
	SearchUrl string    `yaml:"search_url"`
	Process   LuaRunner `yaml:"process"`

	Timer     *FullTimeSpec
	SearchTPL *template.Template
}

func templateHelper(i interface{}, lastErr error) (*template.Template, error) {
	if lastErr != nil {
		return nil, lastErr
	}
	var str string
	if s, ok := i.(string); ok {
		str = s
	} else {
		b, err := yaml.Marshal(i)
		if err != nil {
			return nil, err
		}
		str = string(b)
	}

	return template.New("").Parse(str)
}

func (a *Alert) Init() error {
	var err error
	a.SearchTPL, err = templateHelper(&a.Search, err)
	if err != nil {
		return err
	}

	timer, err := ParseFullTimeSpec(a.Interval)
	if err != nil {
		return fmt.Errorf("parsing interval: %s", err)
	}
	a.Timer = timer

	return nil
}

func (a Alert) Run() {
	now := time.Now()
	c := Context{
		Name:      a.Name,
		StartedTS: uint64(now.Unix()),
		Time:      now,
	}

	searchQuery, err := a.CreateSearchQuery(c)
	if err != nil {
		logger.Error("failed to create search query",
			zap.String("err", err.Error()),
			zap.String("id", a.Name),
		)
		return
	}

	logger.Info("running search step")

	res, err := Search(a.SearchUrl, searchQuery)
	if err != nil {
		logger.Error("failed at search step",
			zap.String("err", err.Error()),
			zap.String("id", a.Name),
		)
		return
	}
	c.Result = res

	logger.Info("running process step",
		zap.Uint64("hits", res.HitInfo.HitCount),
		zap.String("id", a.Name),
	)

	processRes, ok := a.Process.Do(c)
	if !ok {
		logger.Error("failed at process step",
			zap.String("err", err.Error()),
			zap.String("id", a.Name),
		)
		return
	}

	actionsRaw, _ := processRes.([]interface{})
	if len(actionsRaw) == 0 {
		logger.Error("no actions returned",
			zap.String("id", a.Name),
		)
	}

	actions := make([]Action, len(actionsRaw))
	for i := range actionsRaw {
		act, err := ToActioner(actionsRaw[i])
		if err != nil {
			logger.Error("error unpacking action",
				zap.String("id", a.Name),
			)
			return
		}
		actions[i] = act
	}

	for i := range actions {
		logger.Info("running action step")
		if err := actions[i].Do(c); err != nil {
			logger.Error("failed to complete action",
				zap.String("err", err.Error()),
				zap.String("id", a.Name),
			)
			return
		}
	}
}

func (a Alert) CreateSearchQuery(c Context) (interface{}, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	if err := a.SearchTPL.Execute(buf, &c); err != nil {
		return nil, err
	}
	searchRaw := buf.Bytes()
	logger.Info("create search query",
		zap.ByteString("searchRaw", searchRaw),
		zap.String("id", a.Name),
	)

	var search Dict
	if err := yaml.Unmarshal(searchRaw, &search); err != nil {
		return nil, err
	}

	return search, nil
}
