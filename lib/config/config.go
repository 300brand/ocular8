package config

import (
	"github.com/300brand/ocular8/lib/etcd"
	"github.com/golang/glog"
	"time"
)

type Config struct {
	Config   []etcd.Item
	Handlers []HandlerConfig
}

type HandlerConfig struct {
	Handler string
	Command []string
	Config  []etcd.Item
}

var Data Config

func (c Config) AssetsDir() string {
	return find(c.Config, "assets")
}

func (c Config) Etcd() string {
	return find(c.Config, "etcd")
}

func (c Config) Elastic() string {
	return find(c.Config, "elastic")
}

func (c Config) HandlersDir() string {
	return find(c.Config, "handlers")
}

func (c Config) Mongo() string {
	return find(c.Config, "mongo")
}

func (c Config) Nsqhttp() string {
	return find(c.Config, "nsqhttp")
}

func (c Config) WebListen() string {
	return find(c.Config, "weblisten")
}

func (h HandlerConfig) ConsumeTopic() string {
	return find(h.Config, "consume")
}

func (h HandlerConfig) Frequency() time.Duration {
	dur, err := time.ParseDuration(find(h.Config, "frequency"))
	if err != nil {
		glog.Errorf("Invalid frequency for %q: %s", h.Handler, err)
		return time.Duration(0)
	}
	return dur
}

func (h HandlerConfig) IsConsumer() bool {
	return h.ConsumeTopic() != ""
}

func (h HandlerConfig) IsProducer() bool {
	return h.Frequency() > time.Duration(0)
}

func find(items []etcd.Item, key string) string {
	for i := range items {
		if items[i].Key == key {
			return items[i].Value
		}
	}
	return ""
}
