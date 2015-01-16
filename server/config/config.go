package config

import (
	"flag"
)

var Config = struct {
	Etcd struct {
		Host string
	}
	Mongo struct {
		Host     string
		Database string
	}
	Handlers struct {
		Path string
	}
	Web struct {
		Assets string
		Listen string
	}
}{}

func init() {
	flag.StringVar(&Config.Handlers.Path, "handlers.path", "./", "Directory for handlers")
	flag.StringVar(&Config.Web.Listen, "web.listen", ":8080", "Directory for handlers")
	flag.StringVar(&Config.Web.Assets, "web.assets", "./", "Directory for web assets")
	flag.StringVar(&Config.Etcd.Host, "etcd.host", "127.0.0.1:4001", "Etcd server for configs")
}
