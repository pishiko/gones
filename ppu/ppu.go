package ppu

//TODO
//CtrlReg1 5
//CtrlReg2 2,1,0

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	nesColor = [64][3]uint8{
		{0x80, 0x80, 0x80}, {0x00, 0x3D, 0xA6}, {0x00, 0x12, 0xB0}, {0x44, 0x00, 0x96},
		{0xA1, 0x00, 0x5E}, {0xC7, 0x00, 0x28}, {0xBA, 0x06, 0x00}, {0x8C, 0x17, 0x00},
		{0x5C, 0x2F, 0x00}, {0x10, 0x45, 0x00}, {0x05, 0x4A, 0x00}, {0x00, 0x47, 0x2E},
		{0x00, 0x41, 0x66}, {0x00, 0x00, 0x00}, {0x05, 0x05, 0x05}, {0x05, 0x05, 0x05},
		{0xC7, 0xC7, 0xC7}, {0x00, 0x77, 0xFF}, {0x21, 0x55, 0xFF}, {0x82, 0x37, 0xFA},
		{0xEB, 0x2F, 0xB5}, {0xFF, 0x29, 0x50}, {0xFF, 0x22, 0x00}, {0xD6, 0x32, 0x00},
		{0xC4, 0x62, 0x00}, {0x35, 0x80, 0x00}, {0x05, 0x8F, 0x00}, {0x00, 0x8A, 0x55},
		{0x00, 0x99, 0xCC}, {0x21, 0x21, 0x21}, {0x09, 0x09, 0x09}, {0x09, 0x09, 0x09},
		{0xFF, 0xFF, 0xFF}, {0x0F, 0xD7, 0xFF}, {0x69, 0xA2, 0xFF}, {0xD4, 0x80, 0xFF},
		{0xFF, 0x45, 0xF3}, {0xFF, 0x61, 0x8B}, {0xFF, 0x88, 0x33}, {0xFF, 0x9C, 0x12},
		{0xFA, 0xBC, 0x20}, {0x9F, 0xE3, 0x0E}, {0x2B, 0xF0, 0x35}, {0x0C, 0xF0, 0xA4},
		{0x05, 0xFB, 0xFF}, {0x5E, 0x5E, 0x5E}, {0x0D, 0x0D, 0x0D}, {0x0D, 0x0D, 0x0D},
		{0xFF, 0xFF, 0xFF}, {0xA6, 0xFC, 0xFF}, {0xB3, 0xEC, 0xFF}, {0xDA, 0xAB, 0xEB},
		{0xFF, 0xA8, 0xF9}, {0xFF, 0xAB, 0xB3}, {0xFF, 0xD2, 0xB0}, {0xFF, 0xEF, 0xA6},
		{0xFF, 0xF7, 0x9C}, {0xD7, 0xE8, 0x95}, {0xA6, 0xED, 0xAF}, {0xA2, 0xF2, 0xDA},
		{0x99, 0xFF, 0xFC}, {0xDD, 0xDD, 0xDD}, {0x11, 0x11, 0x11}, {0x11, 0x11, 0x11},
	}
	bgColor = [4]color.Color{
		color.RGBA{0xff, 0x00, 0xff, 0xff}, color.RGBA{0x00, 0xff, 0x00, 0xff},
		color.RGBA{0x00, 0x00, 0xff, 0xff}, color.RGBA{0xff, 0x00, 0x00, 0xff},
	}
)

type PPU struct {
	////////////////////////////////////////////////////////////////
	//Reg,RAM

	ioRegister     [0x08]uint8
	chrRom         []uint8
	OAMAddr        uint8
	OAM            [0x0100]uint8
	isPPUAddrUp    bool
	PPUAddr        uint16
	vRAM           [0x4000]uint8
	statusRegister uint8
	ctrlReg1       uint8
	ctrlReg2       uint8
	ppuBuffer      uint8

	////////////////////////////////////////////////////////////////
	//other

	tiles            [][4]*ebiten.Image
	background       *ebiten.Image
	sprites          *ebiten.Image
	cycle            int
	line             int
	IsNMIOccured     bool
	scrollX          uint8
	scrollY          uint8
	isScrollCounterY bool
	//0->true
	isHorizontalMirror bool
	backgroundPallet   [4 * 0x0400]uint8
}

