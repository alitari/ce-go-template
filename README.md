# Ce-go-template

![build](https://github.com/alitari/ce-go-template/workflows/TestAndBuild/badge.svg)
![build](https://github.com/alitari/ce-go-template/workflows/PublishImages/badge.svg)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/alitari/ce-go-template?style=plastic)
[![Go Report Card](https://goreportcard.com/badge/github.com/alitari/ce-go-template)](https://goreportcard.com/report/github.com/alitari/ce-go-template)
[![codecov](https://codecov.io/gh/alitari/ce-go-template/branch/main/graph/badge.svg)](https://codecov.io/gh/alitari/ce-go-template)

`Ce-go-template` is a collection of services which process a [CloudEvent] with a [go template]. As the go-template includes the [sprig functions] you can use built-in functionality for collections, strings, math, security/encryption, etc. 
A main purpose is to use the services for building [knative eventing](https://knative.dev/docs/eventing/) primitives.
We can group the services according to their role in an event processing chain:

| Group | Knative eventing primitives | 
| --- | --- |
| producers | [ContainerSource],[Sinkbinding] |
| mappers   | [Sequence],[Parallel],[Sinkbinding] |
| filters   | [Parallel] |


## producers

![producers](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/alitari/ce-go-template/master/docs/iuml/producers.iuml)

Go-Template transforms an input data structure to a cloudEvent and sends them to an [event sink]. In [knative] a producer can be applied as an event source using a [ContainerSource] or a [Sinkbinding]

| producer name | Input | Description |
| ------------- | ------| ------------|
| ce-go-template-periodic-producer | void | Sends events frequently based on a configurable time period. See [details](docs/periodic-producer.md)
| ce-go-template-http-server-producer | HTTP-Request | Sends events based on an incoming http request. See [details](docs/http-server-producer.md) |


## mappers

A mapper transforms an incoming CloudEvent to an outgoing CloudEvent. Depending whether an [event sink] is present, the new event is either sent to the sink ( *send mode*), or is the payload of the http response (*reply mode*)

![mappers](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/alitari/ce-go-template/master/docs/iuml/mappers.iuml)

| mapper name | Description |
| ------------- | ------------|
| ce-go-template-mapper | Transforms events based on a go-template. See [details](docs/ce-go-template-mapper.md)|
| ce-go-template-http-client-mapper | Transforms an event to HTTP-Request and sends it to a HTTP server. The response is transformed to the outgoing cloud event. See [details](docs/ce-go-template-http-client-mapper.md) |


## filters

A filter replies with the incoming CloudEvent, if a predicate string built by a go-template resolves to "true". Otherwise the response has no content. In [knative] a filter can be applied in [Flows] like [Parallel]

![filters](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/alitari/ce-go-template/master/docs/iuml/filters.iuml)

| filter name | Description |
| ------------- | ------------|
| ce-go-template-filter | Transforms events to a predicate string based on a go-template. See [details](docs/ce-go-template-filter.md)|
| ce-go-template-http-client-filter | Transforms an event to HTTP-Request and sends it to a HTTP server. The response is transformed to the outgoing cloud event. See [details](docs/ce-go-template-http-client-mapper.md) |


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
kubectl create deployment event-producer --image=docker.io/alitari/ce-go-template-periodic-producer
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

## development

See [development](./docs/development.md)


[CloudEvent]: https://github.com/cloudevents/spec
[knative]: https://knative.dev/
[CloudEvents spec]: https://github.com/cloudevents/spec/blob/v1.0/spec.md
[CloudEvent Data]: https://github.com/cloudevents/spec/blob/v1.0/spec.md#event-data
[CloudEvent context attributes]: https://github.com/cloudevents/spec/blob/v1.0/spec.md#context-attributes
[go template]: https://golang.org/pkg/text/template/
[ContainerSource]: https://knative.dev/docs/eventing/sources/containersource/
[Sinkbinding]: https://knative.dev/docs/eventing/sources/sinkbinding/
[Sequence]: https://knative.dev/docs/eventing/flows/sequence/
[Parallel]: https://knative.dev/docs/eventing/flows/parallel/
[httpie]: https://httpie.org/
[Flows]: https://knative.dev/docs/eventing/flows/
[event sink]: https://redhat-developer-demos.github.io/knative-tutorial/knative-tutorial-eventing/eventing-src-to-sink.html#eventing-sink
[JSON representation of CloudEvent]: https://github.com/cloudevents/spec/blob/v1.0/json-format.md
[sprig functions]: http://masterminds.github.io/sprig/
