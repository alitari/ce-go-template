package cerequesttransformer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
	"github.com/alitari/ce-go-template/pkg/transformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// RequestTransformer bla
type RequestTransformer struct {
	ceTransformer *transformer.Transformer
	ceType        string
	ceSource      string
	debug         bool
}

// NewRequestTransformer bla
func NewRequestTransformer(cetemplate string, ceType string, ceSource string, debug bool) (*RequestTransformer, error) {
	chs := new(RequestTransformer)
	chs.ceSource = ceSource
	chs.ceType = ceType
	chs.debug = debug
	ceTransformer, err := transformer.NewTransformer(cetemplate, nil, debug)
	if err != nil {
		return nil, err
	}
	chs.ceTransformer = ceTransformer
	return chs, nil
}

// CreateEvent bla
func (ct *RequestTransformer) CreateEvent(input interface{}) (*cloudevents.Event, error) {
	request := input.(http.Request)
	reqmap := map[string]interface{}{}
	reqmap["method"] = request.Method
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		bodyMap := map[string]interface{}{}
		if err := json.Unmarshal(body, &bodyMap); err != nil {
			return nil, err
		}
		reqmap["body"] = bodyMap
	}
	reqmap["host"] = request.Host

	url := map[string]interface{}{}
	url["scheme"] = request.URL.Scheme
	url["hostname"] = request.URL.Hostname()
	url["query"] = request.URL.Query()
	url["path"] = request.URL.Path
	reqmap["url"] = url
	reqmap["header"] = request.Header

	eventBytes, err := ct.ceTransformer.TransformInputToBytes(reqmap)
	if err != nil {
		return nil, err
	}
	result := cloudevents.NewEvent()
	result.SetID(uuid.New().String())
	result.SetType(ct.ceType)
	result.SetSource(ct.ceSource)
	if err := cetransformer.Unmarshal(eventBytes, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
