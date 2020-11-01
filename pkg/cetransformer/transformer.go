package cetransformer

import (
	"bytes"
	"encoding/json"
	"log"
	"text/template"

	sprig "github.com/Masterminds/sprig"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// CloudEventTransformerConfig bla
type CloudEventTransformerConfig struct {
	CeTemplate       string
	Debug            bool
	OnlyPayload      bool
	InputGenerator   func(event *cloudevents.Event) (map[string]interface{}, error)
	FuncMapExtension template.FuncMap
}

// CloudEventTransformer bla
type CloudEventTransformer struct {
	Config CloudEventTransformerConfig
	tplt   *template.Template
	count  uint64
}

// Init bla
func (ct *CloudEventTransformer) Init() {
	if ct.Config.InputGenerator == nil {
		ct.Config.InputGenerator = EventInputGenerator
	}
	if ct.Config.FuncMapExtension == nil {
		ct.Config.FuncMapExtension = template.FuncMap{
			"count": func() uint64 {
				return ct.count
			},
		}
	}
	ct.tplt = template.Must(template.New("ceTemplate").Funcs(sprig.TxtFuncMap()).Funcs(ct.Config.FuncMapExtension).Parse(ct.Config.CeTemplate))
	ct.count = 0
}

func (ct *CloudEventTransformer) transformInput(input map[string]interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}

	err := ct.tplt.Execute(buf, input)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EventInputGenerator Generates input from event
func EventInputGenerator(event *cloudevents.Event) (map[string]interface{}, error) {
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
	return evt, nil
}

func (ct *CloudEventTransformer) transformEventToBytes(event *cloudevents.Event) ([]byte, error) {
	input, err := ct.Config.InputGenerator(event)
	if err != nil {
		return nil, err
	}
	return ct.transformInput(input)
}

func (ct *CloudEventTransformer) unmarshal(source []byte, event *cloudevents.Event) error {
	var err error
	if ct.Config.OnlyPayload {
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

// TransformEvent bla
func (ct *CloudEventTransformer) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	if ct.Config.Debug {
		log.Printf("source event: '%v'", sourceEvent)
	}
	ct.count++
	resultEventBytes, err := ct.transformEventToBytes(sourceEvent)
	if err != nil {
		return nil, err
	}
	if ct.Config.Debug {
		if ct.Config.OnlyPayload {
			log.Printf("destination event payload as Json:   '%s'", string(resultEventBytes))
		} else {
			log.Printf("destination event as Json:   '%s'", string(resultEventBytes))
		}
	}
	var resultEvent cloudevents.Event
	if ct.Config.OnlyPayload {
		resultEvent = cloudevents.NewEvent()
		resultEvent.Context = sourceEvent.Context.Clone()
		resultEvent.SetID(uuid.New().String())
		ct.unmarshal(resultEventBytes, &resultEvent)
	} else {
		ct.unmarshal(resultEventBytes, &resultEvent)
	}
	if ct.Config.Debug {
		log.Printf("destination event:   '%v'", resultEvent)
	}
	return &resultEvent, nil
}

// PredicateEvent bla
func (ct *CloudEventTransformer) PredicateEvent(sourceEvent *cloudevents.Event) (bool, error) {
	if ct.Config.Debug {
		log.Printf("source event: '%v'", sourceEvent)
	}
	ct.count++
	resultEventBytes, err := ct.transformEventToBytes(sourceEvent)
	if err != nil {
		return false, err
	}
	resultStr := string(resultEventBytes)
	if ct.Config.Debug {
		log.Printf("predicate result:   '%s'", resultStr)
	}
	return resultStr == "true", nil
}
