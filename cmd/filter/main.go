package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alitari/ce-go-template/pkg/cetransformer"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kelseyhightower/envconfig"
)

// Configuration bla
type Configuration struct {
	Verbose    bool   `default:"true"`
	CeTemplate string `split_words:"true" default:"true"`
	CePort     int    `split_words:"true" default:"8080"`
}

func (c Configuration) info() string {
	return fmt.Sprintf("Configuration:\n====================================\nVerbose: %v\nServing on Port: %v\nCeTemplate: '%v'", c.Verbose, c.CePort, c.CeTemplate)
}

var ceTransformer *cetransformer.CloudEventTransformer
var ceClient cloudevents.Client = nil
var config Configuration

func main() {
	config = Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	ceTransformer = cetransformer.NewCloudEventTransformer(config.CeTemplate, false, config.Verbose)

	httpProtocol, err := cloudevents.NewHTTP(cloudevents.WithPort(config.CePort))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	ceClient, err = cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = ceClient.StartReceiver(context.Background(), ReceiveAndReply); err != nil {
		log.Fatal(err)
	}
}

// ReceiveAndReply is invoked whenever we receive an event in reply mode
func ReceiveAndReply(ctx context.Context, sourceEvent cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	reply, err := ceTransformer.PredicateEvent(&sourceEvent)
	if err != nil {
		return nil, http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	if reply {
		return &sourceEvent, nil
	}
	return nil, http.NewResult(204, "predicate is false")
}
