package config

import (
	"errors"
	"fmt"
	"github.com/expgo/structure"
	"github.com/expgo/sync"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

var __context = &context{
	configs:        []pathConfig{},
	configsLock:    sync.NewRWMutex(),
	configTree:     map[string]any{},
	configTreeLock: sync.NewRWMutex(),
}

type pathConfig struct {
	paths    []string
	filename string
}

type context struct {
	configs        []pathConfig
	configsLock    sync.RWMutex
	configTree     map[string]any
	configTreeLock sync.RWMutex
}

func checkFilenameValid(filename string) error {
	if len(filename) == 0 {
		return errors.New("filename must not be empty")
	}

	// Determine whether the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return errors.New("file does not exist")
	}

	return nil
}

func (c *context) parseConfigFile(filename string, paths ...string) error {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(absFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var fileMap map[string]any

	if err = yaml.Unmarshal(buf, &fileMap); err != nil {
		return err
	}

	c.configTreeLock.Lock()
	defer c.configTreeLock.Unlock()

	if len(paths) > 0 {
		configTreeMap := c.configTree
		var lastConfigMapParent map[string]any

		for i, path := range paths {
			lastConfigMapParent = configTreeMap

			if lastPathValue, ok := configTreeMap[path]; ok {
				if ctm, ok1 := lastPathValue.(map[string]any); ok1 {
					configTreeMap = ctm
				} else {
					return fmt.Errorf("parse '%s' err. path '%s' already exists, but not map[string]any type", filename, strings.Join(paths[:i+1], "."))
				}
			} else {
				configTreeMap[path] = map[string]any{}
				configTreeMap = configTreeMap[path].(map[string]any)
			}
		}

		if err = structure.Map(fileMap, &configTreeMap); err != nil {
			return err
		}

		lastConfigMapParent[paths[len(paths)-1]] = configTreeMap
		return nil
	} else {
		return structure.Map(fileMap, &c.configTree)
	}
}

func (c *context) addPathConfig(filename string, paths ...string) error {
	c.configsLock.Lock()
	defer c.configsLock.Unlock()

	if err := checkFilenameValid(filename); err != nil {
		return err
	}

	c.configs = append(c.configs, pathConfig{paths: paths, filename: filename})

	return nil
}

func (c *context) getConfig(cfg any, paths ...string) error {
	c.configTreeLock.RLock()
	defer c.configTreeLock.RUnlock()

	if len(paths) > 0 {
		fileMap := c.configTree
		var lastPathValue any
		for _, path := range paths {
			if fileMap != nil {
				if pathFileMap, ok := fileMap[path]; ok {
					lastPathValue = pathFileMap

					if pathFileMap != nil {
						if fileMap, ok = pathFileMap.(map[string]any); !ok {
							fileMap = nil
						}
					}
				} else {
					lastPathValue = nil
				}
			} else {
				lastPathValue = nil
			}
		}

		if lastPathValue != nil {
			return structure.Map(lastPathValue, cfg)
		}
	} else {
		return structure.Map(c.configTree, cfg)
	}

	return nil
}
