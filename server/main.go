package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/server/web"
	"github.com/golang/glog"
)

var (
	noweb     = flag.Bool("noweb", false, "Don't start web server")
	nohandler = flag.Bool("nohandler", false, "Don't start handlers")
)

func startHandlers(handlerCfg []config.HandlerConfig, stop chan bool) {
	if err := Handlers(handlerCfg, stop); err != nil {
		glog.Fatalf("Handlers(): %s", err)
	}
	glog.Info("Waiting for handler cleanup")
}

func startWeb(addr, dir string, stop chan bool) {
	go func() {
		if err := http.ListenAndServe(addr, web.Handler(dir)); err != nil {
			glog.Fatalf("http.ListenAndServe(%s): %s", addr, err)
		}
	}()
	<-stop
	glog.Info("Waiting for web cleanup")
	web.Close()
	stop <- true
}

func main() {
	if err := config.Parse(); err != nil {
		glog.Fatalf("config.Parse(): %s", err)
	}

	if *doPrime {
		if err := prime(config.ElasticHosts(), config.ElasticIndex(), config.MysqlDSN()); err != nil {
			glog.Fatalf("[prime] %s", err)
		}
		return
	}

	// Listen for signals
	stopMux := make([]chan bool, 0, 2)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	// Start web frontend
	if !*noweb {
		ch := make(chan bool)
		stopMux = append(stopMux, ch)
		go startWeb(config.WebListen(), config.AssetsDir(), config.Mongo(), ch)
	}

	// Start handlers
	if !*nohandler {
		ch := make(chan bool)
		stopMux = append(stopMux, ch)
		go startHandlers(config.Data.Handlers, ch)
	}

	// Idle
	glog.Infof("Running")

	// Catch kill, clean up, exit
	s := <-signalChan
	glog.Info("Caught signal:", s)

	for _, ch := range stopMux {
		ch <- true
		<-ch
	}

	glog.Info("Exiting")
}
