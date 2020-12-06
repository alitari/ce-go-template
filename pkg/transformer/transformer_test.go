package transformer

import (
	"math/rand"
	"testing"
	"text/template"
)

type Foo struct {
	Name string
}

func TestTransformer_TransformInputToBytes(t *testing.T) {
	type fields struct {
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name                  string
		givenTemplate         string
		givenFuncMapExtension template.FuncMap
		whenInput             interface{}
		thenWantErr           bool
		thenWant              string
	}{
		{
			name:          "constant",
			givenTemplate: "hello World!",
			whenInput:     "doesn't matter",
			thenWantErr:   false,
			thenWant:      "hello World!",
		},
		{
			name:          "template syntax error",
			givenTemplate: "{{ .foo }}",
			whenInput:     "bar",
			thenWantErr:   true,
			thenWant:      "",
		},
		{
			name:          "simple json",
			givenTemplate: "{{ toJson . }}",
			whenInput:     Foo{Name: "Alex"},
			thenWantErr:   false,
			thenWant:      `{"Name":"Alex"}`,
		},
		{
			name:          "simple",
			givenTemplate: "{{ .Name }}",
			whenInput:     Foo{Name: "Alex"},
			thenWantErr:   false,
			thenWant:      `Alex`,
		},
		{
			name:          "count func",
			givenTemplate: "{{ count }}",
			whenInput:     "doesn't matter",
			thenWantErr:   false,
			thenWant:      `1`,
		},
		{
			name:          "myfunc",
			givenTemplate: "{{ myFunc }}",
			givenFuncMapExtension: template.FuncMap{
				"myFunc": func() uint64 {
					return 12
				},
			},
			whenInput:   "doesn't matter",
			thenWantErr: false,
			thenWant:    `12`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tranformer, err := NewTransformer(tt.givenTemplate, tt.givenFuncMapExtension, rand.Float32() < 0.5)
			if err != nil {
				if !tt.thenWantErr {
					t.Errorf("Transformer.TransformInputToBytes() error = %v, wantErr %v", err, tt.thenWantErr)
				}
				if tranformer != nil {
					t.Errorf("Transformer must be nil, but is %v", tranformer)
				}
				return
			}
			actualBytes, err := tranformer.TransformInputToBytes(tt.whenInput)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("Transformer.TransformInputToBytes() error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			if string(actualBytes) != tt.thenWant {
				t.Errorf("Transformer.TransformInputToBytes() = '%v', want '%v'", string(actualBytes), tt.thenWant)
			}
		})
	}
}
