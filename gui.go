package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// chat view (top 80%)
	if v, err := g.SetView("chat", 0, 0, maxX-1, int(float64(maxY)*.8)-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Chat History"
		v.Wrap = true
		v.Autoscroll = true
	}

	// input view (remaining space)
	if v, err := g.SetView("input", 0, int(float64(maxY)*.8), maxX-1, maxY-3, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Your Message (Press Enter to send, Ctrl+C to quit) "
		v.Wrap = true
		v.Editable = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return nil
		}
	}

	// status view
	if v, err := g.SetView("status", 0, maxY-2, maxX-1, maxY-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprint(v, "Ready | Enter: Send | Ctrl+C: Quit")
	}

	return nil
}

func keybindings(g *gocui.Gui) error {
	// Quit on Ctrl+C
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
