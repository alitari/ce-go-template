package cehttpclienttransformer

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

// CeHTTPClientTransformer bla
type CeHTTPClientTransformer struct {
	config          Config
	request         *http.Request
	client          *http.Client
	httpTransformer *transformer.Transformer
	ceTransformer   *transformer.Transformer
}

// NewCeHTTPClientTransformer bla
func NewCeHTTPClientTransformer(httpTemplate string, ceTemplate string, timeout time.Duration, jsonBody bool, onlyPayload bool, debug bool) (*CeHTTPClientTransformer, error) {
	cht := new(CeHTTPClientTransformer)
	cht.config = Config{HTTPTemplate: httpTemplate, CeTemplate: ceTemplate, Timeout: timeout, JSONBody: jsonBody, OnlyPayload: onlyPayload, Debug: debug}
	httpTransformer, err := transformer.NewTransformer(cht.config.HTTPTemplate, nil, cht.config.Debug)
	if err != nil {
		return nil, err
	}
	ceTransformer, err := transformer.NewTransformer(cht.config.CeTemplate, nil, cht.config.Debug)
	if err != nil {
		return nil, err
	}
	cht.httpTransformer = httpTransformer
	cht.ceTransformer = ceTransformer
	return cht, nil
}

// TransformEvent bla
func (ct *CeHTTPClientTransformer) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	inputEventData := cetransformer.EventToMap(sourceEvent)
	httpBytes, err := ct.httpTransformer.TransformInputToBytes(inputEventData)
	if err != nil {
		return nil, err
	}
	sender, err := NewHTTPProtocolSender(string(httpBytes), ct.config.Timeout, ct.config.Debug)
	if err != nil {
		return nil, err
	}
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

	if err := cetransformer.Unmarshal(eventBytes, &result); err != nil {
		return nil, err
	}
	if ct.config.OnlyPayload {
		result.Context = sourceEvent.Context.Clone()
	}
	return &result, nil
}

// PredicateEvent bla
func (ct *CeHTTPClientTransformer) PredicateEvent(sourceEvent *cloudevents.Event) (bool, error) {
	inputEventData := cetransformer.EventToMap(sourceEvent)
	httpBytes, err := ct.httpTransformer.TransformInputToBytes(inputEventData)
	if err != nil {
		return false, err
	}
	sender, err := NewHTTPProtocolSender(string(httpBytes), ct.config.Timeout, ct.config.Debug)
	if err != nil {
		return false, err
	}
	resp, err := sender.Send()
	if err != nil {
		return false, err
	}
	respData, err := ResponseToMap(resp, ct.config.JSONBody)
	if err != nil {
		return false, err
	}
	input := map[string]interface{}{}
	input["inputce"] = inputEventData
	input["httpresponse"] = respData
	booleanBytes, err := ct.ceTransformer.TransformInputToBytes(input)
	resultStr := string(booleanBytes)
	return resultStr == "true", nil
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
