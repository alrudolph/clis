package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

var (
	packageJson Package
	selected    int
	offsetIdx   int
	command     string
)

const paddingTop = 2
const paddingBottom = 2

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		log.Panicln(err)
	}
	// defer g.Close()

	// load commands once
	packageJson, err = loadPackageJsonCommands("package.json")

	if err != nil {
		log.Panicln(err)
	}

	g.SetManagerFunc(layout)

	for _, key := range []gocui.Key{gocui.KeyArrowDown, 's', 'j'} {
		if err = g.SetKeybinding("", key, gocui.ModNone, cursorDown); err != nil {
			log.Panicln(err)
		}
	}

	for _, key := range []gocui.Key{gocui.KeyArrowUp, 'w', 'k'} {
		if err = g.SetKeybinding("", key, gocui.ModNone, cursorUp); err != nil {
			log.Panicln(err)
		}
	}

	if err = g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, runCommand); err != nil {
		log.Panicln(err)
	}
	if err = g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, closeOutput); err != nil {
		log.Panicln(err)
	}

	if err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	for i := range 10 {
		if err := g.SetKeybinding("", rune('0'+i), gocui.ModNone, runCommandShortcut(i)); err != nil {
			log.Panicln(err)
		}
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}

	g.Close()

	if command == "" {
		return
	}

	cmd := exec.Command("npm", "run", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("%s failed with %s\n", command, err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	v, err := g.SetView("commands", -1, -1, maxX, maxY)

	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
	}

	v.Frame = false
	err = v.SetOrigin(0, 0)

	if err != nil {
		return err
	}

	v.Clear()

	if err = refreshCommandsView(g); err != nil {
		return err
	}

	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func refreshCommandsView(g *gocui.Gui) error {
	v, err := g.View("commands")

	if err != nil {
		return err
	}

	v.Clear()

	maxX, _ := g.Size()
	position := 0

	var header string

	if offsetIdx == 0 {
		header = fmt.Sprintf(" (%d commands)", packageJson.nCommands())
	} else {
		header = fmt.Sprintf(" (%d above)", offsetIdx)
	}

	title := head("["+packageJson.title+"]", maxX-len(header)-1)
	t := color.New(color.Bold)
	_, err = t.Fprint(v, title)

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(v, "%s%s\n", strings.Repeat(" ", maxX-len(title)-len(header)), header)

	if err != nil {
		return err
	}

	for i := 1; i < paddingTop; i++ {
		_, err = fmt.Fprintln(v, "")

		if err != nil {
			return err
		}
	}

	nShortcutLines := 10

	for j := range min(getHeight(g), packageJson.nCommands()) {
		i := j + offsetIdx

		if i >= packageJson.nCommands() || i < 0 {
			break
		}

		cmd := packageJson.commands[i]

		var line bytes.Buffer

		selectedCommand := " [Enter]"
		cursor := "> "
		emptyCursor := strings.Repeat(" ", len(cursor))

		if i == selected {
			a := color.New(color.FgYellow, color.Bold)
			_, err = a.Fprint(&line, cursor)

			if err != nil {
				return err
			}

			remainingColumns := maxX - len(selectedCommand) - len(cursor)

			c := color.New(color.FgCyan, color.Bold, color.Underline)
			_, err = c.Fprint(&line, head(cmd.Name, remainingColumns))
			remainingColumns -= len(head(cmd.Name, remainingColumns))

			if err != nil {
				return err
			}

			d := color.New(color.FgWhite)
			cmd := ": " + cmd.Command
			_, err = d.Fprint(&line, head(cmd, remainingColumns))
			remainingColumns -= len(head(cmd, remainingColumns))

			if err != nil {
				return err
			}

			_, err = d.Fprint(&line, strings.Repeat(" ", max(0, remainingColumns)))

			if err != nil {
				return err
			}

			e := color.New(color.FgGreen, color.Bold)
			_, err = e.Fprint(&line, selectedCommand)

			if err != nil {
				return err
			}
		} else {
			remainingColumns := maxX - len(emptyCursor)

			if position < nShortcutLines {
				remainingColumns -= 4
			}

			c := color.New(color.FgCyan, color.Bold)
			_, err = c.Fprintf(&line, "%s%s", emptyCursor, head(cmd.Name, remainingColumns))
			remainingColumns -= len(head(cmd.Name, remainingColumns))

			if err != nil {
				return err
			}

			d := color.New(color.FgBlack)
			cmd := ": " + cmd.Command
			_, err = d.Fprint(&line, head(cmd, remainingColumns))
			remainingColumns -= len(head(cmd, remainingColumns))

			if err != nil {
				return err
			}

			_, err = d.Fprint(&line, strings.Repeat(" ", max(0, remainingColumns)))

			if err != nil {
				return err
			}

			if position < nShortcutLines {
				e := color.New(color.FgBlack)
				_, err = e.Fprintf(&line, " (%d)", position)

				if err != nil {
					return err
				}
			}

			position++
		}

		_, err = fmt.Fprintln(v, line.String())

		if err != nil {
			return err
		}
	}

	for i := 1; i < paddingBottom; i++ {
		_, err = fmt.Fprintln(v, "")

		if err != nil {
			return err
		}
	}

	footer := ""

	if offsetIdx+getHeight(g) < packageJson.nCommands() {
		footer = fmt.Sprintf(" (%d below)", packageJson.nCommands()-(offsetIdx+getHeight(g)))
	}

	c := color.New(color.FgBlack)
	commands := head("CTRL+C to quit", maxX-len(footer))
	_, err = c.Fprint(v, commands)

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(v, "%s%s", strings.Repeat(" ", maxX-len(footer)-len(commands)), footer)

	if err != nil {
		return err
	}

	return nil
}

func head(value string, n int) string {
	if len(value) <= n {
		return value
	}

	return value[:n]
}

func cursorDown(g *gocui.Gui, _ *gocui.View) error {
	if packageJson.nCommands() == 0 {
		return nil
	}

	selected = min(selected+1, packageJson.nCommands()-1)

	height := getHeight(g)

	if selected >= offsetIdx+height {
		offsetIdx = selected - height + 1
	}

	return refreshCommandsView(g)
}

func getHeight(g *gocui.Gui) int {
	_, maxY := g.Size()

	return maxY - paddingTop - paddingBottom
}

func cursorUp(g *gocui.Gui, _ *gocui.View) error {
	if packageJson.nCommands() == 0 {
		return nil
	}

	selected = max(selected-1, 0)

	if selected < offsetIdx {
		offsetIdx = selected
	}

	return refreshCommandsView(g)
}

func runCommand(g *gocui.Gui, v *gocui.View) error {
	if packageJson.nCommands() == 0 {
		return nil
	}

	cmd := packageJson.commands[selected]
	command = cmd.Name

	return quit(g, v)
}

func runCommandShortcut(shortcut int) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		var idx int

		if shortcut < selected-offsetIdx {
			idx = offsetIdx + shortcut
		} else {
			idx = offsetIdx + shortcut + 1
		}

		if idx == -1 || idx >= packageJson.nCommands() {
			return nil
		}

		command = packageJson.commands[idx].Name

		return quit(g, v)
	}
}

func closeOutput(g *gocui.Gui, _ *gocui.View) error {
	return g.DeleteView("output")
}
