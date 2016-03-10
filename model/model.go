package model

import (
	"github.com/kaixinmao/emailsender/setting"

	"errors"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var Orm *xorm.Engine
var ErrNotExist error

func Init() {
	//初始化变量
	ErrNotExist = errors.New("not exist")

	//初始化引擎
	//获取配置信息
	mysqlJson, has := setting.Cfg.CheckGet("mysql")
	if !has {
		panic("no mysql config")
	}

	host := mysqlJson.Get("host").MustString()
	port := mysqlJson.Get("port").MustString("3306")
	username := mysqlJson.Get("username").MustString()
	password := mysqlJson.Get("password").MustString()
	database := mysqlJson.Get("database").MustString()

	if host == "" || username == "" || database == "" {
		panic("please set host,username,database config")
	}

	charset := "utf8"

	//dsn格式 "root:123@/test?charset=utf8"
	dsnStr := username
	if password != "" {
		dsnStr = dsnStr + ":" + password
	}

	dsnStr = dsnStr + "@tcp(" + host + ":" + port + ")/" + database + "?charset=" + charset
	var err error
	Orm, err = xorm.NewEngine("mysql", dsnStr)
	if err != nil {
		panic("init mysql engine error")
	}

	Orm.ShowSQL(true)
}
