package transformer

import (
	"bytes"
	"log"
	"text/template"

	sprig "github.com/Masterminds/sprig"
)

// Config Coinfiguration for the transformer
type Config struct {
	CeTemplate       string
	Debug            bool
	FuncMapExtension template.FuncMap
}

// Transformer bla
type Transformer struct {
	config Config
	tplt   *template.Template
	count  uint64
}

func NewTransformer(config Config) *Transformer {
	t := new(Transformer)
	t.config = config
	if t.config.FuncMapExtension == nil {
		t.config.FuncMapExtension = template.FuncMap{
			"count": func() uint64 {
				return t.count
			},
		}
	}
	t.tplt = template.Must(template.New("ceTemplate").Funcs(sprig.TxtFuncMap()).Funcs(t.config.FuncMapExtension).Parse(t.config.CeTemplate))
	t.count = 0
	return t
}

// TransformInputToBytes bla
func (ct *Transformer) TransformInputToBytes(input map[string]interface{}) ([]byte, error) {
	ct.count++
	buf := &bytes.Buffer{}
	err := ct.tplt.Execute(buf, input)
	if err != nil {
		return nil, err
	}
	if ct.config.Debug {
		log.Printf("transformed input data: %v\nto String: '%s'", input, buf.String())
	}
	return buf.Bytes(), nil
}
