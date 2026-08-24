//go:build nogui

package ui

import (
	"fmt"
	"os/exec"
)

func OpenGUI(engName string) {
	fmt.Printf("GUI config tool for %s is disabled in nogui build mode.\n", engName)
	_ = exec.Command("ibus-setup").Start()
}
