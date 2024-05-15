package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
)

type Config interface {
	Commit(to map[string]any) bool
	Verify(to map[string]any) error
}

var _defaultConfigFileName = "app.yml"

// SaveIfNotExist if file not exists, create file by cfg struct, then set file to default config file
func SaveIfNotExist(filename string, cfg any) error {
	if len(filename) == 0 {
		return errors.New("filename must not be empty")
	}

	if _, oserr := os.Stat(filename); os.IsNotExist(oserr) {
		// set cfg to file
		if buf, err := yaml.Marshal(cfg); err != nil {
			return err
		} else {
			// save buf to file
			err = os.WriteFile(filename, buf, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DefaultFile(filename string) {
	if len(filename) > 0 {
		_defaultConfigFileName = filename
	}
}

func AddFile(filename string, paths ...string) {
	__context.addFile(filename, paths...)
}

func ReadInConfig() error {
	return __context.readInConfig()
}

func GetConfig(cfg any, paths ...string) error {
	cfgType := reflect.TypeOf(cfg)
	if cfgType.Kind() != reflect.Ptr || cfgType.Elem().Kind() != reflect.Struct {
		return errors.New("config must be a point struct")
	}

	if err := __context.readInConfig(); err != nil {
		return err
	}

	return __context.getConfig(cfg, paths...)
}

func SetConfig(cfg any, paths ...string) error {
	cfgType := reflect.TypeOf(cfg)
	if cfgType.Kind() != reflect.Ptr || cfgType.Elem().Kind() != reflect.Struct {
		return errors.New("config must be a point struct")
	}

	return __context.setConfig(cfg, paths...)
}
