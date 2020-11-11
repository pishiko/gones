package cpu

func (c *CPU) _read16(addr uint16) uint16 {
	return (uint16(c.read(addr+0x0001)) << 8) + uint16(c.read(addr))
}

func uint2int(n uint8) int {
	return int(n&0x7f) - int(n>>7)*128
}

func (c *CPU) accumulator() uint16 {
	c.isNoAddrOP = true
	return 0x0000
}

func (c *CPU) implied() uint16 {
	c.isNoAddrOP = true
	return 0x0000
}

func (c *CPU) immediate() uint16 {
	c.PC++
	return c.PC - 0x0001
}

func (c *CPU) zeropage() uint16 {
	c.PC++
	return uint16(c.read(c.PC - 0x0001))
}

func (c *CPU) zeropageX() uint16 {
	c.PC++
	return uint16(c.read(c.PC-0x0001) + c.X)
}

func (c *CPU) zeropageY() uint16 {
	c.PC++
	return uint16(c.read(c.PC-0x0001) + c.Y)
}

func (c *CPU) absolute() uint16 {
	c.PC += 0x0002
	return c._read16(c.PC - 0x0002)
}

func (c *CPU) absoluteX() uint16 {
	c.PC += 0x0002
	return c._read16(c.PC-0x0002) + uint16(c.X)
}

func (c *CPU) absoluteY() uint16 {
	c.PC += 0x0002
	return c._read16(c.PC-0x0002) + uint16(c.Y)
}

func (c *CPU) indirect() uint16 {
	c.PC += 0x0002
	addrUp := c.read(c.PC - 0x0001)
	addrLow := c.read(c.PC - 0x0002)

	return (uint16(c.read((uint16(addrUp)<<8)+uint16(addrLow+0x01))) << 8) + uint16(c.read((uint16(addrUp)<<8)+uint16(addrLow)))
}

func (c *CPU) Xindirect() uint16 {
	c.PC++
	addr := c.read(c.PC-0x0001) + c.X
	return (uint16(c.read(uint16(addr+0x01))) << 8) + uint16(c.read(uint16(addr)))
}

func (c *CPU) indirectY() uint16 {
	c.PC++
	addr := c.read(c.PC - 0x0001)
	return uint16(c.read(uint16(addr+0x01)))<<8 + uint16(c.read(uint16(addr))) + uint16(c.Y)
}

func (c *CPU) relative() uint16 {
	c.PC++
	return uint16(int(c.PC) + uint2int(c.read(c.PC-0x0001)))
}

func (c *CPU) noAdressing() uint16 {
	println("[NO ADRESSING] CANT REACH HERE")
	return 0x0000
}
