# process [CloudEvent] with [go template]

With the following 2 services you can produce and transform a [CloudEvent] with the [go template] syntax.

- **ce-go-template-producer**: This service creates new cloud events frequently and send them to an [event sink]. In [knative] it can be applied as an event source using a [ContainerSource] or a [Sinkbinding]
- **ce-go-template-mapper**: This service transforms an incoming CloudEvent to an outgoing CloudEvent. Depending whether an [event sink] is present, the new event is either:
   - sent to the sink ( *send mode*), or
   - is the payload of the http response (*reply mode*)

## usage

In order to implement CloudEvent transformations you define a go-template representing the CloudEvent in JSON format. The payload is stored in attribute `data`, on the same level there are the [CloudEvents context attributes]. The mapper can access the input CloudEvent data with the same structure.

### example for a simple mapping

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


### configuration

| Env variable | Description | Default | Producer | Mapper |
| ------------ | ------------| ------- | -------| ---|
| `CE_TEMPLATE` | go template representing the resulting CloudEvent as JSON string | see code [producer](cmd/producer/main.go), [mapper](cmd/mapper/main.go)  | :heavy_check_mark: | :heavy_check_mark: |
| `VERBOSE` | logs details if `true` |`true`| :heavy_check_mark: | :heavy_check_mark: |
| `K_SINK` | destination uri of the outgoing CloudEvent |no | :heavy_check_mark: (mandatory)  | :heavy_check_mark: (empty for reply mode) |
| `PERIOD` | duration between two CloudEvents  |`1000ms`| :heavy_check_mark: | :heavy_minus_sign: |
| `TIMEOUT` | duration for timeout when sending CloudEvent to sink |`1000ms`| :heavy_check_mark: | :heavy_minus_sign: |


## use cases examples

As the go-template includes the [sprig functions] you can use built-in functionality for math, security/encryption, etc.

### eliminate duplicates

```bash
CE_TEMPLATE='{{ $people := .data.people | uniq }} { "data": { "people": {{ toJson $people }} } , "datacontenttype":"application/json","id":"{{ uuidv4 }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}" }' go run cmd/mapper/main.go

http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" people:='[ { "name": "Bob", "age": "23" }, { "name": "John", "age": "17" } , {"name": "Bill", "age": "70"}, { "name": "Bob", "age": "23" } ]'
```

### grouping

```bash
CE_TEMPLATE='{{ $people := .data.people }} {{ $adults := list }} {{ $children := list }} {{ range $people }} {{ $age := .age | atoi }} {{ if gt $age 17 }} {{ $adults = append $adults . }}{{ else }}{{ $children = append $children . }}{{ end }} {{ end }}{ "data": { "adults": {{ toJson $adults }}, "children": {{ toJson $children }} } , "datacontenttype":"application/json","id":"{{ uuidv4 }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}" }' go run cmd/mapper/main.go

http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" people:='[ { "name": "Bob", "age": "23" }, { "name": "John", "age": "17" } , {"name": "Bill", "age": "70"} ]'
```

### encrypt/decrypt secret parts of event payload

```bash
# encrypt
# start encrypt mapper 
CE_TEMPLATE='{ "data": { "foo": {{ toJson .data.foo }}, "secret": "{{ encryptAES (env "SECRET_KEY") (toJson .data.secret) }}" } , "datacontenttype":"application/json","id":" {{ uuidv4 }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}" }' SECRET_KEY="mysecretKey" CE_PORT=8070 go run cmd/mapper/main.go
# encrypt event ( use new shell)
http POST localhost:8070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret:='{ "name": "James", "lastName": "Bond"}'
# save the encrypted response part
ENCRYPTED_SECRET=$(http --print=b POST localhost:8070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret:='{ "name": "James", "lastName": "Bond"}' | jq -r .secret)
# decrypt
# start the decrypt mapper (use a new shell)
CE_TEMPLATE='{ "data": { "foo": {{ toJson .data.foo }}, "secret": {{ .data.secret | decryptAES (env "SECRET_KEY") }} } , "datacontenttype":"application/json","id":" {{ uuidv4 }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}" }' SECRET_KEY="mysecretKey" go run cmd/mapper/main.go
# decrypt encrypted source event 
http --print=Bhb POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret=$ENCRYPTED_SECRET
```

### producing random CloudEvents

```bash
CE_TEMPLATE='{{ $rand :=  randNumeric 1 | atoi }} { "data": { {{ if gt $rand 5 }} "foo": "foovalue" {{ else }} "bar": "barvalue" {{ end }} } , "datacontenttype":"application/json","id": {{ uuidv4 | quote }}, "source":"random producer","specversion":"1.0","type":"random producer type" }' K_SINK=https://httpbin.org/post go run cmd/producer/main.go
```

## deployment options in knative

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

## development

### run local

```bash
# run mapper in reply mode
go run cmd/mapper/main.go
# check
http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: http.demo" "ce-id: 123-abc" name=Alex
# run mapper in send mode
CE_PORT=7070 K_SINK="http://localhost:8080" go run cmd/mapper/main.go
# check mapper in sendmode -> mapper in reply mode
http POST localhost:7070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: http.demo" "ce-id: 123-abc" name=Alex
# run producer
K_SINK="http://localhost:7070" go run cmd/producer/main.go
```

### publish images

```bash
scripts/publish_image.sh producer
scripts/publish_image.sh mapper
```

[CloudEvent]: https://github.com/cloudevents/spec
[knative]: https://knative.dev/
[CloudEvents spec]: https://github.com/cloudevents/spec/blob/v1.0/spec.md
[CloudEvents context attributes]: https://github.com/cloudevents/spec/blob/v1.0/spec.md#context-attributes
[go template]: https://golang.org/pkg/text/template/
[ContainerSource]: https://knative.dev/docs/eventing/sources/containersource/
[Sinkbinding]: https://knative.dev/docs/eventing/sources/sinkbinding/
[httpie]: https://httpie.org/
[event sink]: https://redhat-developer-demos.github.io/knative-tutorial/knative-tutorial-eventing/eventing-src-to-sink.html#eventing-sink
[sprig functions]: http://masterminds.github.io/sprig/