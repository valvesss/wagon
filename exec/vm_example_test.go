// Copyright 2018 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
)

func ExampleVM_add() {
	raw, err := compileWast2Wasm("testdata/add-ex-main.wast")
	if err != nil {
		log.Fatalf("could not compile wast file: %v", err)
	}

	m, err := wasm.ReadModule(bytes.NewReader(raw), func(name string) (*wasm.Module, error) {
		switch name {
		case "add":
			raw, err := compileWast2Wasm("testdata/add-ex.wast")
			if err != nil {
				return nil, fmt.Errorf("could not compile wast file hosting %q: %v", name, err)
			}

			add, err := wasm.ReadModule(bytes.NewReader(raw), nil)
			if err != nil {
				return nil, fmt.Errorf("could not read wasm %q module: %v", name, err)
			}
			return add, nil
		case "go":
			print := func(v int32) {
				fmt.Printf("result = %v\n", v)
			}

			m := wasm.NewModule()
			m.Types = &wasm.SectionTypes{
				Entries: []wasm.FunctionSig{
					{
						Form:       0,
						ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
					},
				},
			}
			m.FunctionIndexSpace = []wasm.Function{
				{
					Sig:  &m.Types.Entries[0],
					Host: reflect.ValueOf(print),
					Body: &wasm.FunctionBody{},
				},
			}
			m.Export = &wasm.SectionExports{
				Entries: map[string]wasm.ExportEntry{
					"print": {
						FieldStr: "print",
						Kind:     wasm.ExternalFunction,
						Index:    0,
					},
				},
			}

			return m, nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})

	vm, err := exec.NewVM(m)
	if err != nil {
		log.Fatalf("could not create wagon vm: %v", err)
	}

	const fct1 = 2 // index of function fct1
	out, err := vm.ExecCode(fct1)
	if err != nil {
		log.Fatalf("could not execute fct1(): %v", err)
	}
	fmt.Printf("fct1() -> %v\n", out)

	const fct2 = 3 // index of function fct2
	out, err = vm.ExecCode(fct2, 40, 6)
	if err != nil {
		log.Fatalf("could not execute fct2(40, 6): %v", err)
	}
	fmt.Printf("fct2() -> %v\n", out)

	const fct3 = 4 // index of function fct3
	out, err = vm.ExecCode(fct3, 42, 42)
	if err != nil {
		log.Fatalf("could not execute fct3(42, 42): %v", err)
	}
	fmt.Printf("fct3() -> %v\n", out)

	// Output:
	// fct1() -> 42
	// fct2() -> 46
	// result = 84
	// fct3() -> <nil>
}

func compileWast2Wasm(fname string) ([]byte, error) {
	switch fname {
	case "testdata/add-ex.wast":
		// obtained by running:
		//  $> wat2wasm -v -o add-ex.wasm add-ex.wast
		return ioutil.ReadFile("testdata/add-ex.wasm")
	case "testdata/add-ex-main.wast":
		// obtained by running:
		//  $> wat2wasm -v -o add-ex-main.wasm add-ex-main.wast
		return ioutil.ReadFile("testdata/add-ex-main.wasm")
	}
	return nil, fmt.Errorf("unknown wast test file %q", fname)
}
