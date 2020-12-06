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
	Verbose          bool          `default:"true"`
	RequestTemplate  string        `split_words:"true" default:""`
	ResponseTemplate string        `split_words:"true" default:"true"`
	HTTPTimeout      time.Duration `split_words:"true" default:"1000ms"`
	HTTPJsonBody     bool          `split_words:"true" default:"true"`
	CePort           int           `split_words:"true" default:"8080"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Configuration:====================================
Verbose: %v
Serving on Port: %v
Request template: '%v
Response template: '%v'
Request timeout: '%v'
Response has JSON body: '%v'`, c.Verbose, c.CePort, c.RequestTemplate, c.ResponseTemplate, c.HTTPTimeout, c.HTTPJsonBody)
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

	transformer, err := cehttpclienttransformer.NewCeHTTPClientTransformer(config.RequestTemplate, config.ResponseTemplate, config.HTTPTimeout, config.HTTPJsonBody, config.Verbose)
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
