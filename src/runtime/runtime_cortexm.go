// +build cortexm

package runtime

import (
	"device/arm"
	"unsafe"
)

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//go:extern _sdata
var _sdata [0]byte

//go:extern _sidata
var _sidata [0]byte

//go:extern _edata
var _edata [0]byte

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Pointer(uintptr(dst) + 4)
		src = unsafe.Pointer(uintptr(src) + 4)
	}
}

func abort() {
	// disable all interrupts
	arm.DisableInterrupts()

	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

// The stack layout at the moment an interrupt occurs.
// Registers can be accessed if the stack pointer is cast to a pointer to this
// struct.
type interruptStack struct {
	R0  uintptr
	R1  uintptr
	R2  uintptr
	R3  uintptr
	R12 uintptr
	LR  uintptr
	PC  uintptr
	PSR uintptr
}

// This function is called at HardFault.
// Before this function is called, the stack pointer is reset to the initial
// stack pointer (loaded from addres 0x0) and the previous stack pointer is
// passed as an argument to this function. This allows for easy inspection of
// the stack the moment a HardFault occurs, but it means that the stack will be
// corrupted by this function and thus this handler must not attempt to recover.
//
// For details, see:
// https://community.arm.com/developer/ip-products/system/f/embedded-forum/3257/debugging-a-cortex-m0-hard-fault
// https://blog.feabhas.com/2013/02/developing-a-generic-hard-fault-handler-for-arm-cortex-m3cortex-m4/
//export handleHardFault
func handleHardFault(sp *interruptStack) {
	print("fatal error: ")
	if uintptr(unsafe.Pointer(sp)) < 0x20000000 {
		print("stack overflow")
	} else {
		// TODO: try to find the cause of the hard fault. Especially on
		// Cortex-M3 and higher it is possible to find more detailed information
		// in special status registers.
		print("HardFault")
	}
	print(" with sp=", sp)
	if uintptr(unsafe.Pointer(&sp.PC)) >= 0x20000000 {
		// Only print the PC if it points into memory.
		// It may not point into memory during a stack overflow, so check that
		// first before accessing the stack.
		print(" pc=", sp.PC)
	}
	println()
	abort()
}
