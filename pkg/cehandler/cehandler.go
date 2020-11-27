package cehandler

import (
	"context"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// CeMapper transforms source cloudEvent in a destination cloudEvent
type CeMapper interface {
	TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error)
}

// CeMapperHandler provides callback functions for handling cloudEvents
type CeMapperHandler struct {
	transformer CeMapper
	ceClient    cloudevents.Client
	sink        string
	debug       bool
}

// NewCeMapperHandler start handling cloudEvents
func NewCeMapperHandler(transformer CeMapper, ceClient cloudevents.Client, sink string, debug bool) (*CeMapperHandler, error) {
	ceh := new(CeMapperHandler)
	ceh.transformer = transformer
	ceh.ceClient = ceClient
	ceh.sink = sink
	ceh.debug = debug
	var receiver interface{} // the SDK reflects on the signature.
	if len(sink) == 0 {
		receiver = ceh.ReceiveReplyCe
	} else {
		receiver = ceh.ReceiveSendCe
	}
	if err := ceh.ceClient.StartReceiver(context.Background(), receiver); err != nil {
		return nil, err
	}
	return ceh, nil
}

// ReceiveSendCe transform event and send it to sink
func (ceh *CeMapperHandler) ReceiveSendCe(ctx context.Context, sourceEvent cloudevents.Event) protocol.Result {
	destEvent, err := ceh.transformer.TransformEvent(&sourceEvent)
	if err != nil {
		return http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	if ceh.debug {
		log.Printf("sending event: %v", destEvent)
	}
	result := ceh.ceClient.Send(cloudevents.ContextWithTarget(ctx, ceh.sink), *destEvent)
	return result
}

// ReceiveReplyCe transform event and put the result in the response
func (ceh *CeMapperHandler) ReceiveReplyCe(ctx context.Context, sourceEvent cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	destEvent, err := ceh.transformer.TransformEvent(&sourceEvent)
	if err != nil {
		return nil, http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	return destEvent, nil
}

// CeFilter transforms source cloudEvent in a bool
type CeFilter interface {
	PredicateEvent(sourceEvent *cloudevents.Event) (bool, error)
}

// CeFilterHandler provides callback function for filtering cloudEvents
type CeFilterHandler struct {
	predicate CeFilter
	ceClient  cloudevents.Client
	debug     bool
}

// NewFilterHandler start handling cloudEvents
func NewFilterHandler(predicate CeFilter, ceClient cloudevents.Client, debug bool) (*CeFilterHandler, error) {
	cph := new(CeFilterHandler)
	cph.predicate = predicate
	cph.ceClient = ceClient
	cph.debug = debug

	if err := cph.ceClient.StartReceiver(context.Background(), cph.HandleCe); err != nil {
		return nil, err
	}
	return cph, nil
}

// HandleCe if predicate is true reply with the sourceEvent , reply with no content otherwise
func (cph *CeFilterHandler) HandleCe(ctx context.Context, sourceEvent cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	reply, err := cph.predicate.PredicateEvent(&sourceEvent)
	if err != nil {
		return nil, http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	if reply {
		return &sourceEvent, nil
	}
	return nil, http.NewResult(204, "predicate is false")
}

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
