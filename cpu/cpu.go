package cpu

import (
	"fmt"

	"github.com/pishiko/gones/ppu"
)

var (
	cycles = [256]int{
		/*0x00*/ 7, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
		/*0x10*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 6, 7,
		/*0x20*/ 6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
		/*0x30*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 6, 7,
		/*0x40*/ 6, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
		/*0x50*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 6, 7,
		/*0x60*/ 6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
		/*0x70*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 6, 7,
		/*0x80*/ 2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
		/*0x90*/ 2, 6, 2, 6, 4, 4, 4, 4, 2, 4, 2, 5, 5, 4, 5, 5,
		/*0xA0*/ 2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
		/*0xB0*/ 2, 5, 2, 5, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4,
		/*0xC0*/ 2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
		/*0xD0*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
		/*0xE0*/ 2, 6, 3, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
		/*0xF0*/ 2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	}
)

//CPU CPU
type CPU struct {
	prgROM                 []uint8
	A, X, Y, SP            uint8
	PC                     uint16
	N, V, R, B, D, I, Z, C bool
	wRAM                   [0x0800]uint8
	opTable                [256]func(uint16)
	adrTable               [256]func() uint16
	PPU                    *ppu.PPU
	//
	keys           [8]bool
	isKeyReset     bool
	keyCounter     int
	addtionalCycle int
}

func (c *CPU) excute(opcode uint8) int {
	//fmt.Printf("[0x%x] %v ### %v\n", c.PC, runtime.FuncForPC(reflect.ValueOf(c.opTable[opcode]).Pointer()).Name(), runtime.FuncForPC(reflect.ValueOf(c.adrTable[opcode]).Pointer()).Name())
	c.opTable[opcode](c.adrTable[opcode]())
	a := c.addtionalCycle
	c.addtionalCycle = 0
	return cycles[opcode] + a
}

//NewCPU Constructer
func NewCPU(prg []uint8, keys [8]bool) *CPU {
	cpu := new(CPU)
	cpu.prgROM = prg
	cpu.SP = 0xFD
	cpu.wRAM = [0x0800]uint8{}
	cpu.R = true
	cpu.B = true
	cpu.I = true
	cpu.keys = keys

	cpu.opTable = [256]func(uint16){
		cpu.BRK, cpu.ORA, cpu.NOP, cpu.NOP, cpu.NOP, cpu.ORA, cpu.ASL, cpu.NOP, cpu.PHP, cpu.ORA, cpu.ASL, cpu.NOP, cpu.NOP, cpu.ORA, cpu.ASL, cpu.NOP,
		cpu.BPL, cpu.ORA, cpu.NOP, cpu.NOP, cpu.NOP, cpu.ORA, cpu.ASL, cpu.NOP, cpu.CLC, cpu.ORA, cpu.NOP, cpu.NOP, cpu.NOP, cpu.ORA, cpu.ASL, cpu.NOP,
		cpu.JSR, cpu.AND, cpu.NOP, cpu.NOP, cpu.BIT, cpu.AND, cpu.ROL, cpu.NOP, cpu.PLP, cpu.AND, cpu.ROL, cpu.NOP, cpu.BIT, cpu.AND, cpu.ROL, cpu.NOP,
		cpu.BMI, cpu.AND, cpu.NOP, cpu.NOP, cpu.NOP, cpu.AND, cpu.ROL, cpu.NOP, cpu.SEC, cpu.AND, cpu.NOP, cpu.NOP, cpu.NOP, cpu.AND, cpu.ROL, cpu.NOP,
		cpu.RTI, cpu.EOR, cpu.NOP, cpu.NOP, cpu.NOP, cpu.EOR, cpu.LSR, cpu.NOP, cpu.PHA, cpu.EOR, cpu.LSR, cpu.NOP, cpu.JMP, cpu.EOR, cpu.LSR, cpu.NOP,
		cpu.BVC, cpu.EOR, cpu.NOP, cpu.NOP, cpu.NOP, cpu.EOR, cpu.LSR, cpu.NOP, cpu.CLI, cpu.EOR, cpu.NOP, cpu.NOP, cpu.NOP, cpu.EOR, cpu.LSR, cpu.NOP,
		cpu.RTS, cpu.ADC, cpu.NOP, cpu.NOP, cpu.NOP, cpu.ADC, cpu.ROR, cpu.NOP, cpu.PLA, cpu.ADC, cpu.ROR, cpu.NOP, cpu.JMP, cpu.ADC, cpu.ROR, cpu.NOP,
		cpu.BVS, cpu.ADC, cpu.NOP, cpu.NOP, cpu.NOP, cpu.ADC, cpu.ROR, cpu.NOP, cpu.SEI, cpu.ADC, cpu.NOP, cpu.NOP, cpu.NOP, cpu.ADC, cpu.ROR, cpu.NOP,
		cpu.NOP, cpu.STA, cpu.NOP, cpu.NOP, cpu.STY, cpu.STA, cpu.STX, cpu.NOP, cpu.DEY, cpu.NOP, cpu.TXA, cpu.NOP, cpu.STY, cpu.STA, cpu.STX, cpu.NOP,
		cpu.BCC, cpu.STA, cpu.NOP, cpu.NOP, cpu.STY, cpu.STA, cpu.STX, cpu.NOP, cpu.TYA, cpu.STA, cpu.TXS, cpu.NOP, cpu.NOP, cpu.STA, cpu.NOP, cpu.NOP,
		cpu.LDY, cpu.LDA, cpu.LDX, cpu.NOP, cpu.LDY, cpu.LDA, cpu.LDX, cpu.NOP, cpu.TAY, cpu.LDA, cpu.TAX, cpu.NOP, cpu.LDY, cpu.LDA, cpu.LDX, cpu.NOP,
		cpu.BCS, cpu.LDA, cpu.NOP, cpu.NOP, cpu.LDY, cpu.LDA, cpu.LDX, cpu.NOP, cpu.CLV, cpu.LDA, cpu.TSX, cpu.NOP, cpu.LDY, cpu.LDA, cpu.LDX, cpu.NOP,
		cpu.CPY, cpu.CMP, cpu.NOP, cpu.NOP, cpu.CPY, cpu.CMP, cpu.DEC, cpu.NOP, cpu.INY, cpu.CMP, cpu.DEX, cpu.NOP, cpu.CPY, cpu.CMP, cpu.DEC, cpu.NOP,
		cpu.BNE, cpu.CMP, cpu.NOP, cpu.NOP, cpu.NOP, cpu.CMP, cpu.DEC, cpu.NOP, cpu.CLD, cpu.CMP, cpu.NOP, cpu.NOP, cpu.NOP, cpu.CMP, cpu.DEC, cpu.NOP,
		cpu.CPX, cpu.SBC, cpu.NOP, cpu.NOP, cpu.CPX, cpu.SBC, cpu.INC, cpu.NOP, cpu.INX, cpu.SBC, cpu.NOP, cpu.NOP, cpu.CPX, cpu.SBC, cpu.INC, cpu.NOP,
		cpu.BEQ, cpu.SBC, cpu.NOP, cpu.NOP, cpu.NOP, cpu.SBC, cpu.INC, cpu.NOP, cpu.SED, cpu.SBC, cpu.NOP, cpu.NOP, cpu.NOP, cpu.SBC, cpu.INC, cpu.NOP,
	}
	cpu.adrTable = [256]func() uint16{
		cpu.implied, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.accumulator, cpu.noAdressing, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.noAdressing,
		cpu.absolute, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.accumulator, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.noAdressing,
		cpu.implied, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.accumulator, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.noAdressing,
		cpu.implied, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.accumulator, cpu.noAdressing, cpu.indirect, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.noAdressing,
		cpu.noAdressing, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.noAdressing, cpu.implied, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.zeropageY, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.implied, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.noAdressing, cpu.noAdressing,
		cpu.immediate, cpu.Xindirect, cpu.immediate, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.implied, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.zeropageY, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.implied, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.absoluteY, cpu.noAdressing,
		cpu.immediate, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.implied, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.noAdressing,
		cpu.immediate, cpu.Xindirect, cpu.noAdressing, cpu.noAdressing, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.noAdressing, cpu.implied, cpu.immediate, cpu.implied, cpu.noAdressing, cpu.absolute, cpu.absolute, cpu.absolute, cpu.noAdressing,
		cpu.relative, cpu.indirectY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.zeropageX, cpu.zeropageX, cpu.noAdressing, cpu.implied, cpu.absoluteY, cpu.noAdressing, cpu.noAdressing, cpu.noAdressing, cpu.absoluteX, cpu.absoluteX, cpu.noAdressing,
	}

	fmt.Printf("[Init CPU] Program Size:0x%x\n", len(cpu.prgROM))
	cpu.RESET()
	return cpu
}

