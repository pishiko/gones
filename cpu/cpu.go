package cpu

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/pishiko/gones/apu"
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
	debugCounter = 7
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
	ppu                    *ppu.PPU
	apu                    *apu.APU
	//
	keys           [8]bool
	isKeyReset     bool
	keyCounter     int
	addtionalCycle int
	isNoAddrOP     bool
	//DEBUG
	IsRecord bool
	DebugLog string
}

//NewCPU Constructer
func NewCPU(prg []uint8, ppu *ppu.PPU, apu *apu.APU) *CPU {
	cpu := new(CPU)
	cpu.prgROM = prg
	cpu.SP = 0xFD
	cpu.wRAM = [0x0800]uint8{}
	cpu.R = true
	cpu.B = false
	cpu.I = true
	cpu.ppu = ppu
	cpu.apu = apu

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
		/*0x00*/ cpu.implied, cpu.Xindirect, cpu.implied, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.accumulator, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0x10*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX,
		/*0x20*/ cpu.absolute, cpu.Xindirect, cpu.implied, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.accumulator, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0x30*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX,
		/*0x40*/ cpu.implied, cpu.Xindirect, cpu.implied, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.accumulator, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0x50*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX,
		/*0x60*/ cpu.implied, cpu.Xindirect, cpu.implied, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.accumulator, cpu.immediate, cpu.indirect, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0x70*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX,
		/*0x80*/ cpu.immediate, cpu.Xindirect, cpu.immediate, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.implied, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0x90*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageY, cpu.zeropageY, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteY, cpu.absoluteY,
		/*0xA0*/ cpu.immediate, cpu.Xindirect, cpu.immediate, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.implied, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0xB0*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageY, cpu.zeropageY, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteY, cpu.absoluteY,
		/*0xC0*/ cpu.immediate, cpu.Xindirect, cpu.immediate, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.implied, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0xD0*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX,
		/*0xE0*/ cpu.immediate, cpu.Xindirect, cpu.immediate, cpu.Xindirect, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.zeropage, cpu.implied, cpu.immediate, cpu.implied, cpu.immediate, cpu.absolute, cpu.absolute, cpu.absolute, cpu.absolute,
		/*0xF0*/ cpu.relative, cpu.indirectY, cpu.implied, cpu.indirectY, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.zeropageX, cpu.implied, cpu.absoluteY, cpu.implied, cpu.absoluteY, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX, cpu.absoluteX,
	}

	//fmt.Printf("[Init CPU] Program Size:0x%x\n", len(cpu.prgROM))
	cpu.RESET()
	return cpu
}

func (c *CPU) GetDebugText() string {
	// opcode := c.prgROM[c.PC-0x8000-(0x8000-uint16(len(c.prgROM)))]
	// return fmt.Sprintf("$%04X %02X %16s\nA:%02X X:%02X Y:%02X P:%02X SP:%02X\n", c.PC-1, opcode,
	// 	strings.Replace(strings.Replace(runtime.FuncForPC(reflect.ValueOf(c.opTable[opcode]).Pointer()).Name()+runtime.FuncForPC(reflect.ValueOf(c.adrTable[opcode]).Pointer()).Name(), "github.com/pishiko/gones/cpu.(*CPU).", "", 2), "fm", "", 2),
	// 	c.A, c.X, c.Y, c.getP(), c.SP)
	return fmt.Sprintf("$57:%X $86:%X $45:%X\n", c.wRAM[0x57], c.wRAM[0x86], c.wRAM[0x45])
}

func (c *CPU) excute(opcode uint8) int {
	if c.IsRecord {
		// if debugCounter >= 1000000 {
		// 	os.Exit(0)
		// }

		c.DebugLog += fmt.Sprintf("%04X %02X %16s A:%02X X:%02X Y:%02X P:%02X SP:%02X\n", c.PC-1, opcode,
			strings.Replace(strings.Replace(runtime.FuncForPC(reflect.ValueOf(c.opTable[opcode]).Pointer()).Name()+runtime.FuncForPC(reflect.ValueOf(c.adrTable[opcode]).Pointer()).Name(),
				"github.com/pishiko/gones/cpu.(*CPU).", "", 2), "fm", "", 2),
			c.A, c.X, c.Y, c.getP(), c.SP)
		// c.DebugLog += fmt.Sprintf("%04X A:%02X X:%02X Y:%02X P:%02X SP:%02X\n", c.PC-1,
		// 	c.A, c.X, c.Y, c.getP(), c.SP)

		// fmt.Printf("%0X\n", c.PC-1)
	}

	c.isNoAddrOP = false
	c.opTable[opcode](c.adrTable[opcode]())

	a := c.addtionalCycle
	c.addtionalCycle = 0
	cycle := cycles[opcode] + a
	debugCounter += cycle
	return cycle
}

func (c *CPU) read(addr uint16) uint8 {
	switch {
	case addr < 0x0800:
		return c.wRAM[addr]
	case addr < 0x2000:
		return c.wRAM[addr%0x0800]
	case addr < 0x2008:
		return c.ppu.ReadRegister(addr)
	case addr < 0x4000:
		fmt.Println("PPUMIRROR")
	case addr < 0x4020:
		switch addr {
		//Joypad 1
		case 0x4016:
			if c.keys[c.keyCounter] {
				c.keyCounter = (c.keyCounter + 1) % 8
				return 0x01
			} else {
				c.keyCounter = (c.keyCounter + 1) % 8
				return 0x00
			}
		//Joypad 2
		case 0x4017:
			//
		default:
			return c.apu.Read(addr)
		}
	case addr < 0xbfff:
		return c.prgROM[addr-0x8000]
	case addr < 0xffff:
		//Mapper 0 Mirror
		return c.prgROM[addr-0x8000-(0x8000-uint16(len(c.prgROM)))]
	}
	//CANT REACH HERE!
	return 0
}

func (c *CPU) write(addr uint16, data uint8) {
	switch {
	case addr < 0x0800:
		c.wRAM[addr] = data
	case addr < 0x2000:
		c.wRAM[addr%0x0800] = data
	case addr < 0x2008:
		c.ppu.WriteRegister(addr, data)
	case addr < 0x4000:
		fmt.Println("PPU MIRROR WRITE")
	case addr < 0x4020:
		switch addr {
		//DMA
		case 0x4014:
			c.DMA(data)
		//Joypad 1
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
		//Joypad 2
		case 0x4017:
			//
		default:
			c.apu.Write(addr, data)
		}
	case addr < 0xFFFF:
		fmt.Printf("#WPRG %x %x\n", addr, data)
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
	if c.ppu.IsNMIOccured {
		c.ppu.IsNMIOccured = false
		c.NMI()
	}
	opcode := c.read(c.PC)
	c.PC++
	return c.excute(opcode)
}

func (c *CPU) DMA(addrUp uint8) {
	addr := uint16(addrUp) << 8
	var i uint16
	for i = 0; i < 0x0100; i++ {
		c.ppu.OAM[i] = c.read(addr + i)
	}
	c.addtionalCycle += 514
}
func (c *CPU) RESET() {
	c.I = true
	//c.PC = 0xc000
	c.PC = (uint16(c.read(0xfffd+(uint16(len(c.prgROM))-0x8000))) << 8) + uint16(c.read(0xfffc+(uint16(len(c.prgROM))-0x8000)))

	return
}
