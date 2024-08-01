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
	configs:          []pathConfig{},
	configsLock:      sync.NewRWMutex(),
	configTree:       map[string]any{},
	configTreeLock:   sync.NewRWMutex(),
	once:             sync.NewOnce(),
	loadedFiles:      map[string][]string{},
	loadedFilesLock:  sync.NewMutex(),
	fileWatchers:     []FileWatcher{},
	fileWatchersLock: sync.NewMutex(),
	pathWatchers:     map[string][]PathWatcher{},
	pathWatchersLock: sync.NewMutex(),
}

type pathConfig struct {
	paths    []string
	filename string
}

type context struct {
	configs          []pathConfig
	configsLock      sync.RWMutex
	configTree       map[string]any
	configTreeLock   sync.RWMutex
	once             sync.Once
	loadedFiles      map[string][]string
	loadedFilesLock  sync.Mutex
	fileWatchers     []FileWatcher
	fileWatchersLock sync.Mutex
	pathWatchers     map[string][]PathWatcher
	pathWatchersLock sync.Mutex
}

func isTestProcess() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func getAbsPath(filename string) (string, error) {
	if isTestProcess() {
		if filepath.IsAbs(filename) {
			return filename, nil
		} else {
			// if direct abs file exist, return it
			absFilePath, err := filepath.Abs(filename)
			if err != nil {
				return "", err
			}
			if fileExists(absFilePath) {
				return absFilePath, nil
			}

			// search go.mod to find project root dir
			projectPath := filepath.Dir(absFilePath)
			for {
				if _, err = os.Stat(filepath.Join(projectPath, "go.mod")); err == nil {
					break
				}
				projectPath = filepath.Dir(projectPath)
				if projectPath == "/" || projectPath == "." {
					// not find go.mod, return absFilePath
					return absFilePath, nil
				}
			}

			return filepath.Join(projectPath, filename), nil
		}
	} else {
		return filepath.Abs(filename)
	}
}

func (c *context) mergeTrees(map1, map2 map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range map1 {
		if vMap1, ok := v.(map[string]any); ok {
			if vMap2, ok := map2[k].(map[string]any); ok {
				result[k] = c.mergeTrees(vMap1, vMap2)
			} else {
				result[k] = v
			}
		} else {
			result[k] = v
		}
	}
	for k, v := range map2 {
		if _, ok := map1[k]; !ok {
			result[k] = v
		}
	}
	return result
}

func (c *context) updateConfigTree(fileMap map[string]any, paths ...string) error {
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
					return fmt.Errorf("path '%s' already exists, but not map[string]any type", strings.Join(paths[:i+1], "."))
				}
			} else {
				configTreeMap[path] = map[string]any{}
				configTreeMap = configTreeMap[path].(map[string]any)
			}
		}

		lastConfigMapParent[paths[len(paths)-1]] = c.mergeTrees(fileMap, c.configTree)
		return nil
	} else {
		c.configTree = c.mergeTrees(fileMap, c.configTree)
		return nil
	}
}

func (c *context) parseConfigFile(filename string, paths ...string) error {
	absFilePath, err := getAbsPath(filename)
	if err != nil {
		return err
	}

	if !fileExists(absFilePath) {
		return nil
	}

	c.loadedFilesLock.Lock()
	c.loadedFiles[absFilePath] = paths
	for _, fw := range c.fileWatchers {
		fw.WatchFile(absFilePath)
	}
	c.loadedFilesLock.Unlock()

	return c.loadConfigToTree(absFilePath, paths...)
}

func (c *context) loadConfigToTree(filename string, paths ...string) error {
	buf, err := os.ReadFile(filename)
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

	return c.updateConfigTree(fileMap, paths...)
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

func (c *context) getValue(paths ...string) (any, error) {
	c.configTreeLock.RLock()
	defer c.configTreeLock.RUnlock()

	fileMap := c.configTree
	for i, path := range paths {
		if fileMap != nil {
			if pathValue, ok := fileMap[path]; ok {
				if fileMap, ok = pathValue.(map[string]any); !ok {
					if i == (len(paths) - 1) {
						return pathValue, nil
					} else {
						return nil, fmt.Errorf("path '%s' must be map[string]any", strings.Join(paths[:len(paths)-1], "."))
					}
				}
			} else {
				return nil, errors.New("path not found")
			}
		} else {
			return nil, errors.New("path not found")
		}
	}
	return fileMap, nil
}

func (c *context) setValue(value any, paths ...string) error {
	c.configTreeLock.Lock()
	defer c.configTreeLock.Unlock()

	configTreeMap := c.configTree
	var lastConfigMapParent map[string]any

	for i, path := range paths {
		lastConfigMapParent = configTreeMap

		if lastPathValue, ok := configTreeMap[path]; ok {
			if ctm, ok1 := lastPathValue.(map[string]any); ok1 {
				configTreeMap = ctm
			} else {
				return fmt.Errorf("path '%s' already exists, but not map[string]any type", strings.Join(paths[:i+1], "."))
			}
		} else {
			configTreeMap[path] = map[string]any{}
			configTreeMap = configTreeMap[path].(map[string]any)
		}
	}

	lastConfigMapParent[paths[len(paths)-1]] = value
	return nil
}
