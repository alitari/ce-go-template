package cehttptransformer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
	"github.com/alitari/ce-go-template/pkg/transformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type CeHTTPTransformerConfig struct {
	Timeout      time.Duration
	JSONBody     bool
	OnlyPayload  bool
	HTTPTemplate string
	CeTemplate   string
	Debug        bool
}

// CeHTTPTransformer bla
type CeHTTPTransformer struct {
	config          CeHTTPTransformerConfig
	request         *http.Request
	client          *http.Client
	httpTransformer *transformer.Transformer
	ceTransformer   *transformer.Transformer
}

func NewCeHTTPTransformer(config CeHTTPTransformerConfig) *CeHTTPTransformer {
	cht := new(CeHTTPTransformer)
	cht.config = config
	cht.httpTransformer = transformer.NewTransformer(transformer.Config{CeTemplate: cht.config.HTTPTemplate, Debug: cht.config.Debug})
	cht.ceTransformer = transformer.NewTransformer(transformer.Config{CeTemplate: cht.config.CeTemplate, Debug: cht.config.Debug})
	return cht
}

// TransformEvent bla
func (ct *CeHTTPTransformer) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	inputEventData := cetransformer.EventAsInput(sourceEvent)
	httpBytes, err := ct.httpTransformer.TransformInputToBytes(inputEventData)
	if err != nil {
		return nil, err
	}
	sender := NewHTTPProtocolSender(string(httpBytes), ct.config.Timeout)
	resp, err := sender.Send()
	if err != nil {
		return nil, err
	}
	respData, err := ResponseToMap(resp, ct.config.JSONBody)
	if err != nil {
		return nil, err
	}
	var input map[string]interface{}
	input["inputce"] = inputEventData
	input["httpresponse"] = respData
	eventBytes, err := ct.ceTransformer.TransformInputToBytes(input)
	result := cloudevents.Event{}

	if err := cetransformer.Unmarshal(eventBytes, &result, ct.config.OnlyPayload); err != nil {
		return nil, err
	}
	if ct.config.OnlyPayload {
		result.Context = sourceEvent.Context.Clone()
	}
	return &result, nil
}

func ResponseToMap(response *http.Response, jsonBody bool) (map[string]interface{}, error) {
	responseMap := map[string]interface{}{}
	if response != nil {
		responseMap["header"] = response.Header
		responseMap["statusCode"] = response.StatusCode
		b := new(bytes.Buffer)
		io.Copy(b, response.Body)
		response.Body.Close()
		if jsonBody {
			bodyData := map[string]interface{}{}
			if err := json.Unmarshal(b.Bytes(), &bodyData); err != nil {
				return nil, err
			}
			responseMap["body"] = bodyData
		} else {
			responseMap["body"] = b.String()
		}
	}
	return responseMap, nil
}
