package config

import (
	"errors"
	"github.com/expgo/generic/set"
	"github.com/expgo/structure"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"
)

type FileWatcher interface {
	WatchFile(filename string)
}

type PathWatcher interface {
	ConfigUpdate(cfg map[string]any)
}

var ERROR_PATH = errors.New("must set at least one path")

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

// DefaultFile 设置默认配置文件名，默认文件为全局config tree
func DefaultFile(filename string) {
	if len(filename) > 0 {
		_defaultConfigFileName = filename
	}
}

// AddFile 通过AddFile将文件和配置树进行绑定，但AddFile不真正加载文件，会在首次GetConfig时，加载所有文件到config tree
func AddFile(filename string, paths ...string) {
	__context.addFile(filename, paths...)
}

// ReadInConfig 强制从添加的文件加载到config tree
func ReadInConfig() error {
	return __context.readInConfig()
}

// GetConfig 从config tree上的path中加载配置对象，首次获取会将通过AddFile和DefaultFile的文件先加载到config tree
func GetConfig(cfg any, paths ...string) error {
	cfgType := reflect.TypeOf(cfg)
	if !(cfgType.Kind() == reflect.Ptr && (cfgType.Elem().Kind() == reflect.Struct || cfgType.Elem().Kind() == reflect.Map)) {
		return errors.New("config must be a point struct or point map")
	}

	if err := __context.readInConfig(); err != nil {
		return err
	}

	return __context.getConfig(cfg, paths...)
}

// SetConfig 将配置对象设置到config tree的某个path上
func SetConfig(cfg any, paths ...string) error {
	cfgType := reflect.TypeOf(cfg)
	if !((cfgType.Kind() == reflect.Ptr && cfgType.Elem().Kind() == reflect.Struct) || cfgType.Kind() == reflect.Map) {
		return errors.New("config must be a point struct or map")
	}

	if fileMap, ok := cfg.(map[string]any); ok {
		return __context.updateConfigTree(fileMap, paths...)
	} else {
		if err := structure.Map(cfg, &fileMap); err != nil {
			return err
		}
		return __context.updateConfigTree(fileMap, paths...)
	}
}

// GetValue 从config tree上根据path获取对应配置值
func GetValue(paths ...string) (any, error) {
	if len(paths) == 0 {
		return nil, ERROR_PATH
	}

	return __context.getValue(paths...)
}

// SetValue 设置配置值到config tree的path上
func SetValue(value any, paths ...string) error {
	if len(paths) == 0 {
		return ERROR_PATH
	}

	return __context.setValue(value, paths...)
}

func MustGet(paths ...string) any {
	ret, err := GetValue(paths...)
	if err != nil {
		panic(err)
	}
	return ret
}

// FileUpdate 调用进行文件更新，该更新的文件名需要是abs路径，且通过AddFile添加，一般用于文件监听器的更新动作
func FileUpdate(filename string) error {
	__context.loadedFilesLock.Lock()
	defer __context.loadedFilesLock.Unlock()

	if paths, ok := __context.loadedFiles[filename]; ok {
		return __context.loadConfigToTree(filename, paths...)
	}

	return nil
}

// AddFileWatcher 添加FileWatcher的监听方法，注册后，通过该接口通知文件修改
func AddFileWatcher(fileWatcher FileWatcher) {
	__context.fileWatchersLock.Lock()
	defer __context.fileWatchersLock.Unlock()

	__context.fileWatchers = append(__context.fileWatchers, fileWatcher)

	__context.loadedFilesLock.Lock()
	defer __context.loadedFilesLock.Unlock()

	for filename, _ := range __context.loadedFiles {
		fileWatcher.WatchFile(filename)
	}
}

func pathsToPath(paths ...string) string {
	path := "/"
	if len(paths) > 0 {
		path = path + strings.Join(paths, "/")
	}
	return path
}

func AddPathWatcher(pathWatcher PathWatcher, paths ...string) {
	__context.pathWatchersLock.Lock()
	defer __context.pathWatchersLock.Unlock()

	path := pathsToPath(paths...)

	if _, ok := __context.pathWatchers[path]; !ok {
		__context.pathWatchers[path] = []PathWatcher{}
	}

	__context.pathWatchers[path], _ = set.Add(__context.pathWatchers[path], pathWatcher)
}
