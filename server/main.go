package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/300brand/ocular8/server/config"
	"github.com/300brand/ocular8/server/web"
	"github.com/golang/glog"
)

func main() {
	config.Parse()

	if *doPrime {
		if err := prime(config.Config.Mongo); err != nil {
			glog.Fatalf("[prime] %s", err)
		}
		return
	}

	glog.Infof("Config %+v", config.Config)

	// Listen for signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	// Start web frontend
	go func(addr, dir, mongo string) {
		if err := web.Mongo(mongo); err != nil {
			glog.Fatalf("web.Mongo(%s): %s", config.Config.Mongo, err)
		}
		if err := http.ListenAndServe(addr, web.Handler(dir)); err != nil {
			glog.Fatalf("http.ListenAndServe(%s): %s", addr, err)
		}
	}(config.Config.WebListen, config.Config.WebAssets, config.Config.Mongo)

	// Start handlers
	stopChan, err := setupHandlers(config.Config.Handlers)
	if err != nil {
		glog.Fatalf("setupHandlers(%s): %s", config.Config.Handlers, err)
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
