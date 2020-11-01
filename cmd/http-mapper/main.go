package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/alitari/ce-go-template/pkg/cetransformer"
	"github.com/alitari/ce-go-template/pkg/httpprotocolsender"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
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
	Verbose      bool   `default:"true"`
	CeTemplate   string `split_words:"true" default:"{{ toJson .data }}"`
	HTTPTemplate string `split_words:"true" default:""`
	HTTPTimeout  string `split_words:"true" default:"1000ms"`
	HTTPJsonBody bool   `split_words:"true" default:"true"`
	OnlyPayload  bool   `split_words:"true" default:"true"`
	CePort       int    `split_words:"true" default:"8080"`
	Sink         string `envconfig:"K_SINK"`
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
	return fmt.Sprintf("Configuration:\n====================================\nSink: %v (using %s)\nVerbose: %v\nHTTP Template: '%s'\nTransform only payload: %v\nServing on Port: %v\nCeTemplate: '%v'", c.Sink, c.mode(), c.Verbose, c.HTTPTemplate, c.OnlyPayload, c.CePort, c.CeTemplate)
}

var ceTransformer = &cetransformer.CloudEventTransformer{}
var count = 0
var httpSender = httpprotocolsender.HTTPProtocolSender{}
var httpTplt = template.Must(template.New("httpTemplate").Funcs(sprig.TxtFuncMap()).Parse(config.HTTPTemplate))

var ceClient cloudevents.Client = nil
var config Configuration

func main() {
	config = Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.info())

	ceTransformer.Config = cetransformer.CloudEventTransformerConfig{CeTemplate: config.CeTemplate, Debug: config.Verbose, OnlyPayload: config.OnlyPayload, InputGenerator: generateInputWithHTTP}
	httpSenderTimeout, err := time.ParseDuration(config.HTTPTimeout)
	if err != nil {
		log.Fatal(err.Error())
	}
	httpSender.Timeout = httpSenderTimeout

	httpProtocol, err := cloudevents.NewHTTP(cloudevents.WithPort(config.CePort))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	ceClient, err = cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}
	ceTransformer.Init()

	var receiver interface{} // the SDK reflects on the signature.
	if config.mode() == send {
		receiver = ReceiveAndSend
	} else {
		receiver = ReceiveAndReply
	}

	if err = ceClient.StartReceiver(context.Background(), receiver); err != nil {
		log.Fatal(err)
	}
}

func generateInputWithHTTP(event *cloudevents.Event) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	evtData, err := cetransformer.EventInputGenerator(event)
	if err != nil {
		return nil, err
	}
	result["inputce"] = evtData

	requestBuf := &bytes.Buffer{}
	err = httpTplt.Execute(requestBuf, evtData)
	if err != nil {
		return nil, err
	}

	err = httpSender.Parse(requestBuf.String())
	if err != nil {
		return nil, err
	}
	resp, err := httpSender.Send()
	if err != nil {
		return nil, err
	}
	respMap, err := httpSender.ResponseToMap(resp, config.HTTPJsonBody)
	if err != nil {
		return nil, err
	}
	result["httpresponse"] = respMap
	return result, nil
}

// ReceiveAndReply is invoked whenever we receive an event in reply mode
func ReceiveAndReply(ctx context.Context, sourceEvent cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	destEvent, err := ceTransformer.TransformEvent(&sourceEvent)
	if err != nil {
		return nil, http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	return destEvent, nil
}

// ReceiveAndSend is invoked whenever we receive an event in send mode
func ReceiveAndSend(ctx context.Context, sourceEvent cloudevents.Event) protocol.Result {
	destEvent, err := ceTransformer.TransformEvent(&sourceEvent)
	if err != nil {
		return http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	if config.Verbose {
		log.Printf("sending event: %v", destEvent)
	}
	result := ceClient.Send(cloudevents.ContextWithTarget(ctx, config.Sink), *destEvent)
	return result
}
