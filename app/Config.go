package app

import (
	"github.com/vskit-tv/vcomm-go/util/json"
	"github.com/vskit-tv/vcomm-go/util/yaml"
	"go-app/model"
)

var _cfg *model.Config

// 初始化配置文件
func initCfg(configPath string) {
	var config model.Config
	if err := yaml.YamlLoadFromPath(configPath, &config); err != nil {
		panic(err)
	}
	json.PrintJSON(config)
	_cfg = &config
}

func GetDataPath() string {
	return _cfg.App["data_dir"]
}

func GetAddr() string {
	return ":" + _cfg.App["port"]
}
