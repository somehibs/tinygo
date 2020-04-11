// Package compileopts contains the configuration for a single to-be-built
// binary.
package compileopts

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// Config keeps all configuration affecting the build in a single struct.
type Config struct {
	Options        *Options
	Target         *TargetSpec
	GoMinorVersion int
	ClangHeaders   string // Clang built-in header include path
	TestConfig     TestConfig
}

// FuncValueImplementation is an enum for the particular implementations of Go
// func values.
type FuncValueImplementation int

// These constants describe the various possible implementations of Go func
// values.
const (
	FuncValueNone FuncValueImplementation = iota

	// A func value is implemented as a pair of pointers:
	//     {context, function pointer}
	// where the context may be a pointer to a heap-allocated struct containing
	// the free variables, or it may be undef if the function being pointed to
	// doesn't need a context. The function pointer is a regular function
	// pointer.
	FuncValueDoubleword

	// As funcValueDoubleword, but with the function pointer replaced by a
	// unique ID per function signature. Function values are called by using a
	// switch statement and choosing which function to call.
	FuncValueSwitch
)

// Triple returns the LLVM target triple, like armv6m-none-eabi.
func (c *Config) Triple() string {
	return c.Target.Triple
}

// CPU returns the LLVM CPU name, like atmega328p or arm7tdmi. It may return an
// empty string if the CPU name is not known.
func (c *Config) CPU() string {
	return c.Target.CPU
}

// Features returns a list of features this CPU supports. For example, for a
// RISC-V processor, that could be ["+a", "+c", "+m"]. For many targets, an
// empty list will be returned.
func (c *Config) Features() []string {
	return c.Target.Features
}

// GOOS returns the GOOS of the target. This might not always be the actual OS:
// for example, bare-metal targets will usually pretend to be linux to get the
// standard library to compile.
func (c *Config) GOOS() string {
	return c.Target.GOOS
}

// GOARCH returns the GOARCH of the target. This might not always be the actual
// archtecture: for example, the AVR target is not supported by the Go standard
// library so such targets will usually pretend to be linux/arm.
func (c *Config) GOARCH() string {
	return c.Target.GOARCH
}

// BuildTags returns the complete list of build tags used during this build.
func (c *Config) BuildTags() []string {
	tags := append(c.Target.BuildTags, []string{"tinygo", "gc." + c.GC(), "scheduler." + c.Scheduler()}...)
	for i := 1; i <= c.GoMinorVersion; i++ {
		tags = append(tags, fmt.Sprintf("go1.%d", i))
	}
	if extraTags := strings.Fields(c.Options.Tags); len(extraTags) != 0 {
		tags = append(tags, extraTags...)
	}
	return tags
}

// CgoEnabled returns true if (and only if) CGo is enabled. It is true by
// default and false if CGO_ENABLED is set to "0".
func (c *Config) CgoEnabled() bool {
	return goenv.Get("CGO_ENABLED") == "1"
}

// GC returns the garbage collection strategy in use on this platform. Valid
// values are "none", "leaking", "extalloc", and "conservative".
func (c *Config) GC() string {
	if c.Options.GC != "" {
		return c.Options.GC
	}
	if c.Target.GC != "" {
		return c.Target.GC
	}
	for _, tag := range c.Target.BuildTags {
		if tag == "baremetal" || tag == "wasm" {
			return "conservative"
		}
	}
	return "extalloc"
}

