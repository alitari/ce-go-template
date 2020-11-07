package cetransformer

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// CeTransformer transforms source cloudEvent in a destination cloudEvent
type CeTransformer interface {
	TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error)
}

// CeTransformHandler provides callback functions for handling cloudEvents
type CeTransformHandler struct {
	transformer CeTransformer
	ceClient    cloudevents.Client
	sink        string
	debug       bool
}

// NewCeTransformHandler start handling cloudEvents
func NewCeTransformHandler(transformer CeTransformer, ceClient cloudevents.Client, sink string, debug bool) (*CeTransformHandler, error) {
	ceh := new(CeTransformHandler)
	ceh.transformer = transformer
	ceh.ceClient = ceClient
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
func (ceh *CeTransformHandler) ReceiveSendCe(ctx context.Context, sourceEvent cloudevents.Event) protocol.Result {
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
func (ceh *CeTransformHandler) ReceiveReplyCe(ctx context.Context, sourceEvent cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	destEvent, err := ceh.transformer.TransformEvent(&sourceEvent)
	if err != nil {
		return nil, http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	return destEvent, nil
}

// CePredicate transforms source cloudEvent in a bool
type CePredicate interface {
	PredicateEvent(sourceEvent *cloudevents.Event) (bool, error)
}

// CePredicateHandler provides callback function for filtering cloudEvents
type CePredicateHandler struct {
	predicate CePredicate
	ceClient  cloudevents.Client
	debug     bool
}

// NewPredicateHandler start handling cloudEvents
func NewPredicateHandler(predicate CePredicate, ceClient cloudevents.Client, debug bool) (*CePredicateHandler, error) {
	cph := new(CePredicateHandler)
	cph.predicate = predicate
	cph.ceClient = ceClient
	cph.debug = debug

	if err := cph.ceClient.StartReceiver(context.Background(), cph.HandleCe); err != nil {
		return nil, err
	}
	return cph, nil
}

// HandleCe if predicate is true reply with the sourceEvent , reply with no content otherwise
func (cph *CePredicateHandler) HandleCe(ctx context.Context, sourceEvent cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	reply, err := cph.predicate.PredicateEvent(&sourceEvent)
	if err != nil {
		return nil, http.NewResult(400, "got error %v while transforming event: %v", err, sourceEvent)
	}
	if reply {
		return &sourceEvent, nil
	}
	return nil, http.NewResult(204, "predicate is false")
}
