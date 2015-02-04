package config

import (
	"flag"
)

type ConfigType struct {
	Etcd          string
	Mongo         string
	NsqdHTTP      string
	NsqdTCP       string
	NsqLookupdTCP string
	Handlers      string
	WebAssets     string
	WebListen     string
}

var Config = &ConfigType{
	Etcd:          "http://127.0.0.1:4001",
	Mongo:         "mongodb://127.0.0.1:27017/ocular8",
	NsqdHTTP:      "http://127.0.0.1:4161",
	NsqdTCP:       "127.0.0.1:4150",
	NsqLookupdTCP: "127.0.0.1:4160",
	Handlers:      ".",
	WebAssets:     ".",
	WebListen:     ":8080",
}

func init() {
	flag.StringVar(&Config.Handlers, "handlers", Config.Handlers, "Directory for handlers")
	flag.StringVar(&Config.WebAssets, "assets", Config.WebAssets, "Directory for web assets")
	flag.StringVar(&Config.WebListen, "listen", Config.WebListen, "Web listen address")
	flag.StringVar(&Config.Etcd, "etcd", Config.Etcd, "Etcd server for configs")
	flag.StringVar(&Config.NsqdHTTP, "nsqdhttp", Config.NsqdHTTP, "Nsqd server for queue (HTTP)")
	flag.StringVar(&Config.NsqdTCP, "nsqdtcp", Config.NsqdTCP, "Nsqd server for queue (TCP)")
	flag.StringVar(&Config.NsqLookupdTCP, "nsqlookupdtcp", Config.NsqLookupdTCP, "Nsq lookupd server for queue (TCP)")
	flag.StringVar(&Config.Mongo, "mongo", Config.Mongo, "Mongo server")
}

func Parse() {
	flag.Parse()
}
