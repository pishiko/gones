package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Need NES ROM.")
		return
	}
	nes := NewNES(os.Args[1])
	for _, a := range os.Args {
		if a == "--debug" || a == "-d" {
			nes.SetDebug()
		}
	}
	nes.Run()
}
