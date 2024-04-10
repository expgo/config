package config

import (
	"github.com/expgo/generic"
	"github.com/expgo/structure"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// @Singleton(localVar)
type context struct {
	fileConfigs  generic.Map[string, []Config]
	fileContents generic.Cache[string, map[string]any]
}

func (c *context) GetConfig(filename string, path string, cfg any) error {
	fn, oserr := filepath.Abs(filename)
	if oserr != nil {
		return oserr
	}

	fileMap, cerr := c.fileContents.GetOrLoad(fn, func(absFilename string) (map[string]any, error) {
		// Load absFilename and parse it into map[string]any using the yml library
		buf, err := os.ReadFile(absFilename)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		var data map[string]any
		if err = yaml.Unmarshal(buf, &data); err != nil {
			return nil, err
		}

		return data, nil
	})

	if cerr != nil {
		return cerr
	}

	if len(path) > 0 {
		if pathFileMap, ok := fileMap[path]; ok {
			return structure.Map(pathFileMap, cfg)
		}
	} else {
		return structure.Map(fileMap, cfg)
	}

	return nil
}
