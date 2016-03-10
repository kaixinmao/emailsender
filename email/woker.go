package email

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/kaixinmao/emailsender/model"
	"github.com/kaixinmao/emailsender/setting"
	"github.com/kaixinmao/emailsender/util"

	"github.com/kr/beanstalk"
)

const (
	WORKER_STATUS_RUNNING = iota
	WORKER_STATUS_STOPED
	WORKER_STATUS_STOPPING
)

type WorkPack struct {
	Record       *model.EmailRecord
	RecordToList model.EmailRecordToList
}

type Worker struct {
	ManagerWg *sync.WaitGroup
	Sender    ISender
	Id        string
	DoNum     int //并发线程数
	Status    int
	DoWg      sync.WaitGroup
	stopChan  chan int
	workChan  chan *WorkPack
	bkConn    *beanstalk.Conn
	bkPutConn *beanstalk.Conn
}

func NewWorkPack(record *model.EmailRecord, recordToList model.EmailRecordToList) (*WorkPack, error) {
	if record == nil || len(recordToList) == 0 {
		return nil, errors.New("empty params")
	}
	return &WorkPack{
		Record:       record,
		RecordToList: recordToList,
	}, nil
}

func NewWorker(mwg *sync.WaitGroup, sender ISender, workerName string, doNum int) (*Worker, error) {
	bkConn, err := util.NewBkConn(workerName)
	if err != nil {
		return nil, err
	}
	bkPutConn, err := util.NewBkConn(workerName)
	if err != nil {
		return nil, err
	}

	return &Worker{
		ManagerWg: mwg,
		Sender:    sender,
		Id:        workerName,
		DoNum:     doNum,
		Status:    WORKER_STATUS_STOPED,
		stopChan:  make(chan int, doNum),
		workChan:  make(chan *WorkPack, 0), //必须有协程拉取处理
		bkConn:    bkConn,
		bkPutConn: bkPutConn,
	}, nil
}

//只有STOPED状态才能Running，否则返回false
func (this *Worker) Run() bool {
	if this.Status != WORKER_STATUS_STOPED {
		return false
	}

	this.ManagerWg.Add(1)
	for i := 0; i < this.DoNum; i++ {
		this.DoWg.Add(1)
		go this.doSend()
	}

	//开始分发
	this.doDispatch()

	this.Status = WORKER_STATUS_RUNNING
	return true
}

func (this *Worker) PutWorkPack(wp *WorkPack) bool {
	ok := false
	select {
	case this.workChan <- wp:
		ok = true
	default:
		ok = false
	}
	return ok
}

//停止线程工作
func (this *Worker) Stop() bool {
	if this.Status == WORKER_STATUS_STOPPING || this.Status == WORKER_STATUS_STOPED {
		return false
	}

	//一个队列分发协程需要关闭
	for i := 0; i < this.DoNum+1; i++ {
		this.stopChan <- 1
	}
	this.Status = WORKER_STATUS_STOPPING
	return true
}

func (this *Worker) doDispatch() {
	this.DoWg.Add(1)
	go func() {
		defer this.DoWg.Done()
		//这里不要等待太长，否则影响进程退出
		for {
			//wait stop
			select {
			case _ = <-this.stopChan:
				return
			default:
			}

			id, body, err := this.bkConn.Reserve(5 * time.Second)
			if err != nil {
				continue
			}
			this.bkConn.Delete(id)

			bodyStr := string(body)
			if bodyStr == "" {
				continue
			}

			emailId, err := strconv.ParseInt(bodyStr, 10, 64)
			if err != nil {
				continue
			}

			//获取数据库信息
			emailRecord := &model.EmailRecord{Id: emailId}

			has, err := model.Orm.Get(emailRecord)
			if !has {
				//没有数据，就不处理了
				continue
			}

			if err != nil {
				//数据库错误 尝试Delay发送
				this.AddEmailId(emailId, setting.QUEUE_MAX_PRI, 5)
				time.Sleep(5)
				continue
			}

			recordToList, err := emailRecord.GetToList()
			if err != nil {
				//肯定是数据库错误
				this.AddEmailId(emailId, setting.QUEUE_MAX_PRI, 5)
				time.Sleep(5)
				continue
			}

			//数据有问题，直接过
			if len(recordToList) == 0 {
				continue
			}
			wp, err := NewWorkPack(emailRecord, recordToList)
			if err != nil {
				continue
			}

			this.workChan <- wp
		}
	}()
}

func (this *Worker) doSend() {
	defer this.DoWg.Done()

	//每个工作协程一个tube
	var wp *WorkPack
	for {
		select {
		case _ = <-this.stopChan:
			return
		case wp = <-this.workChan:
		}

		//开始发送邮件

		ok, err := this.Sender.Send(wp.Record, wp.RecordToList)
		if ok {
			wp.Record.Status = model.EMAIL_RECORD_STATUS_SUCC
			model.Orm.Update(wp.Record)
		} else {
			wp.Record.Status = model.EMAIL_RECORD_STATUS_FAILED
			wp.Record.ErrorInfo = err.Error()
			model.Orm.Update(wp.Record)
		}

	}

}

func (this *Worker) AddEmailId(id int64, pri uint32, delay int) bool {
	idStr := strconv.FormatInt(id, 10)
	_, err := this.bkPutConn.Put([]byte(idStr), pri, time.Duration(delay)*time.Second, setting.QUEUE_TTR*time.Second)
	if err != nil {
		return false
	} else {
		return true
	}
}

//等待工作协程结束，如果为RUNNING状态，返回false
func (this *Worker) Wait() bool {
	if this.Status == WORKER_STATUS_RUNNING {
		return false
	}

	this.DoWg.Wait()
	this.ManagerWg.Done()
	this.Status = WORKER_STATUS_STOPED
	fmt.Printf("%s all worker stoped", this.Id)
	return true
}
