package config

import (
	"errors"
	"github.com/expgo/generic"
	"gopkg.in/yaml.v3"
	"os"
)

type Config interface {
}

type context struct {
	fileConfigs  generic.Map[string, []Config]
	fileContents generic.Cache[string, string]
}

var _defaultConfigFileName = "app.yml"

func SetDefault(filename string, cfg any) error {
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

	_defaultConfigFileName = filename
	return nil
}

func SetDefaultFilename(filename string) error {
	if len(filename) == 0 {
		return errors.New("filename must not be empty")
	}

	// Determine whether the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}

	_defaultConfigFileName = filename
	return nil
}

// SetDefaultConfig if file not exists, create file by cfg struct, then set file to default config file
func SetDefaultConfig[T Config](filename string, cfg *T) error {
	return SetDefault(filename, cfg)
}

// New create config with default config file
func New[T Config](path string) *T {
	//t := factory.New[T]()
	return nil
}

func NewWithFile[T Config](filename string, path string) *T {
	return nil
}
