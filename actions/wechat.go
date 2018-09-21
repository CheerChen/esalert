package actions

import (
	"net/http"
	"fmt"
	"strings"
	"bytes"

	"go.uber.org/zap"

	"github.com/CheerChen/esalert/logger"
)

// 企业微信动作
type Wechat struct {
	Users   []string `mapstructure:"users" json:"users"`
	Content string   `mapstructure:"content" json:"content"`
}

// 群发
func (w *Wechat) Do() error {
	body := "receiver=" + strings.Join(w.Users, ",") + "&subject=Warning&content=" + w.Content

	logger.Info("wechat sending request", zap.String("body", body))

	r, err := http.NewRequest("POST", conf.Action.WechatHost, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
