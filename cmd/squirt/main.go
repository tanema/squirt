package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

	"github.com/tanema/squirt/src/excerpt"
	"github.com/tanema/squirt/src/lang"
	"github.com/tanema/squirt/src/runtime"
	"github.com/tanema/squirt/src/stdlib"
)

var astPtr = flag.Bool("ast", false, "a bool")
var expPtr = flag.Bool("excerpt", false, "a bool")

func main() {
	flag.Parse()
	args := flag.Args()
	scope := runtime.DefaultNamespace(nil)
	runtime.RegisterLib("os", stdlib.OSLib)
	if len(args) > 0 {
		if *astPtr {
			ast(args[0])
		} else if *expPtr {
			exp(args[0])
		} else {
			runFile(scope, args[0], args[1:]...)
		}
	} else {
		runREPL(scope)
	}
}

func ast(path string) {
	log.SetFlags(0)
	block, err := lang.ParseFile(path)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(block)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}

func exp(path string) {
	block, err := lang.ParseFile(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(excerpt.File(path, block.Pos))
}

func runFile(e *runtime.Scope, path string, argv ...string) {
	targv := make([]runtime.Value, len(argv))
	for i, arg := range argv {
		targv[i], _ = runtime.ToValue(e, arg)
	}
	argvTable, _ := runtime.ToValue(e, targv)
	e.Set("ARGV", argvTable)
	if _, err := runtime.EvalFile(e, os.Args[1]); err != nil {
		fmt.Println(runtime.Print(e, err))
	}
}

func runREPL(scope *runtime.Scope) error {
	readline.SetHistoryPath(filepath.Join(os.Getenv("HOME"), ".squirt-history"))
	rl, err := readline.New("> ")
	if err != nil {
		return err
	}

	keepRunning := true
	tbl, _ := runtime.ToValue(scope, []runtime.Value{})
	scope.Set("ARGV", tbl)
	scope.Set("exit", runtime.Fn("exit", func(e *runtime.Scope, self runtime.CVal, a []runtime.Value) (runtime.Value, error) {
		keepRunning = false
		return nil, nil
	}))
	fmt.Println("squirt 0.1 get squirty")
	for keepRunning {
		text, err := rl.Readline()
		if err != nil {
			fmt.Println(runtime.Print(scope, err))
			break
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		val, err := runtime.Eval(scope, text)
		if err != nil {
			fmt.Println(runtime.Print(scope, err))
		} else if val != nil {
			fmt.Println(runtime.Print(scope, val))
		}
	}
	return nil
}
