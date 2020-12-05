package main

import (
	"fmt"
	"log"

	"github.com/alitari/ce-go-template/pkg/cehandler"
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
	Verbose    bool   `default:"true"`
	CeTemplate string `split_words:"true" default:"{{ toJson .data }}"`
	CeSource   string `split_words:"true" default:"https://github.com/alitari/ce-go-template"`
	CeType     string `split_words:"true" default:"com.github.alitari.ce-go-template.mapper"`
	CePort     int    `split_words:"true" default:"8080"`
	Sink       string `envconfig:"K_SINK"`
}

func (c Configuration) mode() Mode {
	if c.Sink == "" {
		return reply
	}
	return send
}

func (c Configuration) info() string {
	return fmt.Sprintf(`
Configuration:
====================================
Verbose: %v
Listening on port: %v
Sink: %v (using %s)
cloudEvent source: '%s'
cloudEvent type: '%s'
CeTemplate: '%v'`, c.Verbose, c.CePort, c.Sink, c.mode(), c.CeSource, c.CeType, c.CeTemplate)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	ceTransformer, err := cetransformer.NewCloudEventTransformer(config.CeTemplate, config.CeSource, config.CeType, config.Verbose)
	if err != nil {
		log.Fatalf("failed to create transformer: %s", err.Error())
	}

	httpProtocol, err := cloudevents.NewHTTP(cloudevents.WithPort(config.CePort))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	ceClient, err := cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = cehandler.NewCeMapperHandler(ceTransformer, ceClient, config.Sink, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}

}
