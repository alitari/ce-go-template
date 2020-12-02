package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cehttpserver"
	"github.com/alitari/ce-go-template/pkg/cerequesttransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

// var ceProducer *cehttpclienttransformer.cehttpclienttransformer
// var ceClient cloudevents.Client = nil
// var ceProducerHandler *cehandler.CeProducerHandler
// var config Configuration

// Configuration bla
type Configuration struct {
	Verbose    bool          `default:"true"`
	CeTemplate string        `split_words:"true" default:"{\"name\":\"Alex\"}"`
	CeSource   string        `split_words:"true" default:"https://github.com/alitari/ce-go-template"`
	CeType     string        `split_words:"true" default:"com.github.alitari.ce-go-template.periodic-producer"`
	Sink       string        `envconfig:"K_SINK"`
	Timeout    time.Duration `default:"1000ms"`
	HTTPPort   int           `split_words:"true" default:"8080"`
	HTTPPath   string        `split_words:"true" default:"/"`
	HTTPMethod string        `split_words:"true" default:"GET"`
	HTTPAccept string        `split_words:"true" default:"application/json"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`
Configuration:
====================================
Verbose: %v
Timeout: %v
Sink: '%v
CeTemplate: '%v'
CloudEvent source: %s
CloudEvent type: %s
Serving HTTP %s on path '%s' listening on port %v accepting '%s'`, c.Verbose, c.Timeout, c.Sink, c.CeTemplate, c.CeSource, c.CeType, c.HTTPMethod, c.HTTPPath, c.HTTPPort, c.HTTPAccept)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	var err error

	httpProtocol, err := cloudevents.NewHTTP(cehttp.WithShutdownTimeout(config.Timeout))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	ceClient, err := cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	ceProducer, err := cerequesttransformer.NewRequestTransformer(config.CeTemplate, config.CeType, config.CeSource, config.Verbose)
	if err != nil {
		log.Fatalf("failed to create request transformer: %s", err.Error())
	}
	ceProducerHandler := cehandler.NewProducerHandler(ceProducer, ceClient, config.Sink, config.Timeout, true)
	cehttpserver.NewCeHTTPServer(config.HTTPPort, config.HTTPPath, config.HTTPMethod, config.Verbose, ceProducerHandler)

	select {}

}
