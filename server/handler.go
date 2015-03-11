package main

import (
	"bytes"
	"path/filepath"
	"time"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/lib/handler"
	"github.com/bitly/go-nsq"
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

func Consumer(c *config.HandlerConfig, stopChan chan bool) {
	glog.Infof("[%s] Starting consumer", c.Handler)

	abs, err := filepath.Abs(config.HandlersDir())
	if err != nil {
		glog.Fatalf("Could not determine absolute path for handlers: %s", err)
	}
	h := handler.New(filepath.Join(abs, c.Handler), c.Command)

	active := c.ActiveItem()
	active.Changed = make(chan bool)
	concurrent := c.ConcurrentItem()
	concurrent.Changed = make(chan bool)
	consume := c.ConsumeItem()
	consume.Changed = make(chan bool)

	nsqConfig := nsq.NewConfig()
	nsqConfig.ClientID = c.Handler
	var consumer *nsq.Consumer
	setupConsumer := func() {
		var err error
		glog.Infof("[%s] Setting up consumer; Concurrent:%d; Topic: %s", c.Handler, c.Concurrent(), c.Consume())
		nsqConfig.MaxInFlight = c.Concurrent()
		consumer, err = nsq.NewConsumer(c.Consume(), c.Handler, nsqConfig)
		if err != nil {
			glog.Fatalf("nsq.NewConsumer(%q, %q, nsqConfig): %s", c.Consume(), c.Handler, err)
		}
		consumer.AddConcurrentHandlers(consumerHandler(h), c.Concurrent())
		if err = consumer.ConnectToNSQLookupd(config.Nsqlookuphttp()); err != nil {
			glog.Fatalf("consumer.ConnectToNSQLookupd(%q): %s", config.Nsqlookuptcp(), err)
		}
	}

Loop:
	for {
		if !c.Active() {
			glog.Infof("[%s] Waiting to become active again", c.Handler)
			select {
			case <-active.Changed:
				continue
			case <-stopChan:
				glog.Infof("[%s] Got stop signal", c.Handler)
				break Loop
			}
		}

		if consumer != nil {
			consumer.Stop()
		}
		setupConsumer()
		select {
		case <-active.Changed:
			glog.Infof("[%s] Active state changed to %v", c.Handler, c.Active())
		case <-concurrent.Changed:
			glog.Infof("[%s] Concurrency changed to %v", c.Handler, c.Concurrent())
		case <-consume.Changed:
			glog.Infof("[%s] Consume topic changed to %v", c.Handler, c.Consume())
		case <-stopChan:
			glog.Infof("[%s] Got stop signal", c.Handler)
			break Loop
		}
	}

	if consumer != nil {
		glog.Infof("[%s] Stopping consumer", c.Handler)
		consumer.Stop()
	}
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
				glog.Infof("[%s] Got stop signal", p.Handler)
				break Loop
			}
		}

		if f := p.Frequency(); f < MinFrequency {
			glog.Errorf("[%s] Current frequency (%s) shorter than minimum (%s). Backing off until changed.", p.Handler, f, MinFrequency)
			select {
			case <-freq.Changed:
				continue
			case <-stopChan:
				glog.Infof("[%s] Got stop signal", p.Handler)
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
			glog.Infof("[%s] Got stop signal", p.Handler)
			break Loop
		}
	}
	glog.Infof("[%s] Stopping producer", p.Handler)
	stopChan <- true
}

func consumerHandler(h *handler.Handler) (f nsq.HandlerFunc) {
	return func(msg *nsq.Message) (err error) {
		glog.Infof("%s: %s %d", h.Command[0], time.Unix(0, msg.Timestamp), msg.Attempts)
		buf := bytes.NewBuffer(msg.Body)
		return h.Run(config.Etcd(), buf.String())
	}
}
