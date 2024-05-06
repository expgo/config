package config

import (
	"errors"
	"github.com/expgo/factory"
	"gopkg.in/yaml.v3"
	"os"
)

type Config interface {
	Commit(to map[string]any) bool
	Verify(to map[string]any) error
}

func Modify[T any](path string, modFunc func(cfg *T)) {

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

func DefaultFile(filename string) error {
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

func WatchFile(filename string, paths ...string) error {

	return nil
}

// New create config with default config file
func New[T any](paths ...string) (*T, error) {
	cfg := factory.New[T]()
	return cfg, __context.GetConfig(_defaultConfigFileName, cfg, paths...)
}

func NewWithFile[T any](filename string, paths ...string) (*T, error) {
	cfg := factory.New[T]()
	return cfg, __context.GetConfig(filename, cfg, paths...)
}
