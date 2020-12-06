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

// HTTPSender bla
type HTTPSender interface {
	Send() (*http.Response, error)
}

// Config bla
type Config struct {
	SenderCreator func(string, time.Duration, bool) (HTTPSender, error)
	RequestTemplate  string
	ResponseTemplate    string
	Timeout       time.Duration
	JSONBody      bool
	OnlyPayload   bool
	Debug         bool
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
func NewCeHTTPClientTransformer(requestTemplate string, responseTemplate string, timeout time.Duration, jsonBody bool, debug bool) (*CeHTTPClientTransformer, error) {
	return ceHTTPClientTransformer(func(protocol string, timeOut time.Duration, debug bool) (HTTPSender, error) {
		return NewHTTPProtocolSender(protocol, timeOut, debug)
	}, requestTemplate, responseTemplate, timeout, jsonBody, debug)
}

func ceHTTPClientTransformer(senderCreator func(string, time.Duration, bool) (HTTPSender, error), requestTemplate string, responseTemplate string, timeout time.Duration, jsonBody bool, debug bool) (*CeHTTPClientTransformer, error) {
	cht := new(CeHTTPClientTransformer)
	cht.config = Config{SenderCreator: senderCreator, RequestTemplate: requestTemplate, ResponseTemplate: responseTemplate, Timeout: timeout, JSONBody: jsonBody, Debug: debug}
	httpTransformer, err := transformer.NewTransformer(cht.config.RequestTemplate, nil, cht.config.Debug)
	if err != nil {
		return nil, err
	}
	cht.httpTransformer = httpTransformer
	ceTransformer, err := transformer.NewTransformer(cht.config.ResponseTemplate, nil, cht.config.Debug)
	if err != nil {
		return nil, err
	}
	cht.ceTransformer = ceTransformer

	return cht, nil
}

func (ct *CeHTTPClientTransformer) transformEventToBytes(sourceEvent *cloudevents.Event) ([]byte, error) {
	inputEventData := cetransformer.EventToMap(sourceEvent)
	httpBytes, err := ct.httpTransformer.TransformInputToBytes(inputEventData)
	if err != nil {
		return nil, err
	}
	sender, err := ct.config.SenderCreator(string(httpBytes), ct.config.Timeout, ct.config.Debug)
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
	if err != nil {
		return nil, err
	}
	return eventBytes, nil
}

// TransformEvent bla
func (ct *CeHTTPClientTransformer) TransformEvent(sourceEvent *cloudevents.Event) (*cloudevents.Event, error) {
	eventBytes, err := ct.transformEventToBytes(sourceEvent)
	if err != nil {
		return nil, err
	}

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
	booleanBytes, err := ct.transformEventToBytes(sourceEvent)
	if err != nil {
		return false, err
	}
	resultStr := string(booleanBytes)
	return resultStr == "true", nil
}

// ResponseToMap bla
func ResponseToMap(response *http.Response, jsonBody bool) (map[string]interface{}, error) {
	responseMap := map[string]interface{}{}
	if response != nil {
		responseMap["header"] = response.Header
		responseMap["status"] = response.Status
		responseMap["statusCode"] = response.StatusCode
		if response.Body != nil {
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
	}
	return responseMap, nil
}
