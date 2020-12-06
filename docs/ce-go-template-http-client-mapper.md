# ce-go-template-http-client-mapper

![ce-go-template-http-client-mapper](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/alitari/ce-go-template/master/docs/iuml/ce-go-template-http-client-mapper.iuml)

## configuration

| Name | Default | Description |
| ---- | ------- | ----------- |
| `VERBOSE` | `true` | if `true` you get an extensive log output |
| `REQUEST_TEMPLATE` |  | Go template for the transformation of the incoming event to a HTTP-Request in form of [RFC2616](https://tools.ietf.org/html/rfc2616#section-5). Payload of the incoming event is available under `data`. |
| `RESPONSE_TEMPLATE` | `{{ .httpresponse.body | toJson }}` | Go template for the transformation of the outcoming HTTP response to the outcoming cloud event payload. |
| `HTTP_JSON_BODY` | `true` | if true marshalls the response payload to a data structure available as `httpresponse.body` |
| `CE_SOURCE` | `https://github.com/alitari/ce-go-template` | [Cloudevent Source](https://github.com/cloudevents/spec/blob/v1.0/spec.md#source-1)  |
| `CE_TYPE` | `com.github.alitari.ce-go-template.mapper` | [Cloudevent Type](https://github.com/cloudevents/spec/blob/v1.0/spec.md#type)  |
| `K_SINK` |  | An adressable K8s resource. see [Sinkbinding](https://knative.dev/docs/eventing/samples/sinkbinding/) |
| `CE_PORT` | `8080` | server port |

### available elements in `RESONSE_TEMPLATE`

 - `inputce`: payload of the incoming event
 - `httpresponse.header`: response header
 - `httpresponse.status`: response status
 - `httpresponse.statusCode`: response status code as int
 - `httpresponse.body`: response body as struct if payload is json and `HTTP_JSON_BODY` is true, string otherwise 

## examples

### use external service to find out the gender of a name

```bash
REQUEST_TEMPLATE="GET https://api.genderize.io?name={{ .data.name }} HTTP/1.1"$'\n'"content-type: application/json"$'\n'$'\n' \
RESPONSE_TEMPLATE='{ "name": {{ .inputce.data.name | quote }}, "gender": {{ .httpresponse.body.gender | quote }} }' go run cmd/http-client-mapper/main.go
# in a new shell
http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" name=Daniela
```