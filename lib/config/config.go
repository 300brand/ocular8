package config

import (
	"flag"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/etcd"
	goetcd "github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

type Config struct {
	Config   []*etcd.Item
	Handlers []*HandlerConfig
}

type HandlerConfig struct {
	Handler string
	Command []string
	Config  []*etcd.Item
}

var Data Config = Config{
	Config: []*etcd.Item{
		&etcd.Item{
			Key:     "etcd",
			Default: "http://localhost:4001",
			Desc:    "Etcd HTTP address",
		},
		&etcd.Item{
			Key:     "elastichosts",
			Default: "localhost",
			Desc:    "ElasticSearch hosts, comma-separated",
		},
		&etcd.Item{
			Key:     "elasticindex",
			Default: "ocular8",
			Desc:    "ElasticSearch index name",
		},
		&etcd.Item{
			Key:     "mysqldsn",
			Default: "root@tcp(localhost:3307)/ocular8",
			Desc:    "DSN Connection string to MySQL",
		},
		&etcd.Item{
			Key:     "mysqlobj",
			Default: `{ "user": "root", "password": "", "host": "localhost", "port": 3307, "database": "ocular8" }`,
			Desc:    "Connection object to MySQL",
		},
		&etcd.Item{
			Key:     "nsqhttp",
			Default: "http://localhost:4151",
			Desc:    "NSQd HTTP address",
		},
		&etcd.Item{
			Key:     "nsqtcp",
			Default: "localhost:4150",
			Desc:    "NSQd TCP address",
		},
		&etcd.Item{
			Key:     "nsqlookuphttp",
			Default: "localhost:4161",
			Desc:    "NSQLookupd HTTP address",
		},
		&etcd.Item{
			Key:     "nsqlookuptcp",
			Default: "localhost:4160",
			Desc:    "NSQLookupd TCP address",
		},
		&etcd.Item{
			Key:     "handlers",
			Default: "./handlers",
			Desc:    "Path to handlers directory",
		},
		&etcd.Item{
			Key:     "assets",
			Default: "./assets",
			Desc:    "Path to web assets directory",
		},
		&etcd.Item{
			Key:     "weblisten",
			Default: ":8080",
			Desc:    "Web listen address",
		},
	},
	Handlers: []*HandlerConfig{
		&HandlerConfig{
			Handler: "download-entry",
			Command: []string{"./download-entry", "-etcd", "{{ .Etcd }}", "{{ .Data }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "consume",
					Default: "entry.id.download",
					Desc:    "Topic to watch for new items",
				},
				&etcd.Item{
					Key:     "concurrent",
					Default: "1",
					Desc:    "How many instances can run concurrently",
				},
				&etcd.Item{
					Key:     "sizelimit",
					Default: "2097152",
					Desc:    "Maximum number of bytes to download for any HTML page",
				},
			},
		},
		&HandlerConfig{
			Handler: "download-feed",
			Command: []string{"./download-feed", "-etcd", "{{ .Etcd }}", "{{ .Data }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "consume",
					Default: "feed.id.download",
					Desc:    "Topic to watch for new items",
				},
				&etcd.Item{
					Key:     "concurrent",
					Default: "1",
					Desc:    "How many instances can run concurrently",
				},
			},
		},
		&HandlerConfig{
			Handler: "elastic-save",
			Command: []string{"./elastic-save", "-etcd", "{{ .Etcd }}", "{{ .Data }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "consume",
					Default: "article.id.elastic",
					Desc:    "Topic to watch for new items",
				},
				&etcd.Item{
					Key:     "concurrent",
					Default: "1",
					Desc:    "How many instances can run concurrently",
				},
			},
		},
		&HandlerConfig{
			Handler: "enqueue-feeds",
			Command: []string{"./enqueue-feeds", "-etcd", "{{ .Etcd }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "frequency",
					Default: "30s",
					Desc:    "How often to run, parsed by time.ParseDuration",
				},
				&etcd.Item{
					Key:     "limit",
					Default: "25",
					Desc:    "How many feeds per batch",
				},
				&etcd.Item{
					Key:     "threshold",
					Default: "100",
					Desc:    "Max number of feeds allowed in queue",
				},
			},
		},
		&HandlerConfig{
			Handler: "extract-goose",
			Command: []string{"./extract-goose", "-etcd", "{{ .Etcd }}", "{{ .Data }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "consume",
					Default: "article.id.extract.goose",
					Desc:    "Topic to watch for new items",
				},
				&etcd.Item{
					Key:     "concurrent",
					Default: "1",
					Desc:    "How many instances can run concurrently",
				},
			},
		},
		&HandlerConfig{
			Handler: "extract-xpath",
			Command: []string{"./extract-xpath", "-etcd", "{{ .Etcd }}", "{{ .Data }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "consume",
					Default: "article.id.extract.xpath",
					Desc:    "Topic to watch for new items",
				},
				&etcd.Item{
					Key:     "concurrent",
					Default: "1",
					Desc:    "How many instances can run concurrently",
				},
			},
		},
		&HandlerConfig{
			Handler: "metabase",
			Command: []string{"./metabase", "-etcd", "{{ .Etcd }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "frequency",
					Default: "30s",
					Desc:    "How often to run, parsed by time.ParseDuration",
				},
				&etcd.Item{
					Key:     "apikey",
					Default: "",
					Desc:    "Metabase API Key",
				},
				&etcd.Item{
					Key:     "sequencereset",
					Default: "48h",
					Desc:    "How long to wait before cutting losses and resetting sequenceId. Used in the event of power/network loss",
				},
			},
		},
		&HandlerConfig{
			Handler: "metabase-ln",
			Command: []string{"./metabase-ln", "-lexisnexis", "-etcd", "{{ .Etcd }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "frequency",
					Default: "30s",
					Desc:    "How often to run, parsed by time.ParseDuration",
				},
				&etcd.Item{
					Key:     "apikey",
					Default: "",
					Desc:    "Metabase API Key",
				},
				&etcd.Item{
					Key:     "sequencereset",
					Default: "48h",
					Desc:    "How long to wait before cutting losses and resetting sequenceId. Used in the event of power/network loss",
				},
			},
		},
		&HandlerConfig{
			Handler: "recount-articles",
			Command: []string{"./recount-articles", "-etcd", "{{ .Etcd }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "frequency",
					Default: "5m",
					Desc:    "How often to run, parsed by time.ParseDuration",
				},
			},
		},
		&HandlerConfig{
			Handler: "resolve-url",
			Command: []string{"./resolve-url", "-etcd", "{{ .Etcd }}", "{{ .Data }}"},
			Config: []*etcd.Item{
				&etcd.Item{
					Key:     "consume",
					Default: "article.id.resolve.url",
					Desc:    "Topic to watch for new items",
				},
				&etcd.Item{
					Key:     "concurrent",
					Default: "1",
					Desc:    "How many instances can run concurrently",
				},
			},
		},
	},
}

