package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/server/web"
	"github.com/golang/glog"
)

func main() {
	if err := config.Parse(); err != nil {
		glog.Fatalf("config.Parse(): %s", err)
	}

	if *doPrime {
		if err := prime(config.Mongo()); err != nil {
			glog.Fatalf("[prime] %s", err)
		}
		return
	}

	// Listen for signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	// Start web frontend
	go func(addr, dir, mongo string) {
		if err := web.Mongo(mongo); err != nil {
			glog.Fatalf("web.Mongo(%s): %s", mongo, err)
		}
		if err := http.ListenAndServe(addr, web.Handler(dir)); err != nil {
			glog.Fatalf("http.ListenAndServe(%s): %s", addr, err)
		}
	}(config.WebListen(), config.AssetsDir(), config.Mongo())

	// Start handlers
	stopChan := make(chan bool)
	if err := Handlers(config.Data.Handlers, stopChan); err != nil {
		glog.Fatalf("Handlers(): %s", err)
	}

	// Idle
	glog.Infof("Running")

	// Catch kill, clean up, exit
	s := <-signalChan
	glog.Info("Caught signal:", s)
	stopChan <- true
	glog.Info("Cleaning up")
	web.Close()
	<-stopChan
	glog.Info("Exiting")
}
