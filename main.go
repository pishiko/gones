package main

import (
	"io/ioutil"
)

func Load(path string) ([]uint8, []uint8) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	prgSize := int(bytes[4]) * 0x4000
	chrSize := int(bytes[5]) * 0x2000

	prgRom := make([]uint8, prgSize)
	chrRom := make([]uint8, chrSize)
	prgRom = bytes[16 : 16+prgSize]
	chrRom = bytes[16+prgSize : 16+prgSize+chrSize]
	return prgRom, chrRom
}

func main() {
	prg, chr := Load("roms/giko011.nes")

	nes := NewNES(prg, chr)
	nes.Run()

	// //CPU時間計測
	// fmt.Printf("[ROM Size] chr:%d\n", len(chr))
	// start := time.Now()
	// c := cpu.NewCPU(prg)
	// for counter := 0; counter < 10000000; {
	// 	counter += c.Run()
	// }
	// fmt.Printf("%.0f Cycle/s\n", 10000000/time.Now().Sub(start).Seconds())
}
