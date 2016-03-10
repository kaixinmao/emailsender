package main

import (
	"github.com/kaixinmao/emailsender/email"
	"github.com/kaixinmao/emailsender/http"
	"github.com/kaixinmao/emailsender/model"
)

func main() {
	model.Init()
	email.EmailManager.Run()
	http.Run()
	email.EmailManager.WaitStop()
}
