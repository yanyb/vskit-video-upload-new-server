package app

import (
	"github.com/vskit-tv/vlog/log"
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
