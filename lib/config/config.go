package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/golang/glog"
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

func ElasticHosts() []string {
	return strings.Split(findValue(Data.Config, "elastichosts"), ",")
}

func ElasticIndex() string {
	return findValue(Data.Config, "elasticindex")
}

func HandlersDir() string {
	return findValue(Data.Config, "handlers")
}

func MysqlDSN() string {
	return findValue(Data.Config, "mysqldsn")
}

func Nsqhttp() string {
	return findValue(Data.Config, "nsqhttp")
}

func Nsqlookuptcp() string {
	return findValue(Data.Config, "nsqlookuptcp")
}

func Nsqlookuphttp() string {
	return findValue(Data.Config, "nsqlookuphttp")
}

func WebListen() string {
	return findValue(Data.Config, "weblisten")
}

func (h HandlerConfig) Active() bool {
	s := findValue(h.Config, "active")
	value, err := strconv.ParseBool(s)
	if err != nil {
		glog.Errorf("Invalid value for /handler/%s/active: %s", h.Handler, err)
		value = false
	}
	return value
}

func (h HandlerConfig) ActiveItem() *etcd.Item {
	return findItem(h.Config, "active")
}

func (h HandlerConfig) Concurrent() int {
	s := findValue(h.Config, "concurrent")
	value, err := strconv.Atoi(s)
	if err != nil {
		glog.Errorf("Invalid value for /handler/%s/concurrent: %s", h.Handler, err)
		value = 0
	}
	return value
}

func (h HandlerConfig) ConcurrentItem() *etcd.Item {
	return findItem(h.Config, "concurrent")
}

func (h HandlerConfig) Consume() string {
	return findValue(h.Config, "consume")
}

func (h HandlerConfig) ConsumeItem() *etcd.Item {
	return findItem(h.Config, "consume")
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
