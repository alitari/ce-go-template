package cetransformer

import (
	"encoding/json"

	"github.com/alitari/ce-go-template/pkg/transformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// CloudEventTransformer bla
type CloudEventTransformer struct {
	transformer  *transformer.Transformer
	resultType   string
	resultSource string
}

// NewCloudEventTransformer new instance of CloudEventTransformer ceTemplate,debug, type, source
func NewCloudEventTransformer(ceTemplate string, debug bool, context ...string) (*CloudEventTransformer, error) {
	cet := new(CloudEventTransformer)
	if len(context) > 0 {
		cet.resultType = context[0]
	}
	if len(context) > 1 {
		cet.resultSource = context[1]
	}

	transformer, err := transformer.NewTransformer(ceTemplate, nil, debug)
	if err != nil {
		return nil, err
	}
	cet.transformer = transformer
	return cet, nil
}

// EventToMap Transform a cloudevent to input data
func EventToMap(event *cloudevents.Event) map[string]interface{} {
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

// Unmarshal an json string to a cloudevent with data in a map
func Unmarshal(source []byte, event *cloudevents.Event) error {
	var err error
	data := map[string]interface{}{}
	err = json.Unmarshal(source, &data)
	event.SetData(cloudevents.ApplicationJSON, data)
	if err != nil {
		return err
	}
	return nil
}

// TransformBytesToEvent bla
func TransformBytesToEvent(eventMarshalled []byte, context cloudevents.EventContext) (*cloudevents.Event, error) {
	var resultEvent cloudevents.Event
	resultEvent = cloudevents.NewEvent()
	resultEvent.Context = context
	resultEvent.SetID(uuid.New().String())

	if err := Unmarshal(eventMarshalled, &resultEvent); err != nil {
		return nil, err
	}
	return &resultEvent, nil
}

// CreateEvent bla
func (ct *CloudEventTransformer) CreateEvent(input interface{}) (*cloudevents.Event, error) {
	ce := cloudevents.NewEvent()
	return ct.TransformEvent(&ce)
}

// TransformEvent bla
func (ct *CloudEventTransformer) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	resultEventBytes, err := ct.transformer.TransformInputToBytes(EventToMap(sourceEvent))
	if err != nil {
		return nil, err
	}

	resultEvent, err := TransformBytesToEvent(resultEventBytes, sourceEvent.Context.Clone())
	if err != nil {
		return nil, err
	}
	if ct.resultType == "" {
		resultEvent.SetType(sourceEvent.Type())
	} else {
		resultEvent.SetType(ct.resultType)
	}
	if ct.resultSource == "" {
		resultEvent.SetSource(sourceEvent.Source())
	} else {
		resultEvent.SetSource(ct.resultSource)
	}

	return resultEvent, nil
}

// PredicateEvent bla
func (ct *CloudEventTransformer) PredicateEvent(sourceEvent *cloudevents.Event) (bool, error) {
	resultEventBytes, err := ct.transformer.TransformInputToBytes(EventToMap(sourceEvent))
	if err != nil {
		return false, err
	}
	resultStr := string(resultEventBytes)
	return resultStr == "true", nil
}
