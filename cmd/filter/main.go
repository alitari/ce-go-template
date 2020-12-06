package main

import (
	"fmt"
	"log"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cetransformer"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
)

// Configuration bla
type Configuration struct {
	Verbose    bool   `default:"true"`
	CeTemplate string `split_words:"true" default:"true"`
	CePort     int    `split_words:"true" default:"8080"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Configuration:
====================================
Verbose: %v
Listening on Port: %v
CeTemplate: '%v'`, c.Verbose, c.CePort, c.CeTemplate)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	ceTransformer, err := cetransformer.NewCloudEventTransformer(config.CeTemplate, "", "", config.Verbose)
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

	_, err = cehandler.NewCeFilterHandler(ceTransformer, ceClient, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}

}
