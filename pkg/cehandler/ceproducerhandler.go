package cehandler

import (
	"context"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// CeProducer create new cloudEvent from an input
type CeProducer interface {
	CreateEvent(input interface{}) (*cloudevents.Event, error)
}

// CeProducerHandler provides callback function for producing cloudEvents
type CeProducerHandler struct {
	producer CeProducer
	ceClient cloudevents.Client
	sink     string
	debug    bool
	timeout  time.Duration
}

// NewProducerHandler create new instance
func NewProducerHandler(producer CeProducer, sink string, timeout time.Duration, debug bool) *CeProducerHandler {
	cph := new(CeProducerHandler)
	cph.producer = producer
	httpProtocol, err := cloudevents.NewHTTP(http.WithShutdownTimeout(timeout))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	ceClient, err := cloudevents.NewClient(httpProtocol)
	if err != nil {
		log.Fatal(err.Error())
	}
	cph.ceClient = ceClient
	cph.sink = sink
	cph.timeout = timeout
	cph.debug = debug
	return cph
}

// SendCe producer and send cloudEvent
func (cph *CeProducerHandler) SendCe(input interface{}) protocol.Result {
	destEvent, err := cph.producer.CreateEvent(input)
	if err != nil {
		return http.NewResult(400, "got error %v while producing event from input : %v", err, input)
	}
	if cph.debug {
		log.Printf("sending event: %v", destEvent)
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), cph.timeout)
	defer cancel()
	sendContext := cloudevents.ContextWithTarget(timeoutCtx, cph.sink)
	result := cph.ceClient.Send(sendContext, *destEvent)
	return result
}