func (c *CPU) read(addr uint16) uint8 {
	switch {
	case addr < 0x2000:
		return c.wRAM[addr]
	case addr < 0x2008:
		return c.PPU.ReadRegister(addr)
	case addr < 0x4000:
		//PPU Mirror
	case addr < 0x4020:
		switch addr {
		//Joypad
		case 0x4016:
			if c.keys[c.keyCounter] {
				c.keyCounter = (c.keyCounter + 1) % 8
				return 0x01
			} else {
				c.keyCounter = (c.keyCounter + 1) % 8
				return 0x00
			}
		}
	case addr < 0xFFFF:
		return c.prgROM[addr-0x8000]
	}
	//CANT REACH HERE!
	return 0
}

func (c *CPU) write(addr uint16, data uint8) {
	switch {
	case addr < 0x2000:
		c.wRAM[addr] = data
	case addr < 0x2008:
		c.PPU.WriteRegister(addr, data)
	case addr < 0x4000:
		//PPU Mirror
	case addr < 0x4020:
		switch addr {
		//DMA
		case 0x4014:
			c.DMA(data)
		//Joypad
		case 0x4016:
			if data == 0x01 {
				c.isKeyReset = true
			} else if c.isKeyReset && data == 0x00 {
				for i := range c.keys {
					c.keys[i] = false
				}
				c.keyCounter = 0
				c.isKeyReset = false
			}
		}
	case addr < 0xFFFF:
		//prgrom
	}
	//CANT REACH HERE!
}

func (c *CPU) push(data uint8) {
	addr := 0x0100 + uint16(c.SP)
	c.write(addr, data)
	c.SP--
	return
}

func (c *CPU) pop() uint8 {
	c.SP++
	addr := 0x0100 + uint16(c.SP)
	return c.read(addr)
}

// Run 実行
func (c *CPU) Run(keys [8]bool) int {
	for k := range keys {
		if keys[k] {
			c.keys[k] = true
		}
	}
	opcode := c.read(c.PC)
	c.PC++
	return c.excute(opcode)
}

func (c *CPU) DMA(addrUp uint8) {
	addr := uint16(addrUp) << 8
	var i uint16
	for i = 0; i < 0x0100; i++ {
		c.PPU.OAM[i] = c.read(addr + i)
	}
	c.addtionalCycle += 514
}
