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

//Mode depending on K_SINK env variable, directly reply with an event or send it to the sink
type Mode string

const (
	send  Mode = "send mode"
	reply Mode = "reply mode"
)

// Configuration bla
type Configuration struct {
	Verbose          bool          `default:"true"`
	RequestTemplate  string        `split_words:"true" default:""`
	ResponseTemplate string        `split_words:"true" default:"{{ .httpresponse.body | toJson }}"`
	HTTPTimeout      time.Duration `split_words:"true" default:"1000ms"`
	HTTPJsonBody     bool          `split_words:"true" default:"true"`
	CePort           int           `split_words:"true" default:"8080"`
	Sink             string        `envconfig:"K_SINK"`
}

func (c Configuration) mode() Mode {
	if c.Sink == "" {
		return reply
	}
	return send
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Configuration:
====================================
Verbose: %v
Sink: %v (using %s)
Request template: '%s'
Response template: '%s'
HTTP Request timeout: %v
HTTP response has json body: %v
Serving on Port: %v'
`, c.Verbose, c.Sink, c.mode(), c.RequestTemplate, c.ResponseTemplate, c.HTTPTimeout, c.HTTPJsonBody, c.CePort)
}


func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	transformer, err := cehttpclienttransformer.NewCeHTTPClientTransformer(config.RequestTemplate, config.ResponseTemplate, config.HTTPTimeout, config.HTTPJsonBody, config.Verbose)
	if err != nil {
		log.Fatalf("failed to create CeHTTPClientTransformer: %s", err)
	}

	httpProtocol, err := cloudevents.NewHTTP(cloudevents.WithPort(config.CePort))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err)
	}

	ceClient, err := cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = cehandler.NewCeMapperHandler(transformer, ceClient, config.Sink, config.Verbose)
	if err != nil {
		log.Fatal(err.Error())
	}
}