func NewPPU(chr []uint8, isHorizontalMirror bool) *PPU {
	p := &PPU{}
	p.chrRom = chr
	p.OAM = [0x0100]uint8{0}
	p.isPPUAddrUp = true
	p.vRAM = [0x4000]uint8{0}
	p.background = ebiten.NewImage(256, 240)
	p.sprites = ebiten.NewImage(256, 240)
	p.ctrlReg1 = 0x40
	p.isHorizontalMirror = isHorizontalMirror
	p.InitTiles()

	//fmt.Printf("[Init PPU] Character Size:0x%x\n", len(chr))
	return p
}

//Run は1画面が描画完了したらtrueを返す．
func (p *PPU) Run(cycle int) bool {
	p.cycle += cycle
	if p.cycle > 341 {
		p.cycle -= 341
		p.line++
		if p.line < 240 {
			if p.line == 0 {
				p.updateBGPallete()
			}
			if p.line%8 == 0 {
				p.drawBGLine()
			}
			p.drawSpLine()
		} else if p.line == 241 {
			//vblank set
			p.statusRegister = (p.statusRegister & 0x7f) + 0x80
			if p.ctrlReg1&0x80 != 0x00 {
				p.IsNMIOccured = true
			}
			return true
		} else if p.line == 262 {
			//0spritehit and vblank clear
			p.statusRegister = p.statusRegister & 0x3f
			p.resetBG()
			p.sprites.Clear()
			p.line = -1
		}
	}
	return false
}

func (p *PPU) updateBGPallete() {
	for i := 0; i < 4; i++ {
		head := 0x23c0 + i*0x400
		for py := 0; py < 8; py++ {
			for px := 0; px < 8; px++ {
				for block := 0; block < 4; block++ {
					color := (p.vRAM[head+py*8+px] >> (block * 2)) & 0x03
					blocky := block / 2
					blockx := block % 2
					index := py*16*8 + px*4 + blocky*32*2 + blockx*2
					p.backgroundPallet[i*0x400+index] = color
					p.backgroundPallet[i*0x400+index+1] = color
					p.backgroundPallet[i*0x400+index+32] = color
					p.backgroundPallet[i*0x400+index+32+1] = color
				}
			}
		}
	}
	return
}

func (p *PPU) resetBG() {
	cindex := 0
	switch p.ctrlReg2 >> 5 {
	case 0x00:
		cindex = 0
	case 0x01:
		cindex = 1
	case 0x02:
		cindex = 2
	case 0x04:
		cindex = 3
	}
	p.background.Fill(bgColor[cindex])
}

func (p *PPU) Draw() (bg *ebiten.Image, sprites *ebiten.Image) {
	if p.ctrlReg2&0x10 == 0x00 {
		p.sprites.Clear()
	}
	if p.ctrlReg2&0x08 == 0x00 {
		p.resetBG()
	}
	return p.background, p.sprites
}

func (p *PPU) drawBGLine() {
	tiley := p.line / 8

	tableNum := int(p.ctrlReg1 & 0x03)

	//Read Name Table
	var head int
	if int(p.scrollY)+p.line > 240 {
		//+0x800
		if tableNum/2 == 0 {
			head = tableNum*0x400 + 0x20*((p.line+int(p.scrollY)-240)/8) + 0x800
		} else {
			head = tableNum*0x400 + 0x20*((p.line+int(p.scrollY)-240)/8) - 0x800
		}
	} else {
		head = tableNum*0x400 + 0x20*((p.line+int(p.scrollY))/8)
	}
	xntOffset := int(p.scrollX / 8)

	nameTable := make([]uint8, 0, 32)
	nameTable = append(nameTable, p.vRAM[0x2000+head+xntOffset:0x2000+head+0x20]...)
	palletTable := make([]uint8, 0, 32)
	palletTable = append(palletTable, p.backgroundPallet[head+xntOffset:head+0x20]...)

	if tableNum%2 == 0 {
		nameTable = append(nameTable, p.vRAM[0x2000+head+0x400:0x2000+head+0x400+xntOffset]...)
		palletTable = append(palletTable, p.backgroundPallet[head+0x400:head+0x400+xntOffset]...)
	} else {
		nameTable = append(nameTable, p.vRAM[0x2000+head-0x400:0x2000+head-0x400+xntOffset]...)
		palletTable = append(palletTable, p.backgroundPallet[head-0x400:head-0x400+xntOffset]...)
	}

	//Read Pattern Table
	var bgPatternOffset int
	if p.ctrlReg1&0x10 != 0x00 {
		bgPatternOffset = 0x100
	} else {
		bgPatternOffset = 0x00
	}

	//BACKGROUND
	for tilex := 0; tilex < 0x20; tilex++ {
		pHead := 0x3f00 + int(palletTable[tilex])*4

		//0
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(tilex*8-int(p.scrollX%8)), float64(tiley*8-int(p.scrollY%8)))
		c := nesColor[p.vRAM[0x3f00]]
		op.ColorM.Scale(float64(c[0]), float64(c[1]), float64(c[2]), 1)
		p.background.DrawImage(p.tiles[int(nameTable[tilex])+bgPatternOffset][0], op)
		//1-3
		for i := 1; i < 4; i++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(tilex*8-int(p.scrollX%8)), float64(tiley*8-int(p.scrollY%8)))
			c := nesColor[p.vRAM[pHead+i]]
			op.ColorM.Scale(float64(c[0]), float64(c[1]), float64(c[2]), 1)
			p.background.DrawImage(p.tiles[int(nameTable[tilex])+bgPatternOffset][i], op)
		}
	}
	return
}

