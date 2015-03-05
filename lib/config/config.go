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

func AssetsDir() string {
	return findValue(Data.Config, "assets")
}

func Etcd() string {
	return findValue(Data.Config, "etcd")
}

func Elastic() string {
	return findValue(Data.Config, "elastic")
}

func HandlersDir() string {
	return findValue(Data.Config, "handlers")
}

func Mongo() string {
	return findValue(Data.Config, "mongo")
}

func Nsqhttp() string {
	return findValue(Data.Config, "nsqhttp")
}

func WebListen() string {
	return findValue(Data.Config, "weblisten")
}

func (h HandlerConfig) Consume() string {
	return findValue(h.Config, "consume")
}

func (h HandlerConfig) ConsumeItem() *etcd.Item {
	return findItem(h.Config, "frequency")
}

func (h HandlerConfig) Frequency() time.Duration {
	dur, err := time.ParseDuration(findValue(h.Config, "frequency"))
	if err != nil {
		glog.Errorf("Invalid frequency for %q: %s", h.Handler, err)
		return time.Duration(0)
	}
	return dur
}

func (h HandlerConfig) FrequencyItem() *etcd.Item {
	return findItem(h.Config, "frequency")
}

func (h HandlerConfig) IsConsumer() bool {
	return h.ConsumeItem() != nil
}

func (h HandlerConfig) IsProducer() bool {
	return h.FrequencyItem() != nil
}

func findItem(items []etcd.Item, key string) *etcd.Item {
	for i := range items {
		if items[i].Key == key {
			return &items[i]
		}
	}
	return nil
}

func findValue(items []etcd.Item, key string) string {
	if item := findItem(items, key); item != nil {
		return item.Value
	}
	return ""
}
