package main

type Thing struct {
	name string
}

type ThingOption func(*Thing)

func WithName(name string) ThingOption {
	return func(t *Thing) {
		t.name = name
	}
}

func NewThing(opts ...ThingOption) *Thing {
	t := &Thing{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t Thing) String() string {
	return t.name
}

func (t Thing) Print(arg string) {
	println("Thing.Print:", t.name, "arg:", arg)
}

type Printer interface {
	Print(string)
}

func main() {
	thing := &Thing{"foo"}

	// function pointers
	runFunc(hello, 5) // must be indirect to avoid obvious inlining

	// deferred functions
	testDefer()

	// defers in loop
	testDeferLoop()

	// Take a bound method and use it as a function pointer.
	// This function pointer needs a context pointer.
	testBound(thing.String)

	// closures
	func() {
		println("thing inside closure:", thing.String())
	}()
	runFunc(func(i int) {
		println("inside fp closure:", thing.String(), i)
	}, 3)

	// functional arguments
	thingFunctionalArgs1 := NewThing()
	thingFunctionalArgs1.Print("functional args 1")
	thingFunctionalArgs2 := NewThing(WithName("named thing"))
	thingFunctionalArgs2.Print("functional args 2")

	// regression testing
	regression1033()
}

func runFunc(f func(int), arg int) {
	f(arg)
}

func hello(n int) {
	println("hello from function pointer:", n)
}

func testDefer() {
	defer exportedDefer()

	i := 1
	defer deferred("...run as defer", i)
	i++
	defer func() {
		println("...run closure deferred:", i)
	}()
	i++
	defer deferred("...run as defer", i)
	i++

	var t Printer = &Thing{"foo"}
	defer t.Print("bar")

	println("deferring...")
}

func testDeferLoop() {
	for j := 0; j < 4; j++ {
		defer deferred("loop", j)
	}
}

func deferred(msg string, i int) {
	println(msg, i)
}

//export __exportedDefer
func exportedDefer() {
	println("...exported defer")
}

func testBound(f func() string) {
	println("bound method:", f())
}

// regression1033 is a regression test for https://github.com/tinygo-org/tinygo/issues/1033.
// In previous versions of the compiler, a deferred call to an interface would create an instruction that did not dominate its uses.
func regression1033() {
	foo(&Bar{})
}

type Bar struct {
	empty bool
}

func (b *Bar) Close() error {
	return nil
}

type Closer interface {
	Close() error
}

func foo(bar *Bar) error {
	var a int
	if !bar.empty {
		a = 10
		if a != 5 {
			return nil
		}
	}

	var c Closer = bar
	defer c.Close()

	return nil
}
