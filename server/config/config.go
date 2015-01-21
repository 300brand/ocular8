package config

import (
	"flag"
)

type ConfigType struct {
	Etcd      string
	Mongo     string
	Nsqd      string
	Handlers  string
	WebAssets string
	WebListen string
}

var Config = &ConfigType{
	"http://127.0.0.1:4001",
	"mongodb://127.0.0.1:27017/test",
	"127.0.0.1:9200",
	".",
	".",
	":8080",
}

func init() {
	flag.StringVar(&Config.Handlers, "handlers", Config.Handlers, "Directory for handlers")
	flag.StringVar(&Config.WebAssets, "assets", Config.WebAssets, "Directory for web assets")
	flag.StringVar(&Config.WebListen, "listen", Config.WebListen, "Web listen address")
	flag.StringVar(&Config.Etcd, "etcd", Config.Etcd, "Etcd server for configs")
	flag.StringVar(&Config.Nsqd, "nsqd", Config.Nsqd, "Nsqd server for queue")
	flag.StringVar(&Config.Mongo, "mongo", Config.Mongo, "Mongo server")
}

func Parse() {
	flag.Parse()
}
