package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/golang/glog"
)

var (
	flagEtcd       = flag.String("etcd", "http://localhost:4001", "Etcd server address")
	flagConfigjson = flag.String("configjson", "", "Path to config.json. Only necessary on initial setup.")
)

func Parse() (err error) {
	flag.Parse()

	glog.Infof("flagEtcd: %s", *flagEtcd)
	client := etcd.New(*flagEtcd)

	if f := *flagConfigjson; f != "" {
		if err = fromFile(f); err != nil {
			return
		}
		if err = setDefaults(client); err != nil {
			glog.Errorf("setDefaults: %s", err)
			return
		}
	} else {
		watchAll(client)
	}

	return
}

func fromFile(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	if err = json.NewDecoder(f).Decode(&Data); err != nil {
		return
	}
	return
}

func setDefaults(c *etcd.Client) (err error) {
	for i := range Data.Config {
		if err = setItem(c, "/config", &Data.Config[i], true); err != nil {
			return
		}
	}
	for h := range Data.Handlers {
		// Ensure sure the active key; initialize as deactivated
		Data.Handlers[h].Config = append(Data.Handlers[h].Config, etcd.Item{
			Key:     "active",
			Default: "false",
			Desc:    "Whether handler is active - valid values are 'true' or 'false'",
		})
		cmd, _ := json.Marshal(Data.Handlers[h].Command)
		Data.Handlers[h].Config = append(Data.Handlers[h].Config, etcd.Item{
			Key:     "command",
			Default: string(cmd),
			Desc:    "Handler command",
		})
		prefix := "/handlers/" + Data.Handlers[h].Handler
		configs := Data.Handlers[h].Config
		for i := range configs {
			if err = setItem(c, prefix, &configs[i], true); err != nil {
				return
			}
		}
	}
	return
}

func setItem(c *etcd.Client, prefix string, item *etcd.Item, watch bool) (err error) {
	key := filepath.Join(prefix, item.Key)
	item.Value, err = c.GetDefault(key, item.Default, item.Desc)
	if err != nil {
		return
	}
	if watch {
		go watchItem(c, key, item)
	}
	return
}

func watchAll(c *etcd.Client) {
	list, err := c.GetList()
	if err != nil {
		glog.Fatalf("etcd.Client.GetList(): %s", err)
	}
	for _, item := range list {
		go watchItem(c, item.Key, item)
	}
}

func watchItem(c *etcd.Client, key string, item *etcd.Item) {
	for {
		resp, err := c.Client.Watch(key, 0, false, nil, nil)
		if err != nil {
			// HTTP timeout
			continue
		}
		glog.Infof("[%s] changed %q -> %q", key, resp.PrevNode.Value, resp.Node.Value)
		item.Value = resp.Node.Value
		if item.Changed != nil {
			item.Changed <- true
		}
	}
}
