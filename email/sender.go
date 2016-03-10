package email

import (
	"errors"
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"

	"github.com/kaixinmao/emailsender/model"
)

const (
	SENDER_TYPE_SMTP = iota
	SENDER_TYPE_GROUP
)

type ISender interface {
	Send(record *model.EmailRecord, recordToList model.EmailRecordToList) (bool, error)
	GetType() int
}

type baseSender struct {
	Type int
}

type SmtpSender struct {
	baseSender
	Host     string
	Port     string
	Username string
	Password string
	Secure   string
}

func (this *baseSender) GetType() int {
	return this.Type
}

func (this *SmtpSender) Send(record *model.EmailRecord, recordToList model.EmailRecordToList) (bool, error) {
	auth := smtp.PlainAuth("", this.Username, this.Password, this.Host)
	toList := make([]string, 0)
	for _, t := range recordToList {
		toList = append(toList, t.To)
	}
	fromAddr, err := record.GetFromAddr()
	if err != nil {
		return false, err
	}
	msg := fmt.Sprintf("subject: %s\r\nContent-Type: text/html; charset=UTF-8\r\nTo: %s\r\nFrom: %s<%s>\r\n\r\n", record.Subject, strings.Join(toList, ","), fromAddr.Addr, fromAddr.Addr)

	err = smtp.SendMail(this.Host+":"+this.Port, auth, this.Username, toList, []byte(msg+record.Content))
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetTypeByStr(typeStr string) (int, error) {
	switch typeStr {
	case "smtp":
		return SENDER_TYPE_SMTP, nil
	case "group":
		return SENDER_TYPE_GROUP, nil
	default:
		return 0, errors.New("no valid type")
	}
}

func NewSmtpSender(host, port, username, password, secure string) (*SmtpSender, error) {
	if host == "" || port == "" || username == "" || password == "" {
		return nil, errors.New("params must be not empty")
	}

	return &SmtpSender{
		baseSender: baseSender{Type: SENDER_TYPE_SMTP},
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		Secure:     secure,
	}, nil
}

type GroupSender struct {
	baseSender
	Senders []ISender
}

func NewGroupSender(senders ...ISender) (*GroupSender, error) {
	return &GroupSender{
		baseSender: baseSender{Type: SENDER_TYPE_GROUP},
		Senders:    senders,
	}, nil
}

func (this *GroupSender) Send(record *model.EmailRecord, recordToList model.EmailRecordToList) (bool, error) {
	senderSize := len(this.Senders)
	senderIndex := rand.Intn(senderSize)
	sender := this.Senders[senderIndex]
	return sender.Send(record, recordToList)
}
