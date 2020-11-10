package cpu

func (c *CPU) setNZ(data uint8) {
	c.N = data&0x80 != 0x00
	c.Z = data == 0x00
	return
}

func (c *CPU) getP() (p uint8) {
	p = 0x00
	if c.N {
		p += 0x80
	}
	if c.V {
		p += 0x40
	}
	if c.R {
		p += 0x20
	}
	if c.B {
		p += 0x10
	}
	if c.D {
		p += 0x08
	}
	if c.I {
		p += 0x04
	}
	if c.Z {
		p += 0x02
	}
	if c.C {
		p += 0x01
	}
	return p
}

func (c *CPU) setP(p uint8) {
	c.N = p&0x80 != 0x00
	c.V = p&0x40 != 0x00
	c.R = p&0x20 != 0x00
	c.B = p&0x10 != 0x00
	c.D = p&0x08 != 0x00
	c.I = p&0x04 != 0x00
	c.Z = p&0x02 != 0x00
	c.C = p&0x01 != 0x00
	return
}

//演算
func (c *CPU) ADC(m uint16) {
	data := c.read(m)
	aFuture := c.A + data
	if c.C {
		aFuture++
	}

	c.C = c.A > aFuture
	c.V = (c.A^data) == 0x00 && (c.A^aFuture) != 0x00
	c.A = aFuture
	c.setNZ(c.A)
	return
}

func (c *CPU) SBC(m uint16) {
	data := c.read(m)
	aFuture := c.A - data
	if !c.C {
		aFuture--
	}

	c.C = c.A >= aFuture
	c.V = (c.A^data) == 0x00 && (c.A^aFuture) != 0x00
	c.A = aFuture
	c.setNZ(c.A)
	return
}

//論理演算
func (c *CPU) AND(m uint16) {
	c.A = c.A & c.read(m)
	c.setNZ(c.A)
	return
}

func (c *CPU) ORA(m uint16) {
	c.A = c.A | c.read(m)
	c.setNZ(c.A)
	return
}

func (c *CPU) EOR(m uint16) {
	c.A = c.A ^ c.read(m)
	c.setNZ(c.A)
	return
}

//シフト・ローテーション
func (c *CPU) ASL(_ uint16) {
	c.C = c.A&0x80 != 0x00
	c.A = c.A << 1
	c.setNZ(c.A)
	return
}

func (c *CPU) LSR(_ uint16) {
	c.C = c.A&0x01 != 0x00
	c.A = c.A >> 1
	c.setNZ(c.A)
	return
}

func (c *CPU) ROL(_ uint16) {
	aFuture := c.A << 1
	if c.C {
		aFuture++
	}
	c.C = c.A&0x01 != 0x00
	c.A = aFuture
	c.setNZ(c.A)
	return
}

func (c *CPU) ROR(_ uint16) {
	aFuture := c.A >> 1
	if c.C {
		aFuture += 0x80
	}
	c.C = c.A&0x01 != 0x00
	c.A = aFuture
	c.setNZ(c.A)
	return
}

//条件分岐
func (c *CPU) BCC(addr uint16) {
	if !c.C {
		c.PC = addr
	}
	return
}

func (c *CPU) BCS(addr uint16) {
	if c.C {
		c.PC = addr
	}
	return
}

func (c *CPU) BEQ(addr uint16) {
	if c.Z {
		c.PC = addr
	}
	return
}

func (c *CPU) BNE(addr uint16) {
	if !c.Z {
		c.PC = addr
	}
	return
}

func (c *CPU) BVC(addr uint16) {
	if !c.V {
		c.PC = addr
	}
	return
}

func (c *CPU) BVS(addr uint16) {
	if c.V {
		c.PC = addr
	}
	return
}

func (c *CPU) BPL(addr uint16) {
	if !c.N {
		c.PC = addr
	}
	return
}

func (c *CPU) BMI(addr uint16) {
	if c.N {
		c.PC = addr
	}
	return
}

//ビット検査
func (c *CPU) BIT(m uint16) {
	data := c.read(m)
	c.Z = c.A&data != 0x00
	c.N = data&0x80 != 0x00
	c.V = data&0x40 != 0x00
	return
}

//ジャンプ
func (c *CPU) JMP(addr uint16) {
	c.PC = addr
	return
}
func (c *CPU) JSR(addr uint16) {
	word := c.PC - 0x0001
	c.push(uint8(word >> 8))
	c.push(uint8(word & 0x00ff))
	c.PC = addr
	return
}
func (c *CPU) RTS(_ uint16) {
	wordL := c.pop()
	wordU := c.pop()
	c.PC = (uint16(wordU) << 8) + uint16(wordL) + 0x0001
	return
}

//ソフトウェア割込み
func (c *CPU) BRK(_ uint16) {
	if !c.I {
		c.B = true
		c.PC++
		c.push(uint8(c.PC >> 8))
		c.push(uint8(c.PC & 0x00ff))
		c.push(c.getP())
		c.I = true
		c.PC = (uint16(c.read(0xffff)) << 8) + uint16(c.read(0xfffe))
	}
	return
}

