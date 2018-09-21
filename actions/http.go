package actions

import (
	"net/http"
	"bytes"
	"fmt"
)

// HTTP请求动作
type HTTP struct {
	Method  string            `mapstructure:"method"`
	URL     string            `mapstructure:"url"`
	Headers map[string]string `mapstructure:"headers"`
	Body    string            `mapstructure:"body"`
}

// 执行HTTP请求
func (h *HTTP) Do() error {
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
