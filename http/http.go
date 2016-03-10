package http

import (
	_ "github.com/kaixinmao/emailsender/setting"

	"github.com/astaxie/beego"
)

// Http模块，主要完成对外Api的暴露
// 该模块使用了 Beego GraceFul 的热启动方式。
// 极限情况下会出现老进程Http模块已经关闭，新进程开始进入服务，但老进程的非Http服务还在运行的情况

//开启http服务，进入监听状态
func Run() {
	beego.Run()
}

func init() {
	//指定http服务为单独的ini配置文件
	beego.AppConfigPath = "config/http.ini"
}
