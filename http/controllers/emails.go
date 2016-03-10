package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/kaixinmao/emailsender/email"
	"github.com/kaixinmao/emailsender/http/util"
	"github.com/kaixinmao/emailsender/model"

	"github.com/astaxie/beego/validation"
)

type EmailsController struct {
	baseController
}

type EmailsPostForm struct {
	To       string `form:"to" valid:"Required"`
	From     string `form:"from" valid:"Required"`
	Subject  string `form:"subject" valid:"Required;MaxSize(200)"`
	Content  string `form:"content" valid:"Required"`
	AppId    string `form:"app_id" valid:"Required"`
	Priority int    `form:"priority" valid:"Range(0,100)"`
	Type     string `form:"type" valid:"Required"`
}

type EmailsGetForm struct {
	AppId        string `form:"app_id" valid:"Required"`
	To           string `form:"to"`
	Page         int    `form:"page"`
	PerPage      int    `form:"per_page" valid:"Max(50)"`
	StartCreated string `form:"start_created"`
	EndCreated   string `form:"end_created"`
	Search       string `form:"search"`
	EmailId      int64  `form:"email_id"`
}

type EmailsGetResult struct {
	Subject    string               `json:"subject"`
	To         []model.EmailAddress `json:"to"`
	From       model.EmailAddress   `json:"from"`
	Content    string               `json:"content"`
	Priority   int                  `json:"priority"`
	Type       string               `json:"type"`
	Status     int                  `json:"status"`
	Error      string               `json:"error"`
	SendTime   int64                `json:"send_time"`
	CreateTime int64                `json:"create_time"`
}

type EmailsGetResultMap map[string]EmailsGetResult

func newEmailsGetResultMap(emailRecords model.EmailRecordList) (EmailsGetResultMap, error) {
	if len(emailRecords) == 0 {
		return make(EmailsGetResultMap), nil
	}

	toListMap, err := emailRecords.GetToListMap()
	if err != nil {
		return nil, err
	}

	res := make(EmailsGetResultMap)
	for _, e := range emailRecords {
		fromAddr := model.EmailAddress{}
		err := json.Unmarshal([]byte(e.From), &fromAddr)
		if err != nil {
			continue
		}

		toList, ok := toListMap[e.Id]
		if !ok || len(toList) == 0 {
			continue
		}

		toAddrList := make([]model.EmailAddress, 0)
		for _, to := range toList {
			toAddrList = append(toAddrList, model.EmailAddress{
				Addr: to.To,
				Name: to.ToName,
			})
		}
		var sendTime int64
		if e.Status == model.EMAIL_RECORD_STATUS_SUCC {
			sendTime = e.SendTime.Unix()
		}
		res[strconv.FormatInt(e.Id, 10)] = EmailsGetResult{
			Subject:    e.Subject,
			To:         toAddrList,
			From:       fromAddr,
			Content:    e.Content,
			Priority:   e.Priority,
			Type:       e.Type,
			Status:     e.Status,
			Error:      e.ErrorInfo,
			SendTime:   sendTime,
			CreateTime: e.CreateTime.Unix(),
		}
	} //end for emailRecords

	return res, nil
}

//*********controller function
func (this *EmailsController) Get() {
	getForm := EmailsGetForm{}
	if err := this.ParseForm(&getForm); err != nil {
		this.ApiUnknowFailed(err.Error())
		return
	}

	valid := validation.Validation{}
	b, err := valid.Valid(&getForm)
	if err != nil {
		this.ApiUnknowFailed(err.Error())
		return
	}

	if !b {
		//处理表单错误信息
		this.ApiUnknowFailed(valid.Errors[0].Message)
		return
	}

	//参数初始化
	if getForm.Page == 0 {
		getForm.Page = 1
	}

	if getForm.PerPage == 0 {
		getForm.PerPage = 20
	}

	endCreated := time.Now()
	startCreated := endCreated.Add(0 - time.Hour*24*7)

	loc, _ := time.LoadLocation("Local")

	isParseTimeErr := false
	if getForm.EndCreated != "" {
		endCreated, err = time.ParseInLocation("2006-01-02", getForm.EndCreated, loc)
		if err != nil {
			isParseTimeErr = true
		}
	}
	endCreated = endCreated.Add(time.Hour * 24)

	if getForm.StartCreated != "" {
		startCreated, err = time.ParseInLocation("2006-01-02", getForm.StartCreated, loc)
		if err != nil {
			isParseTimeErr = true
		}
	}

	if isParseTimeErr {
		this.ApiUnknowFailed("parse time error!")
		return
	}

	if endCreated.Sub(startCreated) > time.Hour*24*31 {
		this.ApiUnknowFailed("max start to end range is 31 days")
		return
	}

	//构造查询条件
	dbSession := model.Orm.Where("email_record.app_id=?", getForm.AppId).Limit(getForm.PerPage, (getForm.Page-1)*getForm.PerPage).OrderBy("email_record.id desc")
	dbSession = dbSession.And("email_record.create_time > ? and email_record.create_time < ?", startCreated.Format("2006-01-02 15:04:05"), endCreated.Format("2006-01-02  15:04:05"))

	status, err := this.GetInt("status", -1234)
	if status != -1234 && err == nil {
		dbSession = dbSession.And("`email_record`.`status`=?", status)
	}

	if getForm.Search != "" {
		dbSession.And("email_record.subject like ?", "%"+getForm.Search+"%")
	}

	if getForm.EmailId != 0 {
		dbSession.And("email_record.id=?", getForm.EmailId)
	}

	if getForm.To != "" {
		dbSession.Join("LEFT OUTER", "email_record_to", "email_record.id=email_record_to.email_record_id").And("email_record_to.to like ?", "%"+getForm.To+"%")
	}

	statement := dbSession.Statement

	//查询数据
	emailRecords := make(model.EmailRecordList, 0)

	err = dbSession.Find(&emailRecords)
	if err != nil {
		this.ApiUnknowFailed("get emails error!")
		return
	}
	dbSession.Statement = statement
	recordsTotalNum, err := dbSession.Count(&model.EmailRecord{})

	emails, _ := newEmailsGetResultMap(emailRecords)

	this.ApiSucc(map[string]interface{}{
		"total":  recordsTotalNum,
		"emails": emails,
	})
	return

}

