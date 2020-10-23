# process cloudevents with golang templates

With the following 2 services you can produce and transform cloudevents with the go-template syntax in the knative environment.

- **ce-go-template-producer**: This service creates new cloud events frequently and send them to an event sink. In knative it can be applied as an event source using a `ContainerSource` or a `Sinkbinding` 
- **ce-go-template-mapper**: This service transforms an incoming cloudevent to an destination cloudevent. Depending whether an event sink is present, the new event is either:
   - sent to the sink ( *send mode*), or
   - is the payload of the http response (*reply mode*)

## usage

In order to implement cloudevent transformations you define a go-template as an environment variable.

| Env variable | ce-go-template-producer | ce-go-template-mapper |
| ------------ | ----------------------- | --------------------- |
| ------------ | ----------------------- | --------------------- |
| ------------ | ----------------------- | --------------------- |

[TODO]

## use cases

For simple transformation tasks you can use a go-template as a runtime parameter instead of creating own code with build process and images etc.  As the go-template includes the [sprig functions] you can use built-in functionality for math, security/encryption, etc.

### Encrypt secret parts of your event payload

```bash
# encrypt
# start encrypt mapper 
CE_TEMPLATE='{ "data": { "foo": {{ toJson .data.foo }}, "secret": "{{ encryptAES (env "SECRET_KEY") (toJson .data.secret) }}" } , "datacontenttype":"application/json","id":" {{ uuidv4 }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}" }' SECRET_KEY="mysecretKey" CE_PORT=8070 go run cmd/mapper/main.go
# encrypt event ( use new shell)
http POST localhost:8070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret:='{ "name": "James", "lastName": "Bond"}'
# save the encypted response part
ENCRYPTED_SECRET=$(http --print=b POST localhost:8070 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret:='{ "name": "James", "lastName": "Bond"}' | jq -r .secret)
# decrypt
# start the decrypt mapper (use a new shell)
CE_TEMPLATE='{ "data": { "foo": {{ toJson .data.foo }}, "secret": {{ .data.secret | decryptAES (env "SECRET_KEY") }} } , "datacontenttype":"application/json","id":" {{ uuidv4 }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}" }' SECRET_KEY="mysecretKey" go run cmd/mapper/main.go
# decrypt encrypted source event 
http --print=Bhb POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=foovalue secret=$ENCRYPTED_SECRET
```

### producing random types

[TODO]



## some deployment options in knative

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
http POST http://event-mapper.default.157.97.107.125.xip.io "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: http.demo" "ce-id: 123-abc" name=Hase
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

[httpie]: https://httpie.org/
[sprig functions]: http://masterminds.github.io/sprig/