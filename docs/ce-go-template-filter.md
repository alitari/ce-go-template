# ce-go-template-filter

## configuration

| Name | Default | Description |
| ---- | ------- | ----------- |
| `VERBOSE` | `true` | if `true` you get an extensive log output |
| `CE_TEMPLATE` | `true` | A go-template transforming incoming event to a string representating a predicate string|
knative.dev/docs/eventing/samples/sinkbinding/) |
| `CE_PORT` | `8080` | server port |

## examples

### default ( let all through)

```bash
go run cmd/filter/main.go

## with male surname you will get 204
http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" foo=bar
```

### list contains element

```bash
CE_TEMPLATE='{{ .data.names | has "Alex" | toString }}' go run cmd/filter/main.go

## with "Alex" in the names list you get back the event
http POST localhost:8080 "content-type: application/json" "ce-specversion: 1.0" "ce-source: http-command" "ce-type: example" "ce-id: 123-abc" names:='["Bob", "Peter"]'

```

