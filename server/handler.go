package main

import (
	"time"

	"github.com/300brand/ocular8/server/handler"
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
	glog.Infof("Consumers: Starting")
	<-stopChan
	glog.Infof("Consumers: Exiting")
	stopChan <- true
}

func startPolling(stopChan chan bool, configs []handler.Handler) {
	glog.Infof("Polling: Starting")

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
	glog.Infof("Polling: Exiting")
	for _, ch := range chans {
		ch <- true
	}
	stopChan <- true
}
