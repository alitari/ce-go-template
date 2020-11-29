package cehandler

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

type CeFilterMock struct {
	t                 *testing.T
	wantIncomingEvent cloudevents.Event
	wantPredicate     bool
	shouldThrow       error
}

func (fm *CeFilterMock) PredicateEvent(sourceEvent *cloudevents.Event) (bool, error) {
	if !reflect.DeepEqual(fm.wantIncomingEvent, *sourceEvent) {
		fm.t.Errorf("CeFilterMock, unexpected sourceEvent: actual: %v, but want %v", *sourceEvent, fm.wantIncomingEvent)
	}
	return fm.wantPredicate, fm.shouldThrow
}

var incomingEvent = cetransformer.NewEventWithJSONStringData(`{"foo": "foo"}`)

func TestCeFilterHandler_HandleCe(t *testing.T) {

	tests := []struct {
		name                       string
		givenCeClientStartError    error
		givenIncomingEvent         cloudevents.Event
		givenCeFilterError         error
		whenCeFilterPredicate      bool
		thenWantFilterHandlerError error
		thenWantResult             protocol.Result
		thenWantOutgoingEvent      *cloudevents.Event
	}{
		{name: "Happy path filter go through", givenIncomingEvent: incomingEvent, whenCeFilterPredicate: true, thenWantOutgoingEvent: &incomingEvent},
		{name: "Happy path filter blocked", givenIncomingEvent: incomingEvent, whenCeFilterPredicate: false,
			thenWantResult: http.NewResult(204, "predicate is false"), thenWantOutgoingEvent: nil},
		{name: "Filter error", givenIncomingEvent: incomingEvent, givenCeFilterError: errors.New("test"),
			thenWantResult: http.NewResult(400, "got error %v while transforming event: %v", errors.New("test"), incomingEvent), thenWantOutgoingEvent: nil},
		{name: "Client start error", givenCeClientStartError: errors.New("test"), thenWantFilterHandlerError: errors.New("test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ceFilter := &CeFilterMock{t: t, wantIncomingEvent: tt.givenIncomingEvent, wantPredicate: tt.whenCeFilterPredicate, shouldThrow: tt.givenCeFilterError}
			ceClient := &CeClientMock{t: t, wantSend: false, shouldThrowErrorOnStart: tt.givenCeClientStartError}

			ceFilterHandler, err := NewCeFilterHandler(ceFilter, ceClient, true)
			if !cetransformer.CompareErrors(t, "NewCeFilterHandler", err, tt.thenWantFilterHandlerError) {
				return
			}
			if err == nil {
				outgoingEvent, result := ceFilterHandler.HandleCe(context.Background(), tt.givenIncomingEvent)
				if !cetransformer.CompareErrors(t, "CeFilterHandler.HandleCe", result, tt.thenWantResult) {
					return
				}
				if result == nil {
					cetransformer.CompareEvents(t, "CeFilterHandler.HandleCe", *outgoingEvent, tt.givenIncomingEvent)
				}
			}
		})
	}
}