func (c *CPU) RTI(_ uint16) {
	c.setP(c.pop())
	wordL := c.pop()
	wordU := c.pop()
	c.PC = (uint16(wordU) << 8) + uint16(wordL)
	return
}

//比較
func (c *CPU) CMP(m uint16) {
	a := c.A - c.read(m)
	c.C = a >= c.A
	c.setNZ(a)
	return
}

func (c *CPU) CPX(m uint16) {
	a := c.X - c.read(m)
	c.C = a >= c.X
	c.setNZ(a)
	return
}

func (c *CPU) CPY(m uint16) {
	a := c.Y - c.read(m)
	c.C = a >= c.Y
	c.setNZ(a)
	return
}

func (c *CPU) INC(m uint16) {
	a := c.read(m) + 0x0001
	c.write(m, a)
	c.setNZ(a)
	return
}
func (c *CPU) DEC(m uint16) {
	a := c.read(m) - 0x0001
	c.write(m, a)
	c.setNZ(a)
	return
}
func (c *CPU) INX(m uint16) {
	c.X++
	c.setNZ(c.X)
	return
}
func (c *CPU) DEX(m uint16) {
	c.X--
	c.setNZ(c.X)
	return
}
func (c *CPU) INY(m uint16) {
	c.Y++
	c.setNZ(c.Y)
	return
}
func (c *CPU) DEY(m uint16) {
	c.Y--
	c.setNZ(c.Y)
	return
}

//フラグ操作
func (c *CPU) CLC(_ uint16) {
	c.C = false
	return
}
func (c *CPU) SEC(_ uint16) {
	c.C = true
	return
}
func (c *CPU) CLI(_ uint16) {
	c.I = false
	return
}
func (c *CPU) SEI(_ uint16) {
	c.I = true
	return
}
func (c *CPU) CLD(_ uint16) {
	c.D = false
	return
}
func (c *CPU) SED(_ uint16) {
	c.D = true
	return
}
func (c *CPU) CLV(_ uint16) {
	c.V = false
	return
}

//ロード
func (c *CPU) LDA(m uint16) {
	c.A = c.read(m)
	c.setNZ(c.A)
	return
}
func (c *CPU) LDX(m uint16) {
	c.X = c.read(m)
	c.setNZ(c.X)
	return
}
func (c *CPU) LDY(m uint16) {
	c.Y = c.read(m)
	c.setNZ(c.Y)
	return
}

//ストア
func (c *CPU) STA(m uint16) {
	c.write(m, c.A)
	return
}
func (c *CPU) STX(m uint16) {
	c.write(m, c.X)
	return
}
func (c *CPU) STY(m uint16) {
	c.write(m, c.Y)
	return
}

//レジスタ間転送
func (c *CPU) TAX(_ uint16) {
	c.X = c.A
	c.setNZ(c.X)
	return
}
func (c *CPU) TXA(_ uint16) {
	c.A = c.X
	c.setNZ(c.A)
	return
}
func (c *CPU) TAY(_ uint16) {
	c.Y = c.A
	c.setNZ(c.Y)
	return
}
func (c *CPU) TYA(_ uint16) {
	c.A = c.Y
	c.setNZ(c.A)
	return
}
func (c *CPU) TSX(_ uint16) {
	c.X = c.SP
	c.setNZ(c.X)
	return
}
func (c *CPU) TXS(_ uint16) {
	c.SP = c.X
	return
}

//スタック
func (c *CPU) PHA(_ uint16) {
	c.push(c.A)
	return
}
func (c *CPU) PLA(_ uint16) {
	c.A = c.pop()
	c.setNZ(c.A)
	return
}
func (c *CPU) PHP(_ uint16) {
	c.push(c.getP())
	return
}
func (c *CPU) PLP(_ uint16) {
	c.setP(c.pop())
	return
}

//No Operation
func (c *CPU) NOP(_ uint16) {
	return
}

//ハードウェア割り込み
func (c *CPU) NMI() {
	c.B = false
	c.push(uint8(c.PC >> 8))
	c.push(uint8(c.PC & 0x00ff))
	c.push(c.getP())
	c.I = true
	c.PC = (uint16(c.read(0xfffb)) << 8) + uint16(c.read(0xfffa))
}
func (c *CPU) IRQ() {
	if !c.I {
		c.B = false
		c.push(uint8(c.PC >> 8))
		c.push(uint8(c.PC & 0x00ff))
		c.push(c.getP())
		c.I = true
		c.PC = (uint16(c.read(0xffff)) << 8) + uint16(c.read(0xfffe))
	}
	return
}
func (c *CPU) RESET() {
	c.I = true
	if len(c.prgROM) < 0x8000 {
		c.PC = 0x8000
	} else {
		c.PC = (uint16(c.read(0xfffd)) << 8) + uint16(c.read(0xfffc))
	}
	return
}
