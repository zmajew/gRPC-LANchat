package internal

import (
	"bufio"
	"strings"

	"github.com/dbatbold/beep"
)

func (node *Node) BeepMessage() {
	music := beep.NewMusic("") // output can be file as "music.wav"
	volume := node.Volume
	
	beep.OpenSoundDevice("default")

	beep.InitSoundDevice()

	beep.PrintSheet = false
	defer beep.CloseSoundDevice()

	musicScore := `
        VP SA8 SR9
        A9HRDE cc 
    `

	reader := bufio.NewReader(strings.NewReader(musicScore))
	go music.Play(reader, *volume)
	music.Wait()
	beep.FlushSoundBuffer()
}
