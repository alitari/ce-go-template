package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cetransformer"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

// Configuration bla
type Configuration struct {
	Verbose    bool          `default:"true"`
	CeTemplate string        `split_words:"true" default:"{\"name\":\"Alex\"}"`
	CeSource   string        `split_words:"true" default:"https://github.com/alitari/ce-go-template"`
	CeType     string        `split_words:"true" default:"com.github.alitari.ce-go-template.periodic-producer"`
	Sink       string        `envconfig:"K_SINK"`
	Period     time.Duration `default:"1000ms"`
	Timeout    time.Duration `default:"1000ms"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`
Configuration:
====================================
Verbose: %v
Period: %v
Timeout: %v
Sink: '%v'
CeTemplate: '%v'
cloudEvent source: %s
cloudEvent type: %s`, c.Verbose, c.Period, c.Timeout, c.Sink, c.CeTemplate, c.CeSource, c.CeType)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	var err error

	httpProtocol, err := cloudevents.NewHTTP(http.WithShutdownTimeout(config.Timeout))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	ceClient, err := cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	ceTransformer, err := cetransformer.NewCloudEventTransformer(config.CeTemplate, config.CeSource, config.CeType, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}

	ceProducerHandler := cehandler.NewProducerHandler(ceTransformer, ceClient, config.Sink, config.Timeout, config.Verbose)

	ticker := time.NewTicker(config.Period)
	go func() {
		for {
			select {
			case <-ticker.C:
				result := ceProducerHandler.SendCe(nil)
				if result != nil {
					log.Print(result)
				}
			}
		}
	}()
	select {}
}
