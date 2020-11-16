package model

import (
	_model "github.com/vskit-tv/vlog/log/model"
)

type Config struct {
	App    map[string]string   `yaml:"app,omitempty"`
	Logger _model.LoggerConfig `yaml:"logger,omitempty"`
}
