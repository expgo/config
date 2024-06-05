package config

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Cfg struct {
	Abc string
	Def int
}

func TestConfigPath(t *testing.T) {
	cfg := &Cfg{}
	__context.configTree["cfg"] = map[string]any{"Abc": "abc", "Def": 123}

	__context.getConfig(cfg, "cfg", "ttt")
	assert.Equal(t, "", cfg.Abc)
	assert.Equal(t, 0, cfg.Def)

	__context.getConfig(cfg, "cfg")
	assert.Equal(t, "abc", cfg.Abc)
	assert.Equal(t, 123, cfg.Def)
}

func TestParseConfigFile(t *testing.T) {
	__context.parseConfigFile("test.yml")
	__context.parseConfigFile("test-log.yml", "test", "logging")
	__context.parseConfigFile("test-log-ext.yml", "test")

	jsonData, err := json.Marshal(__context.configTree)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %s", err)
	}
	assert.Equal(t, "{\"age\":5,\"env\":\"dev\",\"freq\":1000,\"test\":{\"logbool\":true,\"logext\":\"ttt\",\"logging\":{\"name\":\"abc\",\"size\":10}}}", string(jsonData))

	err = __context.parseConfigFile("test-log-ext.yml", "test", "logbool")
	assert.Equal(t, errors.New("parse 'test-log-ext.yml' err. path 'test.logbool' already exists, but not map[string]any type"), err)
}

func TestSetAndGetValue(t *testing.T) {
	err := __context.setValue(true, "my", "test", "bool")
	assert.Nil(t, err)
	assert.Equal(t, true, MustGet("my", "test", "bool"))

	err = __context.setValue(12, "my", "test", "bool", "int")
	assert.Error(t, errors.New("path 'my.test.bool' already exists, but not map[string]any type"), err)

	_, err = __context.getValue("my", "test", "bool", "int")
	assert.Error(t, errors.New("path 'my.test.bool' must be map[string]any"), err)
}
