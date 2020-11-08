package main

import (
	"fmt"
	"log"

	"github.com/alitari/ce-go-template/pkg/cetransformer"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
)

//Mode depending on K_SINK env variable, directly reply with an event or send it to the sink
type Mode string

const (
	send  Mode = "send mode"
	reply Mode = "reply mode"
)

// Configuration bla
type Configuration struct {
	Verbose     bool   `default:"true"`
	CeTemplate  string `split_words:"true" default:"{{ toJson .data }}"`
	OnlyPayload bool   `split_words:"true" default:"true"`
	CePort      int    `split_words:"true" default:"8080"`
	Sink        string `envconfig:"K_SINK"`
}

func (c Configuration) mode() Mode {
	var mode Mode
	if c.Sink == "" {
		mode = reply
	} else {
		mode = send
	}
	return mode
}

func (c Configuration) info() string {
	return fmt.Sprintf("Configuration:\n====================================\nSink: %v (using %s)\nVerbose: %v\nTransform only payload: %v\nServing on Port: %v\nCeTemplate: '%v'", c.Sink, c.mode(), c.Verbose, c.OnlyPayload, c.CePort, c.CeTemplate)
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

	ceTransformer = cetransformer.NewCloudEventTransformer(config.CeTemplate, config.OnlyPayload, config.Verbose)

	httpProtocol, err := cloudevents.NewHTTP(cloudevents.WithPort(config.CePort))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	ceClient, err = cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = cetransformer.NewCeMapperHandler(ceTransformer, ceClient, config.Sink, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}

}
