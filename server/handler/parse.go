package handler

import (
	"encoding/json"
	"github.com/golang/glog"
	"os"
	"path/filepath"
)

func ParseConfigs(path string) (handlerConfigs []Handler, err error) {
	glog.V(2).Infof("Looking for handlers in '%s'", path)
	absHandlerPath, err := filepath.Abs(path)
	if err != nil {
		glog.Fatalf("filepath.Abs(%s): %s", path, err)
	}

	pattern := filepath.Join(absHandlerPath, "*", "handler.json")
	handlerConfigPaths, err := filepath.Glob(pattern)
	if err != nil {
		glog.Fatalf("filepath.Glob(%s): %s", pattern, err)
	}
	glog.V(2).Infof("Found %d configs", len(handlerConfigPaths))

	handlerConfigs = make([]Handler, 0, len(handlerConfigPaths))
	for _, configPath := range handlerConfigPaths {
		f := new(os.File)
		if f, err = os.Open(configPath); err != nil {
			glog.Errorf("os.Open(%s): %s", configPath, err)
			continue
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		handlerConfig := new(Handler)
		if err := dec.Decode(handlerConfig); err != nil {
			glog.Errorf("json.Decode: %s", err)
			continue
		}
		handlerConfig.SetDir(filepath.Dir(configPath))
		glog.V(2).Infof("Adding handler config: %s", handlerConfig.Name)
		handlerConfigs = append(handlerConfigs, *handlerConfig)
	}
	return
}
