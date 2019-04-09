// credit to https://github.com/fogleman/nes
package ui

import (
	"image"

	// "strconv"

	"github.com/giongto35/cloud-game/nes"
)

const padding = 0

// List key pressed
const (
	a1 = iota
	b1
	select1
	start1
	up1
	down1
	left1
	right1
	save1
	load1
	a2
	b2
	select2
	start2
	up2
	down2
	left2
	right2
	save2
	load2
)
const NumKeys = 10

type GameView struct {
	director *Director
	console  *nes.Console
	title    string
	hash     string

	// equivalent to the list key pressed const above
	keyPressed [NumKeys * 2]bool

	imageChannel chan *image.RGBA
	inputChannel chan int
}

func NewGameView(director *Director, console *nes.Console, title, hash string, imageChannel chan *image.RGBA, inputChannel chan int) View {
	gameview := &GameView{
		director:     director,
		console:      console,
		title:        title,
		hash:         hash,
		keyPressed:   [NumKeys * 2]bool{false},
		imageChannel: imageChannel,
		inputChannel: inputChannel,
	}

	go gameview.ListenToInputChannel()
	return gameview
}

// ListenToInputChannel listen from input channel streamm, which is exposed to WebRTC session
func (view *GameView) ListenToInputChannel() {
	for {
		keysInBinary := <-view.inputChannel
		for i := 0; i < NumKeys*2; i++ {
			view.keyPressed[i] = ((keysInBinary & 1) == 1)
			keysInBinary = keysInBinary >> 1
		}
	}
}

// Enter enter the game view.
func (view *GameView) Enter() {
	// Always reset game
	// Legacy Audio code. TODO: Add it back to support audio
	// view.console.SetAudioChannel(view.director.audio.channel)
	// view.console.SetAudioSampleRate(view.director.audio.sampleRate)

	// load state
	view.console.Reset()

	// load sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash)); err == nil {
			cartridge.SRAM = sram
		}
	}
}

// Exit ...
func (view *GameView) Exit() {
	// view.console.SetAudioChannel(nil)
	// view.console.SetAudioSampleRate(0)
	// save sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		writeSRAM(sramPath(view.hash), cartridge.SRAM)
	}
}

// Update is called for every period of time, dt is the elapsed time from the last frame
func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	console := view.console
	view.updateControllers()
	console.StepSeconds(dt)

	// fps to set frame
	view.imageChannel <- console.Buffer()
}

func (view *GameView) updateControllers() {
	// Divide keyPressed to player 1 and player 2
	// First 10 keys are player 1
	var player1Keys [8]bool
	copy(player1Keys[:], view.keyPressed[:8])
	var player2Keys [8]bool
	copy(player2Keys[:], view.keyPressed[10:18])

	view.console.Controller1.SetButtons(player1Keys)
	view.console.Controller2.SetButtons(player2Keys)

	if view.keyPressed[save1] || view.keyPressed[save2] {
		view.console.SaveState(savePath(view.hash))
	}
	if view.keyPressed[load1] || view.keyPressed[load2] {
		view.console.LoadState(savePath(view.hash))
	}
}
