package transformer

import (
	"bytes"
	"log"
	"text/template"

	sprig "github.com/Masterminds/sprig"
)

// Config Coinfiguration for the transformer
type Config struct {
}

// Transformer bla
type Transformer struct {
	debug bool
	tplt  *template.Template
	count uint64
}

// NewTransformer bla
func NewTransformer(ceTemplate string, funcMapExtension template.FuncMap, debug bool) (*Transformer, error) {
	t := new(Transformer)
	if funcMapExtension == nil {
		funcMapExtension = template.FuncMap{
			"count": func() uint64 {
				return t.count
			},
		}
	}
	tplt, err := template.New("ceTemplate").Funcs(sprig.TxtFuncMap()).Funcs(funcMapExtension).Parse(ceTemplate)
	if err != nil {
		return nil, err
	}
	t.tplt = tplt
	t.count = 0
	t.debug = debug
	return t, nil
}

// TransformInputToBytes bla
func (ct *Transformer) TransformInputToBytes(input interface{}) ([]byte, error) {
	ct.count++
	buf := &bytes.Buffer{}
	err := ct.tplt.Execute(buf, input)
	if err != nil {
		return nil, err
	}
	if ct.debug {
		log.Printf("transformed input data: %v\nto String: '%s'", input, buf.String())
	}
	return buf.Bytes(), nil
}
