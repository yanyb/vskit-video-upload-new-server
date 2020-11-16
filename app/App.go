package app

import (
	"fmt"
	"github.com/vskit-tv/vlog/log"
	"go-app/service"
	"net"
)

var _profile string

// 初始化服务组件
func InitApp(configPath string, profile string) {
	initCfg(configPath)
	log.InitLog(&_cfg.Logger)
	_profile = profile
}

func GetProfile() string {
	return _profile
}

// 启动服务
func Start() {
	addr := ":" + _cfg.App["port"]
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		go func() {
			if err := recover(); err != nil {
				fmt.Println("handle conn:", err)
			}
			service.HandleConn(c)
		}()
	}
}
