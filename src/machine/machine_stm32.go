// +build stm32

package machine

// Peripheral abstraction layer for the stm32.

type PinMode uint8

// Peripheral operations sequence:
//  1. Enable the clock to the alternate function.
//  2. Enable clock to corresponding GPIO
//  3. Attach the alternate function.
//  4. Configure the input-output port and pins (of the corresponding GPIOx) to match the AF .
//  5. If desired enable the nested vector interrupt control to generate interrupts.
//  6. Program the AF/peripheral for the required configuration (eg baud rate for a USART) .

// Given that the stm32 family has the AF and GPIO on different registers based on the chip,
//  use the main function here for configuring, and use hooks in the more specific chip
//  definition files
// Also, the stm32f1xx series handles things differently from the stm32f0/2/3/4

// ---------- General pin operations ----------

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port := p.getPort()
	pin := uint8(p) % 16
	if high {
		port.BSRR.Set(1 << pin)
	} else {
		port.BSRR.Set(1 << (pin + 16))
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	port := p.getPort()
	pin := uint8(p) % 16
	val := port.IDR.Get() & (1 << pin)
	return (val > 0)
}
