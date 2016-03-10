package util

import (
	"errors"

	"github.com/kaixinmao/emailsender/model"

	"github.com/astaxie/beego/validation"
)

func CheckEmailAddress(e *model.EmailAddress) error {

	if e.Addr == "" {
		return errors.New("need addr")
	}

	//使用beego方法来进行验证
	valid := validation.Validation{}

	valid.Email(e.Addr, "addr")
	valid.MaxSize(e.Addr, 100, "addr")
	valid.MaxSize(e.Name, 50, "name")

	if valid.HasErrors() {
		validErr := valid.Errors[0]
		return errors.New(validErr.Message)
	}

	return nil
}
