package javascript

import (
	"fmt"
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"github.com/robertkrimen/otto"
)

type JSRuntime struct {
	VM *otto.Otto
}

func NewRuntime(root *dom.Node) *JSRuntime {
	vm := otto.New()
	vm.Set("console", map[string]interface{}{
		"log": func(call otto.FunctionCall) otto.Value {
			fmt.Printf("JS Console: %s\n", call.Argument(0).String())
			return otto.Value{}
		},
	})
	vm.Set("document", map[string]interface{}{
		"title": "Go Browser Engine",
	})

	return &JSRuntime{VM: vm}
}

func (r *JSRuntime) Execute(script string) error {
	_, err := r.VM.Run(script)
	return err
}
