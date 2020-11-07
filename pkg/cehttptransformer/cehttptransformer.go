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

// Config bla
type Config struct {
	HTTPTemplate string
	CeTemplate   string
	Timeout      time.Duration
	JSONBody     bool
	OnlyPayload  bool
	Debug        bool
}

// CeHTTPTransformer bla
type CeHTTPTransformer struct {
	config          Config
	request         *http.Request
	client          *http.Client
	httpTransformer *transformer.Transformer
	ceTransformer   *transformer.Transformer
}

// NewCeHTTPTransformer bla
func NewCeHTTPTransformer(httpTemplate string, ceTemplate string, timeout time.Duration, jsonBody bool, onlyPayload bool, debug bool) *CeHTTPTransformer {
	cht := new(CeHTTPTransformer)
	cht.config = Config{HTTPTemplate: httpTemplate, CeTemplate: ceTemplate, Timeout: timeout, JSONBody: jsonBody, OnlyPayload: onlyPayload, Debug: debug}
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
	input := map[string]interface{}{}
	input["inputce"] = inputEventData
	input["httpresponse"] = respData
	eventBytes, err := ct.ceTransformer.TransformInputToBytes(input)
	result := cloudevents.NewEvent()

	if err := cetransformer.Unmarshal(eventBytes, &result, ct.config.OnlyPayload); err != nil {
		return nil, err
	}
	if ct.config.OnlyPayload {
		result.Context = sourceEvent.Context.Clone()
	}
	return &result, nil
}

// ResponseToMap bla
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
