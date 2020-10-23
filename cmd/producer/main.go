package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alitari/ce-go-template/pkg/cetransformer"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

var ceTransformer = &cetransformer.CloudEventTransformer{}
var ceClient cloudevents.Client = nil

// Configuration bla
type Configuration struct {
	Verbose    bool   `default:"true"`
	CeTemplate string `split_words:"true" default:"{ \"data\":{\"name\":\"Alex\"}, \"datacontenttype\":\"application/json\",\"id\":\" {{ uuidv4 }}\",\"source\":\"ce-gotemplate\",\"specversion\":\"1.0\",\"type\":\"test\" }"`
	Sink       string `envconfig:"K_SINK"`
	Period     string `default:"1000ms"`
	Timeout    string `default:"1000ms"`
}

func (c Configuration) info() string {
	return fmt.Sprintf("Configuration:\n====================================\nVerbose: %v\nPeriod: %v\nTimeout: %v\nSink: '%v'\nCeTemplate: '%v'\n", c.Verbose, c.Period, c.Timeout, c.Sink, c.CeTemplate)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	// if _, err := url.ParseRequestURI(config.Sink); err != nil {
	// 	log.Fatalf("error with K_SINK: %v", err)
	// }

	ceTransformer.Config = cetransformer.CloudEventTransformerConfig{CeTemplate: config.CeTemplate, Debug: config.Verbose}

	var err error

	timeout, err := time.ParseDuration(config.Timeout)
	httpProtocol, err := cloudevents.NewHTTP(http.WithShutdownTimeout(timeout))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	ceClient, err = cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}
	duration, err := time.ParseDuration(config.Period)
	if err != nil {
		log.Fatal(err.Error())
	}
	ceTransformer.Init()
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ticker.C:
				destEvent, err := ceTransformer.TransformEvent(nil)
				if err != nil {
					log.Fatal(err.Error())
				}
				if config.Verbose {
					log.Printf("Sending event to: '%s'", config.Sink)
				}

				timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				sendCtx := cloudevents.ContextWithTarget(timeoutCtx, config.Sink)
				result := ceClient.Send(sendCtx, *destEvent)
				re := result.Error()
				if !strings.HasPrefix(re, "20") {
					log.Printf("Failed to send event! error: %v", result.Error())
				} else {
					if cloudevents.IsUndelivered(result) {
						log.Printf("Event was not delivered: %v", result)
					} else {
						if config.Verbose {
							log.Printf("Event successfully delivered: %v", result)
						}
					}
				}
			}
		}
	}()
	select {}

}
