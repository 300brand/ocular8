package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/300brand/ocular8/lib/etcd"
	goetcd "github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

var (
	flagEtcd       = flag.String("etcd", "http://localhost:4001", "Etcd server address")
	flagConfigjson = flag.String("configjson", "./config.json", "Path to config.json. Only necessary on initial setup.")
)

func Parse() (err error) {
	flag.Parse()

	client := etcd.New(*flagEtcd)
	configjson := *flagConfigjson
	resp, err := client.Get("/config/configjson", false, false)
	if e, ok := err.(*goetcd.EtcdError); ok && e.ErrorCode == 100 {
		err = nil
	}
	if err != nil {
		return
	}
	if resp != nil {
		configjson = resp.Node.Value
	}

	f, err := os.Open(configjson)
	if err != nil {
		return
	}
	defer f.Close()
	if err = json.NewDecoder(f).Decode(&Data); err != nil {
		return
	}

	if err = setDefaults(client); err != nil {
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
		prefix := "/handlers/" + Data.Handlers[h].Handler
		for i := range Data.Handlers[h].Config {
			if err = setItem(c, prefix, &Data.Handlers[h].Config[i], false); err != nil {
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