var client *etcd.Client

func Parse() (err error) {
	flag.Parse()
	client = etcd.New(*flagEtcd)
	return
}

func Sync() (err error) {
	err = etcdConfig(&Data)
	if err != nil {
		glog.Errorf("etcdConfig(): %s", err)
	}
	return
}

func AssetsDir() string {
	return value("/config/assets")
}

func Etcd() string {
	return value("/config/etcd")
}

func ElasticHosts() []string {
	return strings.Split(value("/config/elastichosts"), ",")
}

func ElasticIndex() string {
	return value("/config/elasticindex")
}

func HandlersDir() string {
	return value("/config/handlers")
}

func MysqlDSN() string {
	return value("/config/mysqldsn")
}

func Nsqhttp() string {
	return value("/config/nsqhttp")
}

func Nsqlookuptcp() string {
	return value("/config/nsqlookuptcp")
}

func Nsqlookuphttp() string {
	return value("/config/nsqlookuphttp")
}

func WebListen() string {
	return value("/config/weblisten")
}

func Handler(name string) (hc *HandlerConfig) {
	for _, hc = range Data.Handlers {
		if hc.Handler == name {
			return
		}
	}
	hc = nil
	return
}

func (h HandlerConfig) Active() bool {
	value, err := strconv.ParseBool(h.value("active"))
	if err != nil {
		glog.Errorf("Invalid value for /handler/%s/active: %s", h.Handler, err)
		value = false
	}
	return value
}

func (h HandlerConfig) ActiveItem() *etcd.Item {
	return h.item("active")
}

func (h HandlerConfig) Concurrent() int {
	value, err := strconv.Atoi(h.value("concurrent"))
	if err != nil {
		glog.Errorf("Invalid value for /handler/%s/concurrent: %s", h.Handler, err)
		value = 0
	}
	return value
}

func (h HandlerConfig) ConcurrentItem() *etcd.Item {
	return h.item("concurrent")
}

func (h HandlerConfig) Consume() string {
	return h.value("consume")
}

func (h HandlerConfig) ConsumeItem() *etcd.Item {
	return h.item("consume")
}

func (h HandlerConfig) Frequency() time.Duration {
	dur, err := time.ParseDuration(h.value("frequency"))
	if err != nil {
		glog.Errorf("Invalid frequency for %q: %s", h.Handler, err)
		return time.Duration(0)
	}
	return dur
}

func (h HandlerConfig) FrequencyItem() *etcd.Item {
	return h.item("frequency")
}

func (h HandlerConfig) IsConsumer() bool {
	return h.ConsumeItem() != nil
}

func (h HandlerConfig) IsProducer() bool {
	return h.FrequencyItem() != nil
}

func (h HandlerConfig) item(key string) *etcd.Item {
	return item(filepath.Join("/handlers", h.Handler, key))
}

func (h HandlerConfig) value(key string) string {
	return value(filepath.Join("/handlers", h.Handler, key))
}

func item(key string) (item *etcd.Item) {
	if client == nil {
		glog.Fatal("Etcd client not initialized, did you run config.Parse()?")
	}
	resp, err := client.Get(key, false, false)
	if e, ok := err.(*goetcd.EtcdError); ok {
		switch e.ErrorCode {
		case 100:
			return nil
		}
	}
	if err != nil {
		glog.Fatalf("etcd.Client.Get(%s): %s", key, err)
	}
	item = &etcd.Item{
		Key:     filepath.Base(key),
		Path:    key,
		Default: "",
		Desc:    "",
		Value:   resp.Node.Value,
	}
	return
}

func value(key string) (value string) {
	if i := item(key); i != nil {
		value = i.Value
	}
	return
}
