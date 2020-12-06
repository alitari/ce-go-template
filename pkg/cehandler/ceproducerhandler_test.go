package cehandler

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

type CeProducerMock struct {
	t             *testing.T
	wantInput     interface{}
	outgoingEvent cloudevents.Event
	shouldThrow   error
}

func (pm *CeProducerMock) CreateEvent(input interface{}) (*cloudevents.Event, error) {
	if !reflect.DeepEqual(pm.wantInput, input) {
		pm.t.Errorf("CeProducerMock, unexpected input: actual: %v, but want %v", input, pm.wantInput)
	}
	return &pm.outgoingEvent, pm.shouldThrow
}

var outgoingEvent = cetransformer.NewEventWithJSONStringData(`{"foo": "foo"}`)
var input = "input"

func TestCeProducerHandler_SendCe(t *testing.T) {

	tests := []struct {
		name                   string
		givenProducerError     protocol.Result
		givenCeClientSendError error
		whenProducerEvent      *cloudevents.Event
		thenWantResult         protocol.Result
		thenWantOutgoingEvent  *cloudevents.Event
	}{
		{name: "Happy path ", whenProducerEvent: &outgoingEvent, thenWantOutgoingEvent: &outgoingEvent},
		{name: "Client send error ", givenCeClientSendError: errors.New("test"), whenProducerEvent: &outgoingEvent, thenWantResult: errors.New("Failed to send event! error: test")},
		{name: "Producer error ", givenProducerError: errors.New("test"), whenProducerEvent: &outgoingEvent,
			thenWantResult: http.NewResult(400, "got error %v while producing event from input : %v", errors.New("test"), input)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ceProducer := &CeProducerMock{t: t, wantInput: input, outgoingEvent: outgoingEvent, shouldThrow: tt.givenProducerError}
			ceClient := &cetransformer.CeClientMock{T: t, WantSend: true, WantSendEvent: outgoingEvent, ShouldThrowErrorOnSend: tt.givenCeClientSendError}
			ceProcuderHandler := NewProducerHandler(ceProducer, ceClient, "sink", 3*time.Second, true)
			result := ceProcuderHandler.SendCe(input)
			if !cetransformer.CompareErrors(t, "CeProducerHandler.SendCe", result, tt.thenWantResult) {
				return
			}
		})
	}
}
