package config

import (
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
	once:           sync.NewOnce(),
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
	once           sync.Once
}

func (c *context) parseConfigFile(filename string, paths ...string) error {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(absFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var fileMap map[string]any

	if err = yaml.Unmarshal(buf, &fileMap); err != nil {
		return fmt.Errorf("unmarshal file '%s' err: %v", filename, err)
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

func (c *context) addFile(filename string, paths ...string) {
	c.configsLock.Lock()
	defer c.configsLock.Unlock()

	c.configs = append(c.configs, pathConfig{paths: paths, filename: filename})
}

func (c *context) readInConfig() error {
	return c.once.Do(func() error {
		if err := c.parseConfigFile(_defaultConfigFileName); err != nil {
			return err
		}

		c.configsLock.RLock()
		defer c.configsLock.RUnlock()

		for _, config := range c.configs {
			if err := c.parseConfigFile(config.filename, config.paths...); err != nil {
				return err
			}
		}

		return nil
	})
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
