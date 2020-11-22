package apu

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

var lengthTable = [2][16]int{
	{0x0a, 0x14, 0x28, 0x50, 0xa0, 0x3c, 0x0e, 0x1a, 0x0c, 0x18, 0x30, 0x60, 0xc0, 0x48, 0x10, 0x20},
	{0xfe, 0x02, 0x04, 0x06, 0x08, 0x0a, 0x0c, 0x0e, 0x10, 0x12, 0x14, 0x16, 0x18, 0x1a, 0x1c, 0x1e},
}

type APU struct {
	register       [0x16]uint8
	squarePlayers  [2]*audio.Player
	squareStreams  [2]*stream
	trianglePlayer *audio.Player
	triangleStream *stream
	cycle          int
	volumeRate     float64
}

func NewAPU(volume float64) *APU {
	apu := &APU{volumeRate: volume}
	apu.Init()
	return apu
}

func (a *APU) Init() {

	var audioContext = audio.NewContext(sampleRate)
	a.squareStreams[0] = NewStream(squareWave2, 800*a.volumeRate)
	a.squarePlayers[0], _ = audio.NewPlayer(audioContext, a.squareStreams[0])
	a.squarePlayers[0].Play()
	a.squareStreams[1] = NewStream(squareWave2, 800*a.volumeRate)
	a.squarePlayers[1], _ = audio.NewPlayer(audioContext, a.squareStreams[1])
	a.squarePlayers[1].Play()

	a.triangleStream = NewStream(triangleWave, 2000*a.volumeRate)
	a.trianglePlayer, _ = audio.NewPlayer(audioContext, a.triangleStream)
	a.trianglePlayer.Play()
	return

}

func (a *APU) Run(cycle int) {
	a.cycle += cycle
	if a.cycle >= 1200 {
		a.cycle -= 1200
		a.squareStreams[0].Time = math.Max(0, a.squareStreams[0].Time-1)
		a.squareStreams[1].Time = math.Max(0, a.squareStreams[1].Time-1)
		a.triangleStream.Time = math.Max(0, a.triangleStream.Time-1)
		sq1bytes := a.register[0x0001]
		if sq1bytes&0x80 != 0x00 {
			a.squareStreams[0].sweepCount -= 1
			if a.squareStreams[0].sweepCount < 1 {
				a.squareStreams[0].sweepCount = 0 * (int((sq1bytes&0x70)>>4) + 1)
				n := 1
				if (sq1bytes & 0x08) != 0 {
					n = -1
				}
				s := sq1bytes & 0x07
				f := (int(a.register[0x0003]&0x03) << 8) + int(a.register[0x0004])
				f = f + n*(f>>s)
				a.squareStreams[0].Frequency = 1790000 / int((f<<5)+1)
			}
		}
		sq2bytes := a.register[0x0005]
		if sq2bytes&0x80 != 0x00 {
			a.squareStreams[1].sweepCount -= 1
			if a.squareStreams[1].sweepCount < 1 {
				a.squareStreams[1].sweepCount = 10000 * (int((sq2bytes&0x70)>>4) + 1)
				n := 1
				if (sq2bytes & 0x08) != 0 {
					n = -1
				}
				s := sq2bytes & 0x07
				f := a.squareStreams[1].Frequency
				f = f + n*(f>>s)
				if f > 44100 {
					f = 44100
				}
				a.squareStreams[1].Frequency = f
			}
		}
	}

}

func (a *APU) Write(addr uint16, data uint8) {
	a.register[addr-0x4000] = data
	switch addr {
	//SQUARE 1
	case 0x4000:
		duty := (data & 0xc0) >> 6
		switch duty {
		case 0x00:
			a.squareStreams[0].function = squareWave0
		case 0x01:
			a.squareStreams[0].function = squareWave1
		case 0x02:
			a.squareStreams[0].function = squareWave2
		case 0x03:
			a.squareStreams[0].function = squareWave3
		}
	case 0x4002:
		n := (uint32(a.register[0x0003]&0x03) << 8) + uint32(data)
		a.squareStreams[0].Frequency = 1790000 / int((n<<5)+1)
	case 0x4003:
		n := (uint32(data&0x03) << 8) + uint32(a.register[0x0002])
		a.squareStreams[0].Frequency = 1790000 / int((n<<5)+1)

		bit3 := (data & 0x08) >> 3
		index := (data & 0xf0) >> 4
		a.squareStreams[0].Time = float64(lengthTable[bit3][index])
	//SQUARE 2
	case 0x4004:
		duty := (data & 0xc0) >> 6
		switch duty {
		case 0x00:
			a.squareStreams[1].function = squareWave0
		case 0x01:
			a.squareStreams[1].function = squareWave1
		case 0x02:
			a.squareStreams[1].function = squareWave2
		case 0x03:
			a.squareStreams[1].function = squareWave3
		}
	case 0x4006:
		n := (uint32(a.register[0x0007]&0x03) << 8) + uint32(data)
		a.squareStreams[1].Frequency = 1790000 / int((n<<5)+1)
	case 0x4007:
		n := (uint32(data&0x03) << 8) + uint32(a.register[0x0006])
		a.squareStreams[1].Frequency = 1790000 / int((n<<5)+1)

		bit3 := (data & 0x08) >> 3
		index := (data & 0xf0) >> 4
		a.squareStreams[1].Time = float64(lengthTable[bit3][index])
	//triangle
	case 0x4008:

	case 0x400a:
		n := (uint32(a.register[0x000b]&0x03) << 8) + uint32(data)
		a.triangleStream.Frequency = 1790000 / int((n<<6)+1)
	case 0x400b:
		n := (uint32(data&0x03) << 8) + uint32(a.register[0x000a])
		a.triangleStream.Frequency = 1790000 / int((n<<6)+1)

		bit3 := (data & 0x08) >> 3
		index := (data & 0xf0) >> 4
		a.triangleStream.Time = float64(lengthTable[bit3][index])
	//noise
	//active flag
	case 0x4015:
		if data&0x01 != 0x00 {
			a.squareStreams[0].IsActive = true
		}
		if data&0x02 != 0x00 {
			a.squareStreams[1].IsActive = true
		}
		if data&0x04 != 0x00 {
			a.triangleStream.IsActive = true
		}
	}
}

func (a *APU) Read(addr uint16) uint8 {
	return a.register[addr-0x4000]
}
