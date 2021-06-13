package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/pishiko/gones/apu"
	"github.com/pishiko/gones/cpu"
	"github.com/pishiko/gones/ppu"
)

var (
	// A,B,Select,Start,Up,Down,Left,Right
	keymap = []ebiten.Key{
		ebiten.KeyL,
		ebiten.KeyK,
		ebiten.KeyO,
		ebiten.KeyP,
		ebiten.KeyW,
		ebiten.KeyS,
		ebiten.KeyA,
		ebiten.KeyD,
	}
	//A,B,Select,Start
	padmap = []ebiten.GamepadButton{
		ebiten.GamepadButton2,
		ebiten.GamepadButton0,
		ebiten.GamepadButton6,
		ebiten.GamepadButton7,
	}
	GAMEPAD_AXIS_X = 0
	GAMEPAD_AXIS_Y = 4
	pauseBG        *ebiten.Image
	pauseOP        *ebiten.DrawImageOptions
)

type NES struct {
	cpu       *cpu.CPU
	ppu       *ppu.PPU
	apu       *apu.APU
	canvas    *ebiten.Image
	keys      [8]bool
	gamepadID ebiten.GamepadID
	//interface
	isDebug          bool
	isPlay           bool
	isRecording      bool
	isGamepadEnabled bool
}

func Load(path string) ([]uint8, []uint8, []uint8) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	prgSize := int(bytes[4]) * 0x4000
	chrSize := int(bytes[5]) * 0x2000
	prgRom := make([]uint8, prgSize)
	prgRom = bytes[16 : 16+prgSize]
	var chrRom []uint8
	if chrSize != 0 {
		chrRom = make([]uint8, chrSize)
		chrRom = bytes[16+prgSize : 16+prgSize+chrSize]
	} else {
		chrRom = make([]uint8, 0x2000)
	}

	return bytes[:16], prgRom, chrRom
}

func NewNES(path string) *NES {
	n := new(NES)
	n.keys = [8]bool{}
	header, prg, chr := Load(path)
	isHorizontalMirror := header[6]&0x01 == 0x00
	n.ppu = ppu.NewPPU(chr, isHorizontalMirror)
	n.apu = apu.NewAPU(0)
	n.cpu = cpu.NewCPU(prg, n.ppu, n.apu)
	n.canvas = ebiten.NewImage(256, 240)
	n.isPlay = true
	pauseBG = ebiten.NewImage(256, 240)
	pauseBG.Fill(color.Black)
	pauseOP = &ebiten.DrawImageOptions{}
	pauseOP.ColorM.Scale(0, 0, 0, 0.5)
	return n
}

func (n *NES) SetDebug() {
	n.isDebug = true
}

//////////////////////
//ebiten Callbacks

//Draw はPPUから画面データを受け取り描画
func (n *NES) Draw(screen *ebiten.Image) {

	//Draw NES frame
	background, sprites := n.ppu.Draw()
	n.canvas.DrawImage(background, nil)
	n.canvas.DrawImage(sprites, nil)

	//Draw interface
	if !n.isPlay {
		n.canvas.DrawImage(pauseBG, pauseOP)
	} else {

	}
	if n.isDebug {
		if !n.isPlay {
			ebitenutil.DebugPrint(n.canvas, fmt.Sprintf("TPS:%0.2f, %s\nA:%X B:%t X:%X Y:%X", ebiten.CurrentTPS(), n.cpu.GetDebugText(),
				n.cpu.A, n.cpu.B, n.cpu.X, n.cpu.Y))
		} else {
			ebitenutil.DebugPrint(n.canvas, fmt.Sprintf("TPS:%0.2f, %s", ebiten.CurrentTPS(), n.cpu.GetDebugText()))
		}
	}
	if n.isRecording {
		ebitenutil.DebugPrintAt(n.canvas, "REC", 230, 0)
	}
	//x3
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(3, 3)
	screen.DrawImage(n.canvas, op)
	return
}

func (n *NES) Layout(screenWidth, screenHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (n *NES) Update() error {
	if !n.isGamepadEnabled {
		for _, id := range inpututil.JustConnectedGamepadIDs() {
			n.gamepadID = id
			n.isGamepadEnabled = true
		}

	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		n.isPlay = !n.isPlay
	}
	if n.isDebug && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if n.isRecording {
			ioutil.WriteFile("neslog.log", ([]byte)(n.cpu.DebugLog), 0666)
			n.cpu.DebugLog = ""
		}
		n.isRecording = !n.isRecording
		n.cpu.IsRecord = n.isRecording
		n.isPlay = false
	}

	if n.isPlay {
		//NES Emulation
		if n.isGamepadEnabled {
			for k := 0; k < 4; k++ {
				n.keys[k] = ebiten.IsGamepadButtonPressed(n.gamepadID, padmap[k])
			}
			x := ebiten.GamepadAxis(n.gamepadID, GAMEPAD_AXIS_X)
			y := ebiten.GamepadAxis(n.gamepadID, GAMEPAD_AXIS_Y)
			//WSAD
			n.keys[4] = y < -0.5
			n.keys[5] = y > 0.5
			n.keys[6] = x < -0.5
			n.keys[7] = x > 0.5

		} else {
			for k := 0; k < 8; k++ {
				n.keys[k] = ebiten.IsKeyPressed(keymap[k])
			}
		}
		isScreenReady := false
		for !isScreenReady {
			cycle := n.cpu.Run(n.keys)
			isScreenReady = n.ppu.Run(cycle * 3)
			n.apu.Run(cycle)
		}
	} else {

	}
	return nil
}

//////////////////

func (n *NES) init() {
	ebiten.SetWindowSize(256*3, 240*3)
}

func (n *NES) Run() {
	n.init()
	if err := ebiten.RunGame(n); err != nil {
		log.Fatal(err)
	}
}

func (n *NES) DrawPatternTable(sprites [][]uint8) *ebiten.Image {
	canvas := ebiten.NewImage(256, 240)
	canvas.Fill(color.Black)
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
		canvas.DrawImage(image, op)
	}
	return canvas
}
