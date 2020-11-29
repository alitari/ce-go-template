package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cetransformer"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

var ceTransformer *cetransformer.CloudEventTransformer
var ceClient cloudevents.Client = nil
var ceProducerHandler *cehandler.CeProducerHandler

// Configuration bla
type Configuration struct {
	Verbose    bool   `default:"true"`
	CeTemplate string `split_words:"true" default:"{\"name\":\"Alex\"}"`
	CeSource   string `split_words:"true" default:"ce-go-template-producer"`
	CeType     string `split_words:"true" default:"default-type"`
	Sink       string `envconfig:"K_SINK"`
	Period     string `default:"1000ms"`
	Timeout    string `default:"1000ms"`
}

func (c Configuration) info() string {
	return fmt.Sprintf("Configuration:\n====================================\nVerbose: %v\nPeriod: %v\nTimeout: %v\nSink: '%v'\nCeTemplate: '%v'\ncloudEvent source: %s\ncloudEvent type: %s", c.Verbose, c.Period, c.Timeout, c.Sink, c.CeTemplate, c.CeSource, c.CeType)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

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
	ceTransformer, err := cetransformer.NewCloudEventTransformer(config.CeTemplate, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}

	ceProducerHandler = cehandler.NewProducerHandler(ceTransformer, ceClient, config.Sink, timeout, config.Verbose)

	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ticker.C:
				result := ceProducerHandler.SendCe(nil)
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
