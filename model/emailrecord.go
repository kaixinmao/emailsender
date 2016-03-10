package model

import (
	"encoding/json"
	"time"
)

const (
	EMAIL_RECORD_STATUS_SUCC   = 10
	EMAIL_RECORD_STATUS_WAIT   = 0
	EMAIL_RECORD_STATUS_FAILED = -10
)

//参数Json传递过来的邮件地址格式
type EmailAddress struct {
	Addr string `json:"addr"`
	Name string `json:"name"`
}

type EmailRecord struct {
	Id         int64 `xorm:"pk autoincr"`
	AppId      string
	Type       string
	From       string
	Subject    string
	Content    string
	Priority   int
	Status     int
	ErrorInfo  string
	SendTime   time.Time
	CreateTime time.Time `xorm:"created"`
}

type EmailRecordTo struct {
	Id            int64 `xorm:"pk autoincr"`
	EmailRecordId int64
	To            string
	ToName        string
	CreateTime    time.Time `xorm:"created"`
}

type EmailRecordList []EmailRecord
type EmailRecordToList []EmailRecordTo
type EmailRecordToListMap map[int64]EmailRecordToList

func (e *EmailRecord) GetFromAddr() (EmailAddress, error) {
	addr := EmailAddress{}
	err := json.Unmarshal([]byte(e.From), &addr)
	if err != nil {
		return addr, err
	}

	return addr, nil
}

func (e *EmailRecord) GetToList() (EmailRecordToList, error) {
	list := EmailRecordList{*e}

	listMap, err := list.GetToListMap()
	if err != nil {
		return nil, err
	}

	if v, ok := listMap[e.Id]; ok {
		return v, nil
	} else {
		return make(EmailRecordToList, 0), nil
	}
}

func (erl EmailRecordList) GetToListMap() (EmailRecordToListMap, error) {
	toMap := make(EmailRecordToListMap)

	if len(erl) == 0 {
		return toMap, nil
	}

	toList := make(EmailRecordToList, 0)

	ids := make([]int64, 0)
	for _, er := range erl {
		ids = append(ids, er.Id)
	}

	err := Orm.In("`email_record_id`", ids).Find(&toList)
	if err != nil {
		return nil, err
	}

	for _, to := range toList {
		if l, ok := toMap[to.EmailRecordId]; ok {
			toMap[to.EmailRecordId] = append(l, to)
		} else {
			toMap[to.EmailRecordId] = EmailRecordToList{to}
		}
	}

	return toMap, nil

}
