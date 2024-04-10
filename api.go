package config

import (
	"errors"
	"github.com/expgo/factory"
	"gopkg.in/yaml.v3"
	"os"
)

type Config interface {
}

var _defaultConfigFileName = "app.yml"

// SetDefault if file not exists, create file by cfg struct, then set file to default config file
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

// New create config with default config file
func New[T Config](path string) (*T, error) {
	cfg := factory.New[T]()
	err := __context.GetConfig(_defaultConfigFileName, path, cfg)
	return cfg, err
}

func NewWithFile[T Config](filename string, path string) (*T, error) {
	cfg := factory.New[T]()
	err := __context.GetConfig(filename, path, cfg)
	return cfg, err
}
