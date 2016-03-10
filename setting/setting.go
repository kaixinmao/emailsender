package setting

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
)

const (
	APP_VERSION = "0.1"
	//BK QUEUE 默认 ttr 参数
	QUEUE_TTR = 5
	//最大优先级
	QUEUE_MAX_PRI = 100
)

var (
	//监听端口
	AppPath    string
	AppWorkDir string
	IsWindows  bool

	Cfg *simplejson.Json
)

func init() {
	//初始化环境变量
	file, _ := exec.LookPath(os.Args[0])
	AppPath, _ := filepath.Abs(file)
	i := strings.LastIndex(AppPath, "/")
	if i == -1 {
		AppWorkDir = AppPath
	} else {
		AppWorkDir = AppPath[:i]
	}

	//初始化配置文件
	fi, err := os.Open(AppWorkDir + "/config/main.json")
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	Cfg, err = simplejson.NewFromReader(fi)
	if err != nil {
		panic(err)
	}
}
