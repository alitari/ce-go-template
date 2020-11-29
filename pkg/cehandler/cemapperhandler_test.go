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

type CeMapperMock struct {
	t                 *testing.T
	wantIncomingEvent cloudevents.Event
	outgoingEvent     cloudevents.Event
	shouldThrow       error
}

func (mm *CeMapperMock) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	if !reflect.DeepEqual(mm.wantIncomingEvent, *sourceEvent) {
		mm.t.Errorf("CeMapperMock, unexpected sourceEvent: actual: %v, but want %v", *sourceEvent, mm.wantIncomingEvent)
	}
	return &mm.outgoingEvent, mm.shouldThrow
}

type CeClientMock struct {
	t                       *testing.T
	wantSend                bool
	wantSendEvent           cloudevents.Event
	shouldThrowErrorOnStart error
	shouldThrowErrorOnSend  error
}

func (mm *CeClientMock) StartReceiver(ctx context.Context, fn interface{}) error {
	mm.t.Logf("Callback func : %v", fn)
	return mm.shouldThrowErrorOnStart
}

func (mm *CeClientMock) Send(ctx context.Context, event cloudevents.Event) protocol.Result {
	if !mm.wantSend {
		mm.t.Errorf("CeClientMock, Send should not be called: wantSend: %v", mm.wantSend)
	}
	if !reflect.DeepEqual(mm.wantSendEvent, event) {
		mm.t.Errorf("CeClientMock, unexpected sourceEvent: actual: %v, but want %v", event, mm.wantSendEvent)
	}
	return mm.shouldThrowErrorOnSend

}

func (mm *CeClientMock) Request(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	mm.t.Errorf("CeClientMock, unexpected call: 'Request'")
	return nil, nil
}

func TestCeMapperHandler_ReceiveSendCe(t *testing.T) {

	tests := []struct {
		name                       string
		givenCeMapperError         error
		givenCeClientStartError    error
		givenCeClientSendError     error
		thenWantMapperHandlerError error
		thenWantResult             protocol.Result
	}{
		{name: "Happy path", givenCeMapperError: nil, givenCeClientStartError: nil, givenCeClientSendError: nil, thenWantMapperHandlerError: nil, thenWantResult: nil},
		{name: "Mapper error", givenCeMapperError: errors.New("test"), givenCeClientStartError: nil, givenCeClientSendError: nil, thenWantMapperHandlerError: nil,
			thenWantResult: http.NewResult(400, "got error %v while transforming event: %v", errors.New("test"), cetransformer.NewEventWithJSONStringData(`{"foo": "foo"}`))},
		{name: "Client start error", givenCeMapperError: nil, givenCeClientStartError: errors.New("test"), givenCeClientSendError: nil, thenWantMapperHandlerError: errors.New("test"), thenWantResult: nil},
		{name: "Client send error", givenCeMapperError: nil, givenCeClientStartError: nil, givenCeClientSendError: errors.New("test"), thenWantMapperHandlerError: nil,
			thenWantResult: errors.New("test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whenIncomingEvent := cetransformer.NewEventWithJSONStringData(`{"foo": "foo"}`)
			wantSendEvent := cetransformer.NewEventWithJSONStringData(`{"foo": "bar"}`)
			ceMapper := &CeMapperMock{t: t, wantIncomingEvent: whenIncomingEvent, outgoingEvent: wantSendEvent, shouldThrow: tt.givenCeMapperError}
			ceClient := &CeClientMock{t: t, wantSend: true, wantSendEvent: wantSendEvent, shouldThrowErrorOnStart: tt.givenCeClientStartError, shouldThrowErrorOnSend: tt.givenCeClientSendError}

			ceMapperHandler, err := NewCeMapperHandler(ceMapper, ceClient, "sink", true)
			if !cetransformer.CompareErrors(t, "NewCeMapperHandler", err, tt.thenWantMapperHandlerError) {
				return
			}
			if err == nil {
				result := ceMapperHandler.ReceiveSendCe(context.Background(), whenIncomingEvent)
				if !cetransformer.CompareErrors(t, "CeMapperHandler.ReceiveSendCe", result, tt.thenWantResult) {
					return
				}
			}
		})
	}
}

func TestCeMapperHandler_ReceiveReplyCe(t *testing.T) {

	tests := []struct {
		name                       string
		givenCeMapperError         error
		givenCeClientStartError    error
		thenWantMapperHandlerError error
		thenWantResult             protocol.Result
	}{
		{name: "Happy path", givenCeMapperError: nil, givenCeClientStartError: nil, thenWantMapperHandlerError: nil, thenWantResult: nil},
		{name: "Mapper error", givenCeMapperError: errors.New("test"), givenCeClientStartError: nil, thenWantMapperHandlerError: nil,
			thenWantResult: http.NewResult(400, "got error %v while transforming event: %v", errors.New("test"), cetransformer.NewEventWithJSONStringData(`{"foo": "foo"}`))},
		{name: "Client start error", givenCeMapperError: nil, givenCeClientStartError: errors.New("test"), thenWantMapperHandlerError: errors.New("test"), thenWantResult: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whenIncomingEvent := cetransformer.NewEventWithJSONStringData(`{"foo": "foo"}`)
			wantSendEvent := cetransformer.NewEventWithJSONStringData(`{"foo": "bar"}`)
			ceMapper := &CeMapperMock{t: t, wantIncomingEvent: whenIncomingEvent, outgoingEvent: wantSendEvent, shouldThrow: tt.givenCeMapperError}
			ceClient := &CeClientMock{t: t, wantSend: false, wantSendEvent: wantSendEvent, shouldThrowErrorOnStart: tt.givenCeClientStartError}

			ceMapperHandler, err := NewCeMapperHandler(ceMapper, ceClient, "sink", true)
			if !cetransformer.CompareErrors(t, "NewCeMapperHandler", err, tt.thenWantMapperHandlerError) {
				return
			}
			if err == nil {
				outgoingEvent, result := ceMapperHandler.ReceiveReplyCe(context.Background(), whenIncomingEvent)
				if !cetransformer.CompareErrors(t, "CeMapperHandler.ReceiveReplyCe", result, tt.thenWantResult) {
					return
				}
				if result == nil {
					cetransformer.CompareEvents(t, "CeMapperHandler.ReceiveReplyCe", *outgoingEvent, wantSendEvent)
				}
			}
		})
	}
}
