package main

import (
	"reflect"
	"strings"

	"github.com/DeedleFake/wdte"
	"github.com/DeedleFake/wdte/std"
	_ "github.com/DeedleFake/wdte/std/all"
	"github.com/DeedleFake/wdte/wdteutil"
)

const src = `
	let io => import 'io';

	module 'test'
	-> name 'Test'
	;

	action 'hi'
	-> name 'Say Hi'
	-> perform (@ f v => 'Hi.' -- io.writeln io.stdout)
	;
`

type ModuleBuilder struct {
	ID   string
	Name string

	Actions []*Action
}

func (mb *ModuleBuilder) Call(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	return mb
}

type Action struct {
	ID      string
	Name    string
	Perform func() error
}

func (a *Action) Call(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	return a
}

func (mb *ModuleBuilder) funcModule(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	frame = frame.Sub("module")

	if len(args) < 1 {
		return wdteutil.SaveArgs(wdte.GoFunc(mb.funcModule), args...)
	}

	mb.ID = string(args[0].(wdte.String))
	return mb
}

func (mb *ModuleBuilder) funcName(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	frame = frame.Sub("name")

	if len(args) < 2 {
		return wdteutil.SaveArgsReverse(wdte.GoFunc(mb.funcName), args...)
	}

	reflect.ValueOf(args[0]).Elem().FieldByName("Name").SetString(string(args[1].(wdte.String)))
	return args[0]
}

func (mb *ModuleBuilder) funcAction(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	frame = frame.Sub("action")

	if len(args) < 1 {
		return wdteutil.SaveArgs(wdte.GoFunc(mb.funcAction), args...)
	}

	action := &Action{
		ID: string(args[0].(wdte.String)),
	}
	mb.Actions = append(mb.Actions, action)

	return action
}

func (mb *ModuleBuilder) funcPerform(frame wdte.Frame, args ...wdte.Func) wdte.Func {
	frame = frame.Sub("perform")

	if len(args) < 2 {
		return wdteutil.SaveArgsReverse(wdte.GoFunc(mb.funcPerform), args...)
	}

	reflect.ValueOf(args[0]).Elem().FieldByName("Perform").Set(reflect.ValueOf(func() error {
		err, _ := args[1].Call(frame, mb, args[0]).(error)
		return err
	}))
	return args[0]
}

func (mb *ModuleBuilder) Scope() *wdte.Scope {
	return wdte.S().Map(map[wdte.ID]wdte.Func{
		"module":  wdte.GoFunc(mb.funcModule),
		"name":    wdte.GoFunc(mb.funcName),
		"action":  wdte.GoFunc(mb.funcAction),
		"perform": wdte.GoFunc(mb.funcPerform),
	})
}

func main() {
	c, err := wdte.Parse(strings.NewReader(src), std.Import, nil)
	if err != nil {
		panic(err)
	}

	var module ModuleBuilder
	frame := wdte.F().WithScope(std.Scope.Sub(module.Scope()))

	r := c.Call(frame)
	if err, ok := r.(error); ok && (err != nil) {
		panic(err)
	}

	err = module.Actions[0].Perform()
	if err != nil {
		panic(err)
	}
}
