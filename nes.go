package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/pishiko/gones/cpu"
	"github.com/pishiko/gones/ppu"
)

var (
	//A,B,Select,Start,Up,Down,Left,Right
	keymap = []ebiten.Key{
		ebiten.KeyX,
		ebiten.KeyZ,
		ebiten.KeyTab,
		ebiten.KeyEscape,
		ebiten.KeyUp,
		ebiten.KeyDown,
		ebiten.KeyLeft,
		ebiten.KeyRight,
	}
)

type NES struct {
	cpu    *cpu.CPU
	ppu    *ppu.PPU
	canvas *ebiten.Image
	keys   [8]bool
}

func NewNES(prg []uint8, chr []uint8) *NES {
	n := new(NES)
	n.keys = [8]bool{}
	n.cpu = cpu.NewCPU(prg, n.keys)
	n.ppu = ppu.NewPPU(chr)
	n.cpu.PPU = n.ppu
	n.canvas = ebiten.NewImage(256, 240)
	return n
}

//////////////////////
//ebiten functions

//Draw はPPUから画面データを受け取り描画
func (n *NES) Draw(screen *ebiten.Image) {

	background, sprites := n.ppu.Draw()
	n.canvas.DrawImage(background, nil)
	n.canvas.DrawImage(sprites, nil)
	ebitenutil.DebugPrint(n.canvas, fmt.Sprintf("FPS:%0.2f", ebiten.CurrentFPS()))

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(3, 3)
	screen.DrawImage(n.canvas, op)
	return
}

func (n *NES) Layout(screenWidth, screenHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (n *NES) Update() error {

	for k := 0; k < 8; k++ {
		n.keys[k] = ebiten.IsKeyPressed(keymap[k])
	}
	isScreenReady := false
	for !isScreenReady {
		cycle := n.cpu.Run(n.keys)
		isScreenReady = n.ppu.Run(cycle * 3)
	}
	return nil
}

//////////////////

func (n *NES) init() {
	ebiten.SetWindowSize(256*3, 240*3)
	//n.DrawTile(n.ppu.GetTiles())
}

func (n *NES) Run() {
	n.init()
	if err := ebiten.RunGame(n); err != nil {
		log.Fatal(err)
	}
}

func (n *NES) DrawTile(sprites [][]uint8) {
	n.canvas = ebiten.NewImage(256, 240)
	n.canvas.Fill(color.Black)
	for i := 0; i < len(sprites); i++ {
		image := ebiten.NewImageFromImage(&image.RGBA{
			Pix:    sprites[i],
			Stride: 8 * 4,
			Rect:   image.Rect(0, 0, 8, 8),
		})
		y := i / 32
		x := i - y*32
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*8), float64(y*8))
		n.canvas.DrawImage(image, op)
	}
}
