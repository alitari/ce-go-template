package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"net/http"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

// var ceProducer *cehttpclienttransformer.cehttpclienttransformer
var ceClient cloudevents.Client = nil
var ceProducerHandler *cehandler.CeProducerHandler
var config Configuration

// Configuration bla
type Configuration struct {
	Verbose    bool   `default:"true"`
	CeTemplate string `split_words:"true" default:"{ \"data\":{\"name\":\"Alex\"}, \"datacontenttype\":\"application/json\",\"id\":\" {{ uuidv4 }}\",\"source\":\"ce-gotemplate\",\"specversion\":\"1.0\",\"type\":\"test\" }"`
	Sink       string `envconfig:"K_SINK"`
	Timeout    string `default:"1000ms"`
	HTTPPort   int    `split_words:"true" default:"8080"`
	HTTPPath   string `split_words:"true" default:"/"`
	HTTPMethod string `split_words:"true" default:"GET"`
	HTTPAccept string `split_words:"true" default:"application/json"`
}

func (c Configuration) info() string {
	return fmt.Sprintf("Configuration:\n====================================\nVerbose: %v\nTimeout: %v\nSink: '%v'\nCeTemplate: '%v'\nServing HTTP %s %s on port %v \n", c.Verbose, c.Timeout, c.Sink, c.CeTemplate, c.HTTPMethod, c.HTTPPath, c.HTTPPort)
}

func main() {
	config = Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	var err error

	timeout, err := time.ParseDuration(config.Timeout)
	httpProtocol, err := cloudevents.NewHTTP(cehttp.WithShutdownTimeout(timeout))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	ceClient, err = cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	// ceProducer = cehttpclienttransformer.Newcehttpclienttransformer("", config.CeTemplate, 0, true, false, config.Verbose)

	// ceProducerHandler = cetransformer.NewProducerHandler(ceProducer, config.Sink, timeout, config.Verbose)

}

func handleRequest(w http.ResponseWriter, r *http.Request) {
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