func (p *PPU) drawSpLine() {
	//SPRITES
	var spPatternOffset int
	if p.ctrlReg1&0x08 != 0x00 {
		spPatternOffset = 0x100
	} else {
		spPatternOffset = 0x00
	}
	spCounter := 0
	for i := 0; i < 64; i++ {
		y := int(p.OAM[i*4])
		tile := p.OAM[i*4+1]
		attr := p.OAM[i*4+2]
		x := p.OAM[i*4+3]
		if p.line == y {
			spCounter++
			if spCounter == 9 {
				p.statusRegister = (p.statusRegister & 0xdf) + 0x20
				break
			}

			//0 Bomb
			if i == 0 {
				p.statusRegister = 0x40 + (p.statusRegister & 0xbf)
			}

			pHead := 0x3f10 + int(attr&0x03)*4
			//01-11
			for j := 1; j < 4; j++ {
				c := nesColor[p.vRAM[pHead+j]]
				op := &ebiten.DrawImageOptions{}
				if attr&0x80 != 0x00 {
					op.GeoM.Scale(1, -1)
					op.GeoM.Translate(0, 8)
				}
				if attr&0x40 != 0x00 {
					op.GeoM.Scale(-1, 1)
					op.GeoM.Translate(8, 0)
				}
				op.GeoM.Translate(float64(x), float64(y+1))
				op.ColorM.Scale(float64(c[0]), float64(c[1]), float64(c[2]), 1)
				p.sprites.DrawImage(p.tiles[int(tile)+spPatternOffset][j], op)
			}
		}
	}
	if spCounter <= 8 {
		p.statusRegister &= 0xdf
	}
	return
}

func (p *PPU) writeVRAM(addr uint16, data uint8) {
	switch {
	//pattern Table 0,1
	case addr < 0x2000:
		p.vRAM[addr] = data

	//Table 0
	case addr < 0x2400:
		p.vRAM[addr] = data
		p.vRAM[addr+0x1000] = data
		if p.isHorizontalMirror {
			p.vRAM[addr+0x0400] = data
			p.vRAM[addr+0x1400] = data
		} else {
			p.vRAM[addr+0x0800] = data
			p.vRAM[addr+0x1800] = data
		}

	//Table 1
	case addr < 0x2800:
		p.vRAM[addr] = data
		p.vRAM[addr+0x1000] = data
		if p.isHorizontalMirror {
			p.vRAM[addr-0x0400] = data
			p.vRAM[addr-0x0400+0x1000] = data
		} else {
			p.vRAM[addr+0x0800] = data
			if addr+0x1800 < 0x3f00 {
				p.vRAM[addr+0x1800] = data //yabai
			}
		}

	//Table 2
	case addr < 0x2c00:
		p.vRAM[addr] = data
		p.vRAM[addr+0x1000] = data
		if p.isHorizontalMirror {
			p.vRAM[addr+0x0400] = data
			if addr+0x1400 < 0x3f00 {
				p.vRAM[addr+0x1400] = data //yabai
			}
		} else {
			p.vRAM[addr-0x800] = data
			p.vRAM[addr-0x800+0x1000] = data
		}

	//Table 3
	case addr < 0x3000:
		p.vRAM[addr] = data
		if addr+0x1000 < 0x3f00 {
			p.vRAM[addr+0x1000] = data //yabai
		}
		if p.isHorizontalMirror {
			p.vRAM[addr-0x0400] = data
			p.vRAM[addr-0x0400+0x1000] = data
		} else {
			p.vRAM[addr-0x800] = data
			p.vRAM[addr-0x800+0x1000] = data
		}
	//Name table mirror
	case addr < 0x3f00:
		p.writeVRAM(addr-0x1000, data)

	case addr < 0x3f20:
		var i uint16
		p.vRAM[addr] = data

		switch addr {
		case 0x3f10:
			p.vRAM[0x3f00] = data
		}
		for i = 0; i < 7; i++ {
			p.vRAM[addr+0x20+i*0x10] = data
		}
	case addr < 0x4000:
		p.writeVRAM(0x3f00+(addr%0x10), data)
	default:
		p.vRAM[addr] = data
	}
	return
}

