// +build tinygo.riscv,virt,qemu

package runtime

import (
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

// This file implements the VirtIO RISC-V interface implemented in QEMU, which
// is an interface designed for emulation.

type timeUnit int64

const tickMicros = 1

var timestamp timeUnit

func postinit() {}

//export main
func main() {
	preinit()
	run()
	abort()
}

const asyncScheduler = false

func sleepTicks(d timeUnit) {
	// TODO: actually sleep here for the given time.
	timestamp += d
}

func ticks() timeUnit {
	return timestamp
}

// Memory-mapped I/O as defined by QEMU.
// Source: https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
// Technically this is an implementation detail but hopefully they won't change
// the memory-mapped I/O registers.
var (
	// UART0 output register.
	stdoutWrite = (*volatile.Register8)(unsafe.Pointer(uintptr(0x10000000)))
	// SiFive test finisher
	testFinisher = (*volatile.Register16)(unsafe.Pointer(uintptr(0x100000)))
)

func putchar(c byte) {
	stdoutWrite.Set(uint8(c))
}

func abort() {
	// Make sure the QEMU process exits.
	testFinisher.Set(0x5555) // FINISHER_PASS

	// Lock up forever (as a fallback).
	for {
		riscv.Asm("wfi")
	}
}
