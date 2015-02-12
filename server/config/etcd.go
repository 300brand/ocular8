package config

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"path/filepath"
)

const (
	ConfigBase  = "/config"
	HandlerBase = "/handler"
)

func FromEtcd(addr string) (err error) {
	sets := map[string]*string{
		"Mongo":         &Config.Mongo,
		"NsqdHTTP":      &Config.NsqdHTTP,
		"NsqdTCP":       &Config.NsqdTCP,
		"NsqLookupdTCP": &Config.NsqLookupdTCP,
		"Handlers":      &Config.Handlers,
		"WebAssets":     &Config.WebAssets,
		"WebListen":     &Config.WebListen,
	}

	client := etcd.NewClient([]string{addr})
	for k, v := range sets {
		if *v, err = getKey(client, filepath.Join(ConfigBase, k), *v); err != nil {
			return
		}
	}

	go watch(client, ConfigBase)
	return
}

func watch(client *etcd.Client, prefix string) {
	respChan := make(chan *etcd.Response)
	stopChan := make(chan bool)
	go client.Watch(prefix, 0, false, respChan, stopChan)

	for resp := range respChan {
		glog.Infof("%+v", resp)
	}

	return
}

func getKey(client *etcd.Client, key, defaultValue string) (value string, err error) {
	value = defaultValue
	resp, err := client.Get(key, false, false)
	if etcderr, ok := err.(*etcd.EtcdError); ok {
		// 100: Key not found (/config) [22]
		if etcderr.ErrorCode != 100 {
			return
		}
		_, err = client.Set(key, defaultValue, 0)
		return
	}
	value = resp.Node.Value
	return
}