func (p *PPU) WriteRegister(addr uint16, data uint8) {
	switch addr {
	case 0x2000:
		p.ctrlReg1 = data
	case 0x2001:
		p.ctrlReg2 = data
	case 0x2003:
		p.OAMAddr = data
	case 0x2004:
		p.OAM[p.OAMAddr] = data
		p.OAMAddr++
	case 0x2005:
		if !p.isScrollCounterY {
			p.scrollX = data
			p.isScrollCounterY = true
		} else {
			if data < 240 {
				p.scrollY = data
			} else {
				//p.scrollY = 0
			}
		}
	case 0x2006:
		if p.isPPUAddrUp {
			p.PPUAddr = uint16(data) << 8
		} else {
			p.PPUAddr += uint16(data)
		}
		p.isPPUAddrUp = !p.isPPUAddrUp
	case 0x2007:
		if p.PPUAddr < 0x2000 {
			p.chrRom[p.PPUAddr] = data
		} else {
			p.writeVRAM(p.PPUAddr, data)
		}
		if p.ctrlReg1&0x04 != 0x00 {
			p.PPUAddr += 32
		} else {
			p.PPUAddr++
		}
	default:
		//CANT REACH HERE!
	}
}

func (p *PPU) ReadRegister(addr uint16) uint8 {
	switch addr {
	case 0x2002:
		ret := p.statusRegister
		p.statusRegister &= 0x7f
		p.isScrollCounterY = false
		return ret
	case 0x2007:
		var ret uint8
		if p.PPUAddr < 0x2000 {
			ret = p.ppuBuffer
			p.ppuBuffer = p.chrRom[p.PPUAddr]
		} else if p.PPUAddr < 0x3f00 {
			ret = p.ppuBuffer
			p.ppuBuffer = p.vRAM[p.PPUAddr]
		} else {
			ret = p.vRAM[p.PPUAddr]
		}
		if p.ctrlReg1&0x04 != 0x00 {
			p.PPUAddr += 32
		} else {
			p.PPUAddr++
		}
		return ret
	}
	return 0x00
}

func (p *PPU) InitTiles() {
	tileSize := len(p.chrRom) / 16
	t := make([][4]*ebiten.Image, tileSize)

	//tile
	for i := 0; i < tileSize; i++ {
		out := [4][]uint8{}
		for j := 0; j < 4; j++ {
			out[j] = make([]uint8, 64*4)
		}
		// line
		for y := 0; y < 8; y++ {
			line0 := p.chrRom[i*16+y]
			line1 := p.chrRom[i*16+8+y]
			//dot
			for x := 0; x < 8; x++ {
				// px -> pallet index of 0-3
				px := (((line0 >> (7 - x)) & 0x01) + (((line1 >> (7 - x)) & 0x01) << 1))
				out[px][y*8*4+x*4+0] = 1
				out[px][y*8*4+x*4+1] = 1
				out[px][y*8*4+x*4+2] = 1
				out[px][y*8*4+x*4+3] = 0xff
			}
		}
		for j := 0; j < 4; j++ {
			t[i][j] = ebiten.NewImageFromImage(&image.RGBA{
				Pix:    out[j],
				Stride: 8 * 4,
				Rect:   image.Rect(0, 0, 8, 8),
			})
		}
	}
	p.tiles = t
	return
}
