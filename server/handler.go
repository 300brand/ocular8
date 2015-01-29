package main

import (
	"bytes"
	"github.com/300brand/ocular8/server/config"
	"time"

	"github.com/300brand/ocular8/server/handler"
	"github.com/bitly/go-nsq"
	"github.com/golang/glog"
)

func consumerHandler(h handler.Handler) (f nsq.HandlerFunc) {
	return func(msg *nsq.Message) (err error) {
		glog.Infof("%s: %q %s %d %s", h.Name, msg.Body, time.Unix(0, msg.Timestamp), msg.Attempts)
		buf := bytes.NewBuffer(msg.Body)
		return h.Run(buf.String())
	}
}

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

	consumers := make([]*nsq.Consumer, 0, len(configs))
	for _, h := range configs {
		if h.NSQ.Consume.Topic == "" {
			continue
		}
		nsqConfig := nsq.NewConfig()
		nsqConfig.Set("client_id", h.Name)
		topic, channel := h.NSQ.Consume.Topic, h.NSQ.Consume.Channel
		glog.Infof("Setting up consumer for %s -> %s", topic, channel)
		consumer, err := nsq.NewConsumer(topic, channel, nsqConfig)
		if err != nil {
			glog.Fatalf("nsq.NewConsumer(%s, %s, config): %s", topic, channel, err)
		}
		concurrency := 1
		if c := h.NSQ.Consume.Concurrent; c > concurrency {
			concurrency = c
		}
		consumer.AddConcurrentHandlers(consumerHandler(h), concurrency)
		if err := consumer.ConnectToNSQD(config.Config.NsqdTCP); err != nil {
			glog.Fatalf("nsq.ConnectToNSQD(%s): %s", config.Config.NsqdTCP, err)
		}
		consumers = append(consumers, consumer)
	}

	<-stopChan
	glog.Infof("Consumers: Exiting")
	for _, c := range consumers {
		c.Stop()
		<-c.StopChan
	}
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
