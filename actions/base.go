// 满足预警条件后执行的动作
package actions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/koding/multiconfig"
)

// Actioner describes an action type. There all multiple action types, but they
// all simply attempt to perform one action and that's it
type Actioner interface {
	// Do takes in the alert context, and possibly returnes an error if the
	// action failed
	Do() error
}

// Action is a wrapper around an Actioner which contains some type information
type Action struct {
	Type string
	Actioner
}

// ToActioner takes in some arbitrary data (hopefully a map[string]interface{},
// looks at its "type" key, and any other fields necessary based on that type,
// and returns an Actioner (or an error)
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
	case "mail":
		a = &Mail{}
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

type ServerConf struct {
	Action ActionConf
}

type ActionConf struct {
	MailHost     string `default:"127.0.0.1:25"`
	MailUsername string `default:"noreply@admin.com"`
	MailPwd      string `default:""`
	WechatHost   string `default:"127.0.0.1:8082"`
}

var conf *ServerConf

func Load(loader *multiconfig.DefaultLoader) {
	conf = new(ServerConf)
	loader.MustLoad(conf)
}
