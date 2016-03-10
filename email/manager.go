package email

import (
	"errors"
	"sync"

	"github.com/kaixinmao/emailsender/model"
	"github.com/kaixinmao/emailsender/setting"

	simplejson "github.com/bitly/go-simplejson"
)

var EmailManager *Manager

type Manager struct {
	Workers map[string]*Worker
	Senders map[string]ISender
	Wg      sync.WaitGroup
}

//Stop all works and wait works stop
func (this *Manager) WaitStop() {
	for _, worker := range this.Workers {
		worker.Stop()

	}

	for _, worker := range this.Workers {
		worker.Wait()
	}
	this.Wg.Wait()
}

//start run
func (this *Manager) Run() bool {
	allRunOk := true
	for _, worker := range this.Workers {
		if ok := worker.Run(); !ok {
			allRunOk = false
			break
		}
	}

	return allRunOk
}

//Add email to queue
func (this *Manager) AddEmailToWorker(record *model.EmailRecord, recordToList model.EmailRecordToList) bool {
	worker, ok := this.Workers[record.Type]
	if !ok {
		return false
	}

	wp, err := NewWorkPack(record, recordToList)
	if err != nil {
		return false
	}

	//尝试直接发送
	if ok := worker.PutWorkPack(wp); ok {
		return true
	}

	//加入队列
	return worker.AddEmailId(record.Id, uint32(record.Priority), 0)
}

//通过setting配置初始化Manager和相关Sender、Worker
func NewManagerBySetting() (*Manager, error) {
	m := &Manager{}
	senders, err := getSenderMapBySetting(setting.Cfg.Get("senders"))
	if err != nil {
		return nil, err
	}
	m.Senders = senders

	workers, err := getWorkerMapBySetting(setting.Cfg.Get("workers"), senders, &m.Wg)
	if err != nil {
		return nil, err
	}

	m.Workers = workers

	return m, nil
}

func getWorkerMapBySetting(json *simplejson.Json, senderMap map[string]ISender, mwg *sync.WaitGroup) (map[string]*Worker, error) {
	res := make(map[string]*Worker)

	for workerName, _ := range json.MustMap() {
		workerCfgJson := json.Get(workerName)
		senderName := workerCfgJson.Get("sender").MustString()
		doNum := workerCfgJson.Get("worker_num").MustInt(5)
		if sender, has := senderMap[senderName]; has {
			worker, err := NewWorker(mwg, sender, workerName, doNum)
			if err != nil {
				return nil, err
			}

			res[workerName] = worker
		} else {
			continue
		}
	}

	return res, nil
}

func getSenderMapBySetting(json *simplejson.Json) (map[string]ISender, error) {
	res := make(map[string]ISender)
	groupSenderCfgs := make(map[string]*simplejson.Json)
	for senderName, _ := range json.MustMap() {
		senderCfgJson := json.Get(senderName)
		senderType, err := GetTypeByStr(senderCfgJson.Get("type").MustString())
		if err != nil {
			continue
		}
		switch senderType {
		case SENDER_TYPE_SMTP:
			sender, err := getSmtpSenderBySetting(senderCfgJson)
			if err != nil {
				continue
			}
			res[senderName] = sender
		case SENDER_TYPE_GROUP:
			groupSenderCfgs[senderName] = senderCfgJson
		default:
			continue
		}
	}

	for groupSenderName, cfgJson := range groupSenderCfgs {
		sender, err := getGroupSenderBySettingAndSenders(cfgJson, res)
		if err != nil {
			return nil, err
		}
		res[groupSenderName] = sender
	}
	return res, nil
}

func getGroupSenderBySettingAndSenders(cfgJson *simplejson.Json, senders map[string]ISender) (ISender, error) {
	senderNames := cfgJson.Get("senders").MustArray()
	if len(senderNames) == 0 {
		return nil, errors.New("group sender no other senders")
	}

	groupChildSenders := make([]ISender, 0)

	for _, iName := range senderNames {
		name, ok := iName.(string)
		if !ok {
			panic("group sender config error!")
		}
		sender, has := senders[name]

		if sender.GetType() == SENDER_TYPE_GROUP {
			panic("group sender can't add group sender")
		}

		if !has {
			panic("group sender not found sender:" + name)
		}
		groupChildSenders = append(groupChildSenders, sender)
	}

	groupSender, err := NewGroupSender(groupChildSenders...)

	return groupSender, err
}

func getSmtpSenderBySetting(cfgJson *simplejson.Json) (ISender, error) {
	host := cfgJson.Get("smtp_host").MustString()
	port := cfgJson.Get("smtp_port").MustString()
	username := cfgJson.Get("smtp_username").MustString()
	password := cfgJson.Get("smtp_password").MustString()
	secure := cfgJson.Get("smtp_secure").MustString()

	sender, err := NewSmtpSender(host, port, username, password, secure)

	return ISender(sender), err
}

func init() {
	tmpManager, err := NewManagerBySetting()
	if err != nil {
		panic(err)
	}

	EmailManager = tmpManager
}
