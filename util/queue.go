package util

import (
	"github.com/kaixinmao/emailsender/setting"
	"errors"

	"github.com/kr/beanstalk"
)

var bkHost, bkPort string

func NewBkConn(name string) (*beanstalk.Conn, error) {
	prefix := setting.Cfg.GetPath("queue", "tube_prefix").MustString()
	if prefix == "" {
		panic("tube_prefix empty or name prefix")

	}

	if name == "" {
		return nil, errors.New("NewBkTube prefix empty or name prefix")
	}

	tmpBkConn, err := beanstalk.Dial("tcp", bkHost+":"+bkPort)
	if err != nil {
		panic(err)
	}
	tubeName := prefix + name

	tmpBkConn.Tube = beanstalk.Tube{tmpBkConn, tubeName}
	tmpBkConn.TubeSet = *beanstalk.NewTubeSet(tmpBkConn, tubeName)

	return tmpBkConn, nil
}

func init() {
	queueJson := setting.Cfg.Get("queue")
	bkHost = queueJson.Get("host").MustString()
	bkPort = queueJson.Get("port").MustString("11300")
	if bkHost == "" {
		panic("queue host empty")
	}
}
