# periodic producer

## configuration

| Name | Default | Description |
| ---- | ------- | ----------- |
| `VERBOSE` | `true` | if `true` you get an extensive log output |
| `CE_TEMPLATE` | `{"name": "Alex"}` | example valid json |
| `CE_SOURCE` | `https://github.com/alitari/ce-go-template` | [Cloudevent Source](https://github.com/cloudevents/spec/blob/v1.0/spec.md#source-1)  |
| `CE_TYPE` | `com.github.alitari.ce-go-template.periodic-producer` | [Cloudevent Type](https://github.com/cloudevents/spec/blob/v1.0/spec.md#type)  |
| `K_SINK` |  | An adressable K8s resource. see [Sinkbinding](https://knative.dev/docs/eventing/samples/sinkbinding/)  |
| `PERIOD` | `1000ms` | frequency of sending events |
| `TIMEOUT` | `1000ms` | send timeout | 

## examples

### default

```bash
K_SINK=https://httpbin.org/post go run cmd/periodic-producer/main.go
```

### producing random CloudEvents

```bash
CE_TEMPLATE='{{ $rand :=  randNumeric 1 | atoi }}'\
'{ {{ if gt $rand 5 }}'\
'"foo": "foovalue"'\
'{{ else }}'\
'"bar": "barvalue"'\
'{{ end }} }' \
K_SINK=https://httpbin.org/post go run cmd/periodic-producer/main.go
```

