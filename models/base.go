package models

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/koding/multiconfig"
)

var db *sqlx.DB

type ServerConf struct {
	DBDriver  string `default:"mysql"`
	DBMaxIdle int    `default:"200"`
	DBMaxOpen int    `default:"200"`

	Mysql MysqlConf
}

type MysqlConf struct {
	Name     string `default:"root"`
	Pwd      string `default:""`
	Host     string `default:"127.0.0.1"`
	Port     string `default:"3306"`
	Database string `default:""`
}

func InitDB(loader *multiconfig.DefaultLoader) (err error) {
	conf := new(ServerConf)
	loader.MustLoad(conf)
	var dsn string
	switch conf.DBDriver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8",
			conf.Mysql.Name,
			conf.Mysql.Pwd,
			conf.Mysql.Host,
			conf.Mysql.Port,
			conf.Mysql.Database,
		)
		break
	}

	//log.Println(dsn)

	db, err = sqlx.Connect(conf.DBDriver, dsn)
	if err != nil {
		return err
	}
	db.SetMaxIdleConns(conf.DBMaxIdle)
	db.SetMaxOpenConns(conf.DBMaxOpen)
	return nil
}
