package cehandler

import (
	"context"
	"log"

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
func NewCeMapperHandler(ceMapper CeMapper, ceClient cloudevents.Client, sink string, debug bool) (*CeMapperHandler, error) {
	ceh := new(CeMapperHandler)
	ceh.transformer = ceMapper
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
