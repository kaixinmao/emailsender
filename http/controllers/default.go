package controllers

import (
	"github.com/astaxie/beego"
)

type JsonApiRep struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type baseController struct {
	beego.Controller
}

func (this *baseController) ApiSucc(data interface{}) {
	this.ApiReturn(0, "", data)
}

func (this *baseController) ApiUnknowFailed(msg string) {
	this.ApiReturn(500, msg, nil)
}

func (this *baseController) ApiReturn(code int, msg string, data interface{}) {
	jar := &JsonApiRep{code, msg, data}
	this.Data["json"] = jar
	this.ServeJSON()
}

type DefaultController struct {
	baseController
}

func (this *DefaultController) Get() {
	this.Ctx.WriteString("hello heiheihei!")
}
