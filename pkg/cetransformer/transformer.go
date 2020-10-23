package cetransformer

import (
	"bytes"
	"encoding/json"
	"log"
	"text/template"

	sprig "github.com/Masterminds/sprig"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// CloudEventTransformerConfig bla
type CloudEventTransformerConfig struct {
	CeTemplate string
	Debug      bool
}

// CloudEventTransformer bla
type CloudEventTransformer struct {
	Config CloudEventTransformerConfig
	tplt   *template.Template
	count  uint64
}

// Init bla
func (ct *CloudEventTransformer) Init() {
	ct.tplt = template.Must(template.New("ceTemplate").Funcs(sprig.TxtFuncMap()).Funcs(template.FuncMap{
		"count": func() uint64 {
			return ct.count
		},
	}).Parse(ct.Config.CeTemplate))
	ct.count = 0
}

func (ct *CloudEventTransformer) transformEventToBytes(event *cloudevents.Event) ([]byte, error) {
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

	buf := &bytes.Buffer{}

	err := ct.tplt.Execute(buf, evt)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ct *CloudEventTransformer) marshal(event cloudevents.Event) ([]byte, error) {
	result, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ct *CloudEventTransformer) unmarshal(source []byte, event *cloudevents.Event) error {
	err := json.Unmarshal(source, event)
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
		log.Printf("destination event as Json:   '%s'", string(resultEventBytes))
	}
	var resultEvent cloudevents.Event
	ct.unmarshal(resultEventBytes, &resultEvent)
	if ct.Config.Debug {
		log.Printf("destination event:   '%v'", resultEvent)
	}
	return &resultEvent, nil
}
