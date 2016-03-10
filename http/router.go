package http

import (
	"github.com/kaixinmao/emailsender/http/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/test", &controllers.DefaultController{})
	beego.Router("/emails", &controllers.EmailsController{})
}