// NeedsStackObjects returns true if the compiler should insert stack objects
// that can be traced by the garbage collector.
func (c *Config) NeedsStackObjects() bool {
	switch c.GC() {
	case "conservative", "extalloc":
		for _, tag := range c.BuildTags() {
			if tag == "baremetal" {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// Scheduler returns the scheduler implementation. Valid values are "coroutines"
// and "tasks".
func (c *Config) Scheduler() string {
	if c.Options.Scheduler != "" {
		return c.Options.Scheduler
	}
	if c.Target.Scheduler != "" {
		return c.Target.Scheduler
	}
	// Fall back to coroutines, which are supported everywhere.
	return "coroutines"
}

// FuncImplementation picks an appropriate func value implementation for the
// target.
func (c *Config) FuncImplementation() FuncValueImplementation {
	// Always pick the switch implementation, as it allows the use of blocking
	// inside a function that is used as a func value.
	switch c.Scheduler() {
	case "none", "coroutines":
		return FuncValueSwitch
	case "tasks":
		return FuncValueDoubleword
	default:
		panic("unknown scheduler type")
	}
}

// PanicStrategy returns the panic strategy selected for this target. Valid
// values are "print" (print the panic value, then exit) or "trap" (issue a trap
// instruction).
func (c *Config) PanicStrategy() string {
	return c.Options.PanicStrategy
}

// CFlags returns the flags to pass to the C compiler. This is necessary for CGo
// preprocessing.
func (c *Config) CFlags() []string {
	cflags := append([]string{}, c.Options.CFlags...)
	for _, flag := range c.Target.CFlags {
		cflags = append(cflags, strings.Replace(flag, "{root}", goenv.Get("TINYGOROOT"), -1))
	}
	if c.Target.Libc == "picolibc" {
		root := goenv.Get("TINYGOROOT")
		cflags = append(cflags, "--sysroot="+filepath.Join(root, "lib", "picolibc", "newlib", "libc"))
		cflags = append(cflags, "-I"+filepath.Join(root, "lib/picolibc-include"))
	}
	return cflags
}

// LDFlags returns the flags to pass to the linker. A few more flags are needed
// (like the one for the compiler runtime), but this represents the majority of
// the flags.
func (c *Config) LDFlags() []string {
	root := goenv.Get("TINYGOROOT")
	// Merge and adjust LDFlags.
	ldflags := append([]string{}, c.Options.LDFlags...)
	for _, flag := range c.Target.LDFlags {
		ldflags = append(ldflags, strings.Replace(flag, "{root}", root, -1))
	}
	ldflags = append(ldflags, "-L", root)
	if c.Target.GOARCH == "wasm" {
		// Round heap size to next multiple of 65536 (the WebAssembly page
		// size).
		heapSize := (c.Options.HeapSize + (65536 - 1)) &^ (65536 - 1)
		ldflags = append(ldflags, "--initial-memory="+strconv.FormatInt(heapSize, 10))
	}
	if c.Target.LinkerScript != "" {
		ldflags = append(ldflags, "-T", c.Target.LinkerScript)
	}
	return ldflags
}

// ExtraFiles returns the list of extra files to be built and linked with the
// executable. This can include extra C and assembly files.
func (c *Config) ExtraFiles() []string {
	return c.Target.ExtraFiles
}

// DumpSSA returns whether to dump Go SSA while compiling (-dumpssa flag). Only
// enable this for debugging.
func (c *Config) DumpSSA() bool {
	return c.Options.DumpSSA
}

// VerifyIR returns whether to run extra checks on the IR. This is normally
// disabled but enabled during testing.
func (c *Config) VerifyIR() bool {
	return c.Options.VerifyIR
}

// Debug returns whether to add debug symbols to the IR, for debugging with GDB
// and similar.
func (c *Config) Debug() bool {
	return c.Options.Debug
}

// Programmer returns the flash method and OpenOCD interface name given a
// particular configuration. It may either be all configured in the target JSON
// file or be modified using the -programmmer command-line option.
func (c *Config) Programmer() (method, openocdInterface string) {
	switch c.Options.Programmer {
	case "":
		// No configuration supplied.
		return c.Target.FlashMethod, c.Target.OpenOCDInterface
	case "openocd", "msd", "command":
		// The -programmer flag only specifies the flash method.
		return c.Options.Programmer, c.Target.OpenOCDInterface
	default:
		// The -programmer flag specifies something else, assume it specifies
		// the OpenOCD interface name.
		return "openocd", c.Options.Programmer
	}
}

// OpenOCDConfiguration returns a list of command line arguments to OpenOCD.
// This list of command-line arguments is based on the various OpenOCD-related
// flags in the target specification.
func (c *Config) OpenOCDConfiguration() (args []string, err error) {
	_, openocdInterface := c.Programmer()
	if openocdInterface == "" {
		return nil, errors.New("OpenOCD programmer not set")
	}
	if !regexp.MustCompile("^[\\p{L}0-9_-]+$").MatchString(openocdInterface) {
		return nil, fmt.Errorf("OpenOCD programmer has an invalid name: %#v", openocdInterface)
	}
	if c.Target.OpenOCDTarget == "" {
		return nil, errors.New("OpenOCD chip not set")
	}
	if !regexp.MustCompile("^[\\p{L}0-9_-]+$").MatchString(c.Target.OpenOCDTarget) {
		return nil, fmt.Errorf("OpenOCD target has an invalid name: %#v", c.Target.OpenOCDTarget)
	}
	if c.Target.OpenOCDTransport != "" && c.Target.OpenOCDTransport != "swd" {
		return nil, fmt.Errorf("unknown OpenOCD transport: %#v", c.Target.OpenOCDTransport)
	}
	args = []string{"-f", "interface/" + openocdInterface + ".cfg"}
	if c.Target.OpenOCDTransport != "" {
		args = append(args, "-c", "transport select "+c.Target.OpenOCDTransport)
	}
	args = append(args, "-f", "target/"+c.Target.OpenOCDTarget+".cfg")
	return args, nil
}

type TestConfig struct {
	CompileTestBinary bool
	// TODO: Filter the test functions to run, include verbose flag, etc
}
