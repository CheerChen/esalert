package alert

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"encoding/json"
	"strings"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Actioner interface {
	Do(Context) error
}

type Action struct {
	Type string
	Actioner
}

func ToActioner(in interface{}) (Action, error) {
	min, ok := in.(map[string]interface{})
	if !ok {
		return Action{}, errors.New("action definition is not an object")
	}

	var a Actioner
	typ, _ := min["type"].(string)
	typ = strings.ToLower(typ)
	switch typ {
	case "log":
		a = &Log{}
	case "http":
		a = &HTTP{}
	case "wechat":
		a = &Wechat{}
	default:
		return Action{}, fmt.Errorf("unknown action type: %q", typ)
	}

	if err := mapstructure.Decode(min, a); err != nil {
		return Action{}, err
	}
	return Action{Type: typ, Actioner: a}, nil
}

type Log struct {
	Message string `mapstructure:"message"`
}

func (l *Log) Do(_ Context) error {
	logger.Info(l.Message)
	return nil
}

type HTTP struct {
	Method  string            `mapstructure:"method"`
	URL     string            `mapstructure:"url"`
	Headers map[string]string `mapstructure:"headers"`
	Body    string            `mapstructure:"body"`
}

func (h *HTTP) Do(_ Context) error {
	r, err := http.NewRequest(h.Method, h.URL, bytes.NewBufferString(h.Body))
	if err != nil {
		return err
	}

	if h.Headers != nil {
		for k, v := range h.Headers {
			r.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("non 2xx response code returned: %d", resp.StatusCode)
	}

	return nil
}

// wechat service
type Wechat struct {
	Users   []string `mapstructure:"users" json:"users"`
	Content string   `mapstructure:"content" json:"content"`
}

func (w *Wechat) Do(_ Context) error {
	body, err := json.Marshal(w)
	if err != nil {
		return err
	}

	logger.Info("wechat sending request", zap.ByteString("body", body))

	r, err := http.NewRequest("POST", "http://127.0.0.1:8082/broadcast", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("non 2xx response code returned: %d", resp.StatusCode)
	}

	return nil
}
