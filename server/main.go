package main

import (
	"flag"
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
		if err := h.Run(nil); err != nil {
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
	flag.Parse()

	glog.Infof("Config %+v", config.Config)

	stopPollingChan := make(chan bool)
	stopConsumersChan := make(chan bool)

	// Listen for signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	handlerConfigs, err := handler.ParseConfigs(config.Config.Handlers.Path)
	if err != nil {
		glog.Errorf("parseConfigs: %s", err)
	}

	go startPolling(stopPollingChan, handlerConfigs)
	go startConsumers(stopConsumersChan, handlerConfigs)

	go func(addr string) {
		if err := http.ListenAndServe(addr, web.Handler()); err != nil {
			glog.Fatalf("http.ListenAndServe(%s): %s", addr, err)
		}
	}(config.Config.Web.Listen)

	glog.Infof("Running")

	// Catch kill
	s := <-signalChan
	glog.Info("Caught signsl:", s)

	// Cleanup
	stopPollingChan <- true
	stopConsumersChan <- true

	glog.Info("Exiting")

	<-stopPollingChan
	<-stopConsumersChan
}
