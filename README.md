# process [CloudEvent] with [go template]

![build](https://github.com/alitari/ce-go-template/workflows/TestAndBuild/badge.svg)
![build](https://github.com/alitari/ce-go-template/workflows/PublishImages/badge.svg)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/alitari/ce-go-template?style=plastic)
[![Go Report Card](https://goreportcard.com/badge/github.com/alitari/ce-go-template)](https://goreportcard.com/report/github.com/alitari/ce-go-template)
[![codecov](https://codecov.io/gh/alitari/ce-go-template/branch/main/graph/badge.svg)](https://codecov.io/gh/alitari/ce-go-template)

There are 3 kinds of services for transforming a [CloudEvent] with the [go template] syntax:

- producers
- mappers
- filters

## producers

```txt
Input --> **Go-Template for building CloudEvents** --> CloudEvent
```

Go-Template transforms an input data structure to a cloudEvent and sends them to an [event sink]. In [knative] a producer can be applied as an event source using a [ContainerSource] or a [Sinkbinding] 

### ce-go-template-producer

This producer has no input. The transformation is triggerd by a constant frequence which is configurable.

### ce-go-template-http-producer

In this producer the transformation is triggered by an incoming http request.

## mappers

A mapper transforms an incoming CloudEvent to an outgoing CloudEvent. Depending whether an [event sink] is present, the new event is either sent to the sink ( *send mode*), or is the payload of the http response (*reply mode*) 

### ce-go-template-mapper

```txt
CloudEvent --> **Go-Template for building CloudEvent** --> CloudEvent
```

### ce-go-template-http-mapper

```txt
CloudEvent --> **Go-Template for building HTTP-Request** --> Send HTTP-Request --> HTTP-Response -> **Go-Template for building CloudEvent** --> CloudEvent 
```

## filters

A filter replies with the incoming CloudEvent, if a predicate built by a go-template resolves to true. Otherwise the response has no content. In [knative] a filter can be applied in [Flows] like [Parallel]

### ce-go-template-filter

```txt
CloudEvent --> **Go-Template for building a boolean** --> CloudEvent if true, nothing otherwise
```

### ce-go-template-http-filter

```txt
CloudEvent --> **Go-Template for building HTTP-Request** --> Send HTTP-Request --> HTTP-Response --> **Go-Template for building a boolean** --> CloudEvent if true, nothing otherwise
```



## usage

In order to implement CloudEvent transformations you define a go-template representing a result as a string. For the producer and mapper this result is a [JSON representation of CloudEvent]. The result of the filter is a boolean as string. Services with incoming CloudEvents ( mapper, filter) can evaluate the incoming CloudEvent for creating their result. 

### example for a mapper go-template

```txt
{ 
   "data": {{ toJson .data }},
   "datacontenttype":"application/json",
   "id":" {{ uuidv4 }}",
   "source":"{{ .source }}",
   "specversion":"{{ .specversion }}",
   "type":"{{ .type }}" }
}
```
This transformation keeps all data of the input CloudEvent except the id. The id is created with the [sprig function `uuidv4`](http://masterminds.github.io/sprig/uuid.html).

### example for a simple mapping in ONLY_PAYLOAD

In order to reduce the template code, you can switch to `ONLY_PAYLOAD` mode, where the go-template represents just the JSON of [CloudEvent Data]. The values of [CloudEvent context attributes] are taken from the input event. So, the go-template for the same transformation like above, can be reduced to

```txt
{{ toJson .data }}
```

### example for a filter

The simplest go-template for a filter is an empty string (`""`) which implements a blocker. The go-template `"true"` is the implementation for a filter where all events can get through. A bit more demanding is a filter which accepts only events from the source "mysource":

```txt
{{ eq .source "mysource" | toString }}
```


### configuration

| Env variable | Description | Default | Producer | Mapper | Filter |
| ------------ | ------------| ------- | -------  | ---    | ------ | 
| `CE_TEMPLATE` | see details in usage | see code [producer](cmd/producer/main.go), [mapper](cmd/mapper/main.go), [filter](cmd/filter/main.go)  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
| `VERBOSE` | logs details if `true` |`true`| :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark:|
| `K_SINK` | destination uri of the outgoing CloudEvent |no | :heavy_check_mark: | :heavy_check_mark: (empty for reply mode) | :heavy_minus_sign: |
| `ONLY_PAYLOAD` | if `true` go-template represents only [CloudEvent Data] | `true` | :heavy_minus_sign:  | :heavy_check_mark: | :heavy_minus_sign: |
| `PERIOD` | duration between two CloudEvents  |`1000ms`| :heavy_check_mark: | :heavy_minus_sign: | :heavy_minus_sign: |
| `TIMEOUT` | duration for timeout when sending CloudEvent to sink |`1000ms`| :heavy_check_mark: | :heavy_minus_sign: | :heavy_minus_sign: |


## examples

As the go-template includes the [sprig functions] you can use built-in functionality for math, security/encryption, etc.

### mapper

#### eliminate duplicates

```bash
CE_TEMPLATE='{{ $people := .data.people | uniq }} { "people": {{ toJson $people }} }' go run cmd/mapper/main.go

http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" people:='[ { "name": "Bob", "age": "23" }, { "name": "John", "age": "17" } , {"name": "Bill", "age": "70"}, { "name": "Bob", "age": "23" } ]'
```

#### grouping

```bash
CE_TEMPLATE='{{ $people := .data.people }} {{ $adults := list }} {{ $children := list }} {{ range $people }} {{ $age := .age | atoi }} {{ if gt $age 17 }} {{ $adults = append $adults . }}{{ else }}{{ $children = append $children . }}{{ end }} {{ end }}{ "adults": {{ toJson $adults }}, "children": {{ toJson $children }} }' go run cmd/mapper/main.go

http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" people:='[ { "name": "Bob", "age": "23" }, { "name": "John", "age": "17" } , {"name": "Bill", "age": "70"} ]'
```

#### encrypt/decrypt secret parts of event payload

```bash
# encrypt
# start encrypt mapper 
CE_TEMPLATE='{ "foo": {{ toJson .data.foo }}, "secret": "{{ encryptAES (env "SECRET_KEY") (toJson .data.secret) }}" }' SECRET_KEY="mysecretKey" CE_PORT=8070 go run cmd/mapper/main.go
# encrypt event ( use new shell)
http POST localhost:8070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret:='{ "name": "James", "lastName": "Bond"}'
# save the encrypted response part
ENCRYPTED_SECRET=$(http --print=b POST localhost:8070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret:='{ "name": "James", "lastName": "Bond"}' | jq -r .secret)
# decrypt
# start the decrypt mapper (use a new shell)
CE_TEMPLATE='{ "foo": {{ toJson .data.foo }}, "secret": {{ .data.secret | decryptAES (env "SECRET_KEY") }} }' SECRET_KEY="mysecretKey" go run cmd/mapper/main.go
# decrypt encrypted source event 
http --print=Bhb POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret=$ENCRYPTED_SECRET
```

#### enrich content with external services

```bash
HTTP_TEMPLATE="GET https://api.genderize.io?name={{ .data.name }} HTTP/1.1"$'\n'"content-type: application/json"$'\n'$'\n' CE_TEMPLATE='{ "name": {{ .inputce.data.name | quote }}, "gender": {{ .httpresponse.body.gender | quote }} }' go run cmd/http-mapper/main.go

http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" name=Sabine
```

### producer

#### producing random CloudEvents

```bash
CE_TEMPLATE='{{ $rand :=  randNumeric 1 | atoi }} { "data": { {{ if gt $rand 5 }} "foo": "foovalue" {{ else }} "bar": "barvalue" {{ end }} } , "datacontenttype":"application/json","id": {{ uuidv4 | quote }}, "source":"random producer","specversion":"1.0","type":"random producer type" }' K_SINK=https://httpbin.org/post go run cmd/producer/main.go
```

## deployment options in [knative]

### event producer as container source

```bash
# create event display
kn service create event-display --image gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display --cluster-local --scale-min 1
# create event source
kubectl apply -f deployments/producer-display-eventsource.yaml
```

### event producer with sinkbinding

```bash
# create event display
kn service create event-display --image gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display --cluster-local --scale-min 1
# create the sink binding
kubectl apply -f deployments/producer-display-sinkbinding.yaml
# create producer service
kubectl create deployment event-producer --image=docker.io/alitari/ce-go-template-producer
```

### event mapper in sequence

```bash
# create event display
kn service create event-display --image gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display --cluster-local --scale-min 1
# create event mapper in reply mode
kn service create event-mapper --image=docker.io/alitari/ce-go-template-mapper --cluster-local --scale-min 1
# create sequence
kubectl apply -f deployments/sequence.yaml
# create pingsource
kubectl apply -f deployments/pingsource-sequence.yaml
```

### event mapper as subject

```bash
# create event display
kn service create event-display --image gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display --cluster-local --scale-min 1
# create the sink binding
kubectl apply -f deployments/mapper-display-sinkbinding.yaml
# create event mapper in send mode
kn service create event-mapper --image=docker.io/alitari/ce-go-template-mapper --scale-min 1
# make a request
MAPPER_URL=$(kubectl get ksvc event-mapper -o=json | jq -r .status.url)
http POST $MAPPER_URL "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: http.demo" "ce-id: 123-abc" name=Hase
```

[CloudEvent]: https://github.com/cloudevents/spec
[knative]: https://knative.dev/
[CloudEvents spec]: https://github.com/cloudevents/spec/blob/v1.0/spec.md
[CloudEvent Data]: https://github.com/cloudevents/spec/blob/v1.0/spec.md#event-data
[CloudEvent context attributes]: https://github.com/cloudevents/spec/blob/v1.0/spec.md#context-attributes
[go template]: https://golang.org/pkg/text/template/
[ContainerSource]: https://knative.dev/docs/eventing/sources/containersource/
[Sinkbinding]: https://knative.dev/docs/eventing/sources/sinkbinding/
[httpie]: https://httpie.org/
[Flows]: https://knative.dev/docs/eventing/flows/
[Parallel]: https://knative.dev/docs/eventing/flows/parallel/
[event sink]: https://redhat-developer-demos.github.io/knative-tutorial/knative-tutorial-eventing/eventing-src-to-sink.html#eventing-sink
[JSON representation of CloudEvent]: https://github.com/cloudevents/spec/blob/v1.0/json-format.md
[sprig functions]: http://masterminds.github.io/sprig/
