package cehandler

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

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

// NewCeFilterHandler start handling cloudEvents
func NewCeFilterHandler(predicate CeFilter, ceClient cloudevents.Client, debug bool) (*CeFilterHandler, error) {
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
