package main

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

const (
	ViewChat   = "chat"
	ViewInput  = "input"
	ViewStatus = "status"
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// chat view (top 80%)
	if v, err := g.SetView(ViewChat, 0, 0, maxX-1, int(float64(maxY)*.8)-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Chat History"
		v.Wrap = true
		v.Autoscroll = true
	}

	// input view (remaining space)
	if v, err := g.SetView(ViewInput, 0, int(float64(maxY)*.8), maxX-1, maxY-3, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Your Message (Press Enter to send, Ctrl+C to quit) "
		v.Wrap = true
		v.Editable = true
		if _, err := g.SetCurrentView(ViewInput); err != nil {
			return nil
		}
	}

	// status view
	if v, err := g.SetView(ViewStatus, 0, maxY-2, maxX-1, maxY-1, 0); err != nil {
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

	// Send message on Enter
	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, sendMessage); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func sendMessage(g *gocui.Gui, v *gocui.View) error {
	message := strings.TrimSpace(v.Buffer())
	if message == "" {
		return nil
	}

	v.Clear()
	v.SetCursor(0, 0)

	addMessageToChat("user", message)

	updateStatus("Processing...")

	// go routines!
	go func() {
		reply, err := processChat(message, history)
		if err != nil {
			g.Update(func(g *gocui.Gui) error {
				addMessageToChat("error", fmt.Sprintf("Error: %v", err))
				updateStatus("Ready")
				return nil
			})
			return
		}

		history = append(history, ChatMessage{Role: "user", Content: message})
		history = append(history, ChatMessage{Role: "assistant", Content: reply})

		g.Update(func(g *gocui.Gui) error {
			addMessageToChat("assistant", reply)
			updateStatus("Ready")
			return nil
		})
	}()

	return nil
}
