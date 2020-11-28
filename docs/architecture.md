```plantuml
@startuml


package "cehttpserver" {
    [httpserver]
}

package "cehandler" {
  interface CeMapper
  interface CeFilter
  interface CeProducer
  [ceMapperHandler] ..> CeMapper
  [ceFilterHandler] ..> CeFilter
  [ceProducerHandler] ..> CeProducer
  [httpserver] ..> [ceProducerHandler]
}

package "cetransformer" {
    CeMapper -- [cetransformer]
    CeFilter -- [cetransformer]
    CeProducer -- [cetransformer]
}

package "cerequesttransformer" {
    CeProducer --- [requesttransformer]
}

package "cehttpclienttransformer" {
    CeMapper --- [cehttpclienttransformer]
    CeFilter --- [cehttpclienttransformer]
    [cehttpclienttransformer] ..> [cetransformer]
    [cehttpclienttransformer] .> [httpProtocolSender]
}


package "transformer" {
  [transform]
    [cetransformer] ..> [transform]
    [cehttpclienttransformer] ..> [transform]
    [requesttransformer] ..> [transform]
}
@enduml
```