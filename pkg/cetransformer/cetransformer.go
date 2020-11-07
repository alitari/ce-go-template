package cetransformer

import (
	"encoding/json"

	"github.com/alitari/ce-go-template/pkg/transformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// CloudEventTransformer bla
type CloudEventTransformer struct {
	transformer *transformer.Transformer
	onlyPayload bool
}

// NewCloudEventTransformer new instance of CloudEventTransformer
func NewCloudEventTransformer(ceTemplate string, onlyPayload bool, debug bool) *CloudEventTransformer {
	cet := new(CloudEventTransformer)
	cet.onlyPayload = onlyPayload
	cet.transformer = transformer.NewTransformer(transformer.Config{CeTemplate: ceTemplate, Debug: debug})
	return cet
}

// EventAsInput Transform a cloudevent to input data
func EventAsInput(event *cloudevents.Event) map[string]interface{} {
	evt := map[string]interface{}{}
	if event != nil {
		evtData := map[string]interface{}{}
		event.DataAs(&evtData)
		evt["data"] = evtData
		evt["type"] = event.Type()
		evt["source"] = event.Source()
		evt["id"] = event.ID()
		evt["datacontenttype"] = event.DataContentType()
		evt["specversion"] = event.SpecVersion()
	}
	return evt
}

// Unmarshal an cloudevent in json format
func Unmarshal(source []byte, event *cloudevents.Event, onlyPayload bool) error {
	var err error
	if onlyPayload {
		data := map[string]interface{}{}
		err = json.Unmarshal(source, &data)
		event.SetData(cloudevents.ApplicationJSON, data)
	} else {
		err = json.Unmarshal(source, event)
	}
	if err != nil {
		return err
	}
	return nil
}

// TransformBytesToEventOnlyPayload bla
func (ct *CloudEventTransformer) TransformBytesToEventOnlyPayload(eventMarshalled []byte, context cloudevents.EventContext) (*cloudevents.Event, error) {
	var resultEvent cloudevents.Event
	resultEvent = cloudevents.NewEvent()
	resultEvent.Context = context
	resultEvent.SetID(uuid.New().String())
	if err := Unmarshal(eventMarshalled, &resultEvent, ct.onlyPayload); err != nil {
		return nil, err
	}
	return &resultEvent, nil
}

// TransformBytesToEvent bla
func (ct *CloudEventTransformer) TransformBytesToEvent(eventMarshalled []byte) (*cloudevents.Event, error) {
	var resultEvent cloudevents.Event
	if err := Unmarshal(eventMarshalled, &resultEvent, ct.onlyPayload); err != nil {
		return nil, err
	}
	return &resultEvent, nil
}

// TransformEvent bla
func (ct *CloudEventTransformer) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	resultEventBytes, err := ct.transformer.TransformInputToBytes(EventAsInput(sourceEvent))
	if err != nil {
		return nil, err
	}
	var resultEvent *cloudevents.Event
	if ct.onlyPayload {
		resultEvent, err = ct.TransformBytesToEventOnlyPayload(resultEventBytes, sourceEvent.Context.Clone())
	} else {
		resultEvent, err = ct.TransformBytesToEvent(resultEventBytes)
	}
	if err != nil {
		return nil, err
	}
	return resultEvent, nil
}

// PredicateEvent bla
func (ct *CloudEventTransformer) PredicateEvent(sourceEvent *cloudevents.Event) (bool, error) {
	resultEventBytes, err := ct.transformer.TransformInputToBytes(EventAsInput(sourceEvent))
	if err != nil {
		return false, err
	}
	resultStr := string(resultEventBytes)
	return resultStr == "true", nil
}
