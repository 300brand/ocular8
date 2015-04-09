package config

import (
	"encoding/json"
	"flag"
	"path/filepath"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/golang/glog"
)

var (
	flagEtcd = flag.String("etcd", "http://localhost:4001", "Etcd server address")
	mapping  = make(map[string]*etcd.Item)
)

func etcdConfig(cfg *Config) (err error) {
	flatten("/config", cfg.Config)
	for _, h := range cfg.Handlers {
		h.Config = append(h.Config, &etcd.Item{
			Key:     "active",
			Default: "false",
			Desc:    "Whether handler is active - valid values are 'true' or 'false'",
		})
		// Add the handler command to the configurables to manage via the web api
		cmd, _ := json.Marshal(h.Command)
		h.Config = append(h.Config, &etcd.Item{
			Key:     "command",
			Default: string(cmd),
			Desc:    "Handler command",
		})
		flatten(filepath.Join("/handlers", h.Handler), h.Config)
	}

	if err = syncItems(mapping); err != nil {
		glog.Errorf("syncItems(): %s", err)
		return
	}

	go watch(cfg, mapping)

	return
}

func flatten(prefix string, items []*etcd.Item) {
	for _, item := range items {
		path := filepath.Join(prefix, item.Key)
		item.Path = path
		mapping[path] = item
	}
}

func syncItem(item *etcd.Item) (err error) {
	client := etcd.New(*flagEtcd)
	item.Value, err = client.GetDefault(item.Path, item.Default, item.Desc)
	if err != nil {
		return
	}
	return
}

func syncItems(mapping map[string]*etcd.Item) (err error) {
	for _, item := range mapping {
		if err = syncItem(item); err != nil {
			return
		}
	}
	return
}

func watch(cfg *Config, mapping map[string]*etcd.Item) {
	client := etcd.New(*flagEtcd)
	for {
		resp, err := client.Watch("/", 0, true, nil, nil)
		if err != nil {
			// HTTP timeout
			continue
		}
		key, value := resp.Node.Key, resp.Node.Value
		item, ok := mapping[key]
		if !ok {
			// Unknown key
			continue
		}
		glog.Infof("[%s] changed %q -> %q", key, item.Value, value)
		item.Value = resp.Node.Value
		// Handle a changed command
		if filepath.Base(key) == "command" {
			handler := filepath.Base(filepath.Dir(key))
			hc := Handler(handler)
			if hc == nil {
				glog.Warningf("Could not find handler: %q", handler)
				continue
			}
		}
		if item.Changed != nil {
			item.Changed <- true
		}
	}
}
