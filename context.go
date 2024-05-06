package config

import (
	"github.com/expgo/structure"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sync"
)

var __context = &context{}

type context struct {
	fileConfigs      map[string][]Config
	fileConfigsLock  sync.RWMutex
	fileContents     map[string]map[string]any
	fileContentsLock sync.RWMutex
}

func (c *context) GetConfig(filename string, cfg any, paths ...string) error {
	fn, oserr := filepath.Abs(filename)
	if oserr != nil {
		return oserr
	}

	c.fileContentsLock.RLock()
	fileMap, ok := c.fileContents[fn]
	c.fileContentsLock.RUnlock()

	if !ok {
		buf, err := os.ReadFile(fn)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if err = yaml.Unmarshal(buf, &fileMap); err != nil {
			return err
		}

		c.fileContentsLock.Lock()
		c.fileContents[fn] = fileMap
		c.fileContentsLock.Unlock()
	}

	if len(paths) > 0 {
		var lastPathValue any
		for _, path := range paths {
			if pathFileMap, ok := fileMap[path]; ok {
				lastPathValue = pathFileMap

				if pathFileMap != nil {
					if fileMap, ok = pathFileMap.(map[string]any); ok {
						continue
					}
				}

				break
			} else {
				lastPathValue = nil
			}
		}

		if lastPathValue != nil {
			return structure.Map(lastPathValue, cfg)
		}
	} else {
		return structure.Map(fileMap, cfg)
	}

	return nil
}
