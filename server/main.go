package main

import (
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/300brand/ocular8/server/config"
	"github.com/300brand/ocular8/server/handler"
	"github.com/300brand/ocular8/server/web"
	"github.com/golang/glog"
)

func poll(stopChan chan bool, h handler.Handler) {
	duration := time.Hour / time.Duration(h.Frequency)
	glog.Infof("Polling %s every %s", h.Name, duration)
	for {
		glog.Infof("Polling %s", h.Name)
		if err := h.Run(time.Now().Format(time.RFC3339Nano)); err != nil {
			glog.Errorf("Poll run %s: %s", h.Name, err)
		}
		select {
		case <-time.After(duration):
		case <-stopChan:
			glog.Infof("Stop polling: %s", h.Name)
			return
		}
	}
}

func setupHandlers(dir string) (stopChan chan bool, err error) {
	handlerConfigs, err := handler.ParseConfigs(dir)
	if err != nil {
		glog.Errorf("parseConfigs: %s", err)
		return
	}

	stopChan, poll, consume := make(chan bool), make(chan bool), make(chan bool)

	go func(ch chan bool) {
		<-ch
		poll <- true
		consume <- true
		<-poll
		<-consume
		ch <- true
	}(stopChan)

	go startPolling(poll, handlerConfigs)
	go startConsumers(consume, handlerConfigs)

	return
}

func startConsumers(stopChan chan bool, configs []handler.Handler) {
	glog.V(1).Infof("Consumers: Starting")
	<-stopChan
	glog.V(1).Infof("Consumers: Exiting")
	stopChan <- true
}

func startPolling(stopChan chan bool, configs []handler.Handler) {
	glog.V(1).Infof("Polling: Starting")

	chans := make([]chan bool, 0, len(configs))
	for _, h := range configs {
		if h.Frequency <= 0 {
			continue
		}
		ch := make(chan bool)
		go poll(ch, h)
		chans = append(chans, ch)
	}

	<-stopChan
	glog.V(1).Infof("Polling: Exiting")
	for _, ch := range chans {
		ch <- true
	}
	stopChan <- true
}

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

	if err := web.Mongo(config.Config.Mongo); err != nil {
		glog.Fatalf("web.Mongo(%s): %s", config.Config.Mongo, err)
	}

	go func(addr, dir string) {
		if err := http.ListenAndServe(addr, web.Handler(dir)); err != nil {
			glog.Fatalf("http.ListenAndServe(%s): %s", addr, err)
		}
	}(config.Config.WebListen, config.Config.WebAssets)

	stopChan, err := setupHandlers(config.Config.Handlers)
	if err != nil {
		glog.Fatalf("setupHandlers(%s): %s", config.Config.Handlers, err)
	}

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
