package cehandler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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
func NewProducerHandler(producer CeProducer, ceClient cloudevents.Client, sink string, timeout time.Duration, debug bool) *CeProducerHandler {
	cph := CeProducerHandler{producer: producer, ceClient: ceClient, sink: sink, timeout: timeout, debug: debug}
	return &cph
}

// SendCe producer and send cloudEvent
func (cph *CeProducerHandler) SendCe(input interface{}) error {
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
	if result != nil {
		if !strings.HasPrefix(result.Error(), "20") {
			return fmt.Errorf("Failed to send event! error: %v", result.Error())
		}
		if cloudevents.IsUndelivered(result) {
			return fmt.Errorf("Event was not delivered: %v", result)
		}
		if cph.debug {
			log.Printf("Event successfully delivered: %s", result.Error())
		}
	}
	return nil
	
}