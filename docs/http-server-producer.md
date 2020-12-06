# http-server-producer

## configuration

| Name | Default | Description |
| ---- | ------- | ----------- |
| `VERBOSE` | `true` | if `true` you get an extensive log output |
| `CE_TEMPLATE` | `{"name": "Alex"}` | example valid json |
| `CE_SOURCE` | `https://github.com/alitari/ce-go-template` | [Cloudevent Source](https://github.com/cloudevents/spec/blob/v1.0/spec.md#source-1)  |
| `CE_TYPE` | `com.github.alitari.ce-go-template.periodic-producer` | [Cloudevent Type](https://github.com/cloudevents/spec/blob/v1.0/spec.md#type)  |
| `K_SINK` |  | An adressable K8s resource. see [Sinkbinding](https://knative.dev/docs/eventing/samples/sinkbinding/)  |
| `TIMEOUT` | `1000ms` | send timeout |
| `HTTP_PORT` | `8080` | server port |
| `HTTP_PATH` | `/` | server path |
| `HTTP_METHOD` |  `GET` | server method |
| `HTTP_ACCEPT` | `application/json` | Http Accept header | 

## examples

### default

```bash
K_SINK=https://httpbin.org/post go run cmd/http-server-producer/main.go
# in a new shell
curl -v localhost:8080/
```

### path param to cloudevent payload

```bash
CE_TEMPLATE='{{ $name := trimPrefix "/person/" .url.path }}'\
'{'\
'"person": {{ $name | quote }}'\
'}' \
HTTP_PATH="/person/" K_SINK=https://httpbin.org/post go run cmd/http-server-producer/main.go
# in a new shell
curl -v localhost:8080/person/alex
```

### query param to cloudevent payload

```bash
CE_TEMPLATE='{{ $query := .url.query }}'\
'{'\
'"person": {{ index (index $query "person") 0 | quote }}'\
'}' \
K_SINK=https://httpbin.org/post go run cmd/http-server-producer/main.go
# in a new shell
curl -v localhost:8080/?person=alex
```

### request body to cloudevent payload

```bash
CE_TEMPLATE='{{ $name := .body.name }}'\
'{'\
'"person": {{ $name | quote }}'\
'}' \
HTTP_METHOD=POST \
K_SINK=https://httpbin.org/post go run cmd/http-server-producer/main.go
# in a new shell
curl -v -X POST localhost:8080/ -d '{ "name" : "Alex" }'
```



