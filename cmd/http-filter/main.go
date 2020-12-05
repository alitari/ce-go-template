package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cehttpclienttransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
)

// Configuration bla
type Configuration struct {
	Verbose      bool   `default:"true"`
	CeTemplate   string `split_words:"true" default:"true"`
	HTTPTemplate string `split_words:"true" default:""`
	HTTPTimeout  string `split_words:"true" default:"1000ms"`
	HTTPJsonBody bool   `split_words:"true" default:"true"`
	CePort       int    `split_words:"true" default:"8080"`
}

func (c Configuration) info() string {
	return fmt.Sprintf("Configuration:\n====================================\nVerbose: %v\nServing on Port: %v\nHTTP-Template: '%v'\nCeTemplate: '%v'", c.Verbose, c.CePort, c.HTTPTemplate, c.CeTemplate)
}

var transformer *cehttpclienttransformer.CeHTTPClientTransformer
var ceClient cloudevents.Client = nil
var config Configuration

func main() {
	config = Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	httpSenderTimeout, err := time.ParseDuration(config.HTTPTimeout)
	if err != nil {
		log.Fatal(err.Error())
	}

	transformer, err = cehttpclienttransformer.NewCeHTTPClientTransformer(config.HTTPTemplate, config.CeTemplate, httpSenderTimeout, config.HTTPJsonBody, config.Verbose)
	if err != nil {
		log.Fatalf("failed to create CeHTTPClientTransformer: %s", err.Error())
	}

	httpProtocol, err := cloudevents.NewHTTP(cloudevents.WithPort(config.CePort))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	ceClient, err = cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = cehandler.NewCeFilterHandler(transformer, ceClient, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}

}
