package main

import (
	"path/filepath"
	"time"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/lib/handler"
	"github.com/golang/glog"
)

const MinFrequency = 10 * time.Second

func Handlers(handlers []config.HandlerConfig, stopChan chan bool) (err error) {
	glog.Infof("Handlers: %+v", handlers)

	stopChans := make([]chan bool, 0, len(handlers))
	for i := range handlers {
		cfg := &handlers[i]
		if cfg.IsConsumer() {
			ch := make(chan bool)
			go Consumer(cfg, ch)
			stopChans = append(stopChans, ch)
		}
		if cfg.IsProducer() {
			ch := make(chan bool)
			go Producer(cfg, ch)
			stopChans = append(stopChans, ch)
		}
	}

	go func() {
		<-stopChan
		glog.Infof("Shutting down handlers")
		for _, ch := range stopChans {
			ch <- true
			<-ch
		}
		stopChan <- true
	}()
	return
}

func Consumer(cfg *config.HandlerConfig, stopChan chan bool) {
	glog.Infof("[%s] Starting consumer", cfg.Handler)
	<-stopChan
	glog.Infof("[%s] Stopping consumer", cfg.Handler)
	stopChan <- true
}

func Producer(p *config.HandlerConfig, stopChan chan bool) {
	glog.Infof("[%s] Starting producer", p.Handler)

	abs, err := filepath.Abs(config.HandlersDir())
	if err != nil {
		glog.Fatalf("Could not determine absolute path for handlers: %s", err)
	}
	h := handler.New(filepath.Join(abs, p.Handler), p.Command)

	active := p.ActiveItem()
	active.Changed = make(chan bool)
	freq := p.FrequencyItem()
	freq.Changed = make(chan bool)

Loop:
	for {
		if !p.Active() {
			glog.Infof("[%s] Waiting to become active again", p.Handler)
			select {
			case <-active.Changed:
				continue
			case <-stopChan:
				glog.Infof("[%s] Got stop", p.Handler)
				break Loop
			}
		}

		if f := p.Frequency(); f < MinFrequency {
			glog.Errorf("[%s] Current frequency (%s) shorter than minimum (%s). Backing off until changed.", p.Handler, f, MinFrequency)
			select {
			case <-freq.Changed:
				continue
			case <-stopChan:
				glog.Infof("[%s] Got stop", p.Handler)
				break Loop
			}
			continue
		}

		select {
		case <-active.Changed:
			glog.Infof("[%s] Active state changed to %v", p.Handler, p.Active())
		case <-time.After(p.Frequency()):
			glog.Infof("[%s] Running", p.Handler)
			if err := h.Run(config.Etcd(), ""); err != nil {
				glog.Errorf("[%s] %s - %s", p.Handler, h.ParsedCmd(config.Etcd(), ""), err)
			}
		case <-freq.Changed:
			glog.Infof("[%s] Frequency changed to %s", p.Handler, p.Frequency())
		case <-stopChan:
			glog.Infof("[%s] Got stop", p.Handler)
			break Loop
		}
	}
	glog.Infof("[%s] Stopping producer", p.Handler)
	stopChan <- true
}

// func consumerHandler(h handler.Handler) (f nsq.HandlerFunc) {
// 	return func(msg *nsq.Message) (err error) {
// 		glog.Infof("%s: %s %d", h.Command[0], time.Unix(0, msg.Timestamp), msg.Attempts)
// 		buf := bytes.NewBuffer(msg.Body)
// 		return h.Run(config.Etcd(), buf.String())
// 	}
// }

// func poll(stopChan chan bool, h handler.Handler) {
// 	duration := time.Hour / time.Duration(h.Frequency)
// 	glog.Infof("Polling %s every %s", h.Name, duration)
// 	for {
// 		glog.Infof("Polling %s", h.Name)
// 		if err := h.Run(time.Now().Format(time.RFC3339Nano)); err != nil {
// 			glog.Errorf("Poll run %s: %s", h.Name, err)
// 		}
// 		select {
// 		case <-time.After(duration):
// 		case <-stopChan:
// 			glog.Infof("Stop polling: %s", h.Name)
// 			return
// 		}
// 	}
// }

// func startConsumers(stopChan chan bool, configs []handler.Handler) {
// 	glog.Infof("Consumers: Starting")

// 	consumers := make([]*nsq.Consumer, 0, len(configs))
// 	for _, h := range configs {
// 		if h.NSQ.Consume.Topic == "" {
// 			continue
// 		}
// 		nsqConfig := nsq.NewConfig()
// 		nsqConfig.Set("client_id", h.Name)
// 		topic, channel := h.NSQ.Consume.Topic, h.NSQ.Consume.Channel
// 		glog.Infof("Setting up consumer for %s -> %s", topic, channel)
// 		consumer, err := nsq.NewConsumer(topic, channel, nsqConfig)
// 		if err != nil {
// 			glog.Fatalf("nsq.NewConsumer(%s, %s, config): %s", topic, channel, err)
// 		}
// 		concurrency := 1
// 		if c := h.NSQ.Consume.Concurrent; c > concurrency {
// 			concurrency = c
// 		}
// 		consumer.ChangeMaxInFlight(concurrency)
// 		consumer.AddConcurrentHandlers(consumerHandler(h), concurrency)
// 		if err := consumer.ConnectToNSQD(config.Config.NsqdTCP); err != nil {
// 			glog.Fatalf("nsq.ConnectToNSQD(%s): %s", config.Config.NsqdTCP, err)
// 		}
// 		consumers = append(consumers, consumer)
// 	}

// 	<-stopChan
// 	glog.Infof("Consumers: Exiting")
// 	for _, c := range consumers {
// 		c.Stop()
// 		<-c.StopChan
// 	}
// 	stopChan <- true
// }

// func startPolling(stopChan chan bool, configs []handler.Handler) {
// 	glog.Infof("Polling: Starting")

// 	chans := make([]chan bool, 0, len(configs))
// 	for _, h := range configs {
// 		if h.Frequency <= 0 {
// 			continue
// 		}
// 		ch := make(chan bool)
// 		go poll(ch, h)
// 		chans = append(chans, ch)
// 	}

// 	<-stopChan
// 	glog.Infof("Polling: Exiting")
// 	for _, ch := range chans {
// 		ch <- true
// 	}
// 	stopChan <- true
// }