func (this *EmailsController) Post() {
	postForm := EmailsPostForm{}

	if err := this.ParseForm(&postForm); err != nil {
		this.ApiUnknowFailed(err.Error())
		return
	}

	valid := validation.Validation{}
	b, err := valid.Valid(&postForm)
	if err != nil {
		this.ApiUnknowFailed(err.Error())
		return
	}

	if !b {
		//处理表单信息
		this.ApiUnknowFailed(valid.Errors[0].Message)
		return
	}

	//检查参数正确性
	var toList []model.EmailAddress

	err = json.Unmarshal([]byte(postForm.To), &toList)

	if err != nil {
		//json parse error
		this.ApiUnknowFailed("to:" + err.Error())
		return
	}

	//check to emails
	for _, to := range toList {
		err := util.CheckEmailAddress(&to)
		if err != nil {
			this.ApiUnknowFailed("to:" + err.Error())
			return
		}
	}

	//check from emails
	fromAddr := &model.EmailAddress{}
	err = json.Unmarshal([]byte(postForm.From), &fromAddr)
	if err != nil {
		this.ApiUnknowFailed("from:" + err.Error())
		return
	}

	err = util.CheckEmailAddress(fromAddr)
	if err != nil {
		this.ApiUnknowFailed("from:" + err.Error())
		return
	}

	//check type
	if _, ok := email.EmailManager.Workers[postForm.Type]; !ok {
		this.ApiUnknowFailed("type: not allow type")
		return
	}

	//insert into db
	emailRecord := &model.EmailRecord{
		AppId:    postForm.AppId,
		Type:     postForm.Type,
		From:     postForm.From,
		Subject:  postForm.Subject,
		Content:  postForm.Content,
		Priority: postForm.Priority,
		SendTime: time.Now(),
	}

	_, err = model.Orm.Insert(emailRecord)
	if err != nil {
		this.ApiUnknowFailed("insert email record error!")
		return
	}

	recordToList := make(model.EmailRecordToList, 0)
	//插入 To 数据
	for _, toAddr := range toList {
		emailRecordTo := &model.EmailRecordTo{
			EmailRecordId: emailRecord.Id,
			To:            toAddr.Addr,
			ToName:        toAddr.Name,
		}
		_, err := model.Orm.Insert(emailRecordTo)
		if err != nil {
			//失败了尝试删除一次
			model.Orm.Delete(emailRecord)
			this.ApiUnknowFailed("insert email record to error!")
			return
		}
		recordToList = append(recordToList, *emailRecordTo)
	}

	ok := email.EmailManager.AddEmailToWorker(emailRecord, recordToList)
	if !ok {
		//添加队列失败，删除邮件
		model.Orm.Where("email_record_id=?", emailRecord.Id).Delete(model.EmailRecordTo{})
		model.Orm.Delete(emailRecord)
		this.ApiUnknowFailed("add email to queue error")
		return
	}
	this.ApiSucc(emailRecord.Id)
}

func (this *EmailsController) Put() {
	emailIdsStr := this.GetString("email_ids", "")
	appId := this.GetString("app_id", "")
	if appId == "" || emailIdsStr == "" {
		this.ApiUnknowFailed("params error!")
		return
	}

	emailIdsStrList := strings.Split(emailIdsStr, ",")
	emailIds := make([]int64, 0)
	for _, eStr := range emailIdsStrList {
		eId, err := strconv.ParseInt(eStr, 10, 64)
		if err != nil {
			this.ApiUnknowFailed("email_ids must be number")
			return
		}
		emailIds = append(emailIds, eId)
	}

	emailRecords := make(model.EmailRecordList, 0)
	err := model.Orm.In("id", emailIds).Find(&emailRecords)
	if err != nil {
		this.ApiUnknowFailed(err.Error())
		return
	}

	succIds := make([]int64, 0)

	for _, e := range emailRecords {
		recordToList, err := e.GetToList()
		if err != nil {
			continue
		}

		ok := email.EmailManager.AddEmailToWorker(&e, recordToList)
		if !ok {
			continue
		}

		succIds = append(succIds, e.Id)
	}

	this.ApiSucc(succIds)
	return
}
