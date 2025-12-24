package main

import (
	"bytes"
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

const paddingTop = 1
const paddingBottom = 1

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

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 's', gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'j', gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'w', gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'k', gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, runCommand); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, closeOutput); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	g.Close()

	cmd := exec.Command("npm", "run", command)

	// Assign the command's stdout and stderr to the Go process's stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command and wait for it to complete
	if err := cmd.Run(); err != nil {
		log.Fatalf("%s failed with %s\n", command, err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	v, err := g.SetView("commands", -1, -1, maxX, maxY)
	v.Frame = false
	v.SetOrigin(0, 0)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	v.Clear()
	_ = refreshCommandsView(g)

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func refreshCommandsView(g *gocui.Gui) error {
	v, err := g.View("commands")
	if err != nil {
		return nil
	}

	v.Clear()

	maxX, _ := g.Size()
	position := 0

	header := ""

	if offsetIdx == 0 {
		header = fmt.Sprintf(" (%d commands)", packageJson.nCommands())
	} else {
		header = fmt.Sprintf(" (%d above)", offsetIdx)
	}

	title := head("["+packageJson.title+"]", maxX-len(header)-1)
	t := color.New(color.Bold)
	t.Fprint(v, title)
	fmt.Fprintf(v, "%s%s\n", strings.Repeat(" ", maxX-len(title)-len(header)), header)

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
			a.Fprint(&line, cursor)

			remainingColumns := maxX - len(selectedCommand) - len(cursor)

			c := color.New(color.FgCyan, color.Bold, color.Underline)
			c.Fprint(&line, head(cmd.Name, remainingColumns))
			remainingColumns -= len(head(cmd.Name, remainingColumns))

			d := color.New(color.FgWhite)
			cmd := fmt.Sprintf(": %s", cmd.Command)
			d.Fprint(&line, head(cmd, remainingColumns))
			remainingColumns -= len(head(cmd, remainingColumns))

			d.Fprint(&line, strings.Repeat(" ", max(0, remainingColumns)))

			e := color.New(color.FgGreen, color.Bold)
			e.Fprint(&line, selectedCommand)
		} else {
			remainingColumns := maxX - len(emptyCursor)

			if position < 10 {
				remainingColumns -= 4
			}

			c := color.New(color.FgCyan, color.Bold)
			c.Fprintf(&line, "%s%s", emptyCursor, head(cmd.Name, remainingColumns))
			remainingColumns -= len(head(cmd.Name, remainingColumns))

			d := color.New(color.FgBlack)
			cmd := fmt.Sprintf(": %s", cmd.Command)
			d.Fprint(&line, head(cmd, remainingColumns))
			remainingColumns -= len(head(cmd, remainingColumns))

			d.Fprint(&line, strings.Repeat(" ", max(0, remainingColumns)))

			if position < 10 {
				e := color.New(color.FgBlack)
				e.Fprintf(&line, " (%d)", position)
			}

			position++
		}

		fmt.Fprintln(v, line.String())
	}

	footer := ""

	if offsetIdx+getHeight(g) < packageJson.nCommands() {
		footer = fmt.Sprintf(" (%d below)", packageJson.nCommands()-(offsetIdx+getHeight(g)))
	}

	c := color.New(color.FgBlack)
	commands := head("CTRL+C to quit", maxX-len(footer))
	c.Fprint(v, commands)

	fmt.Fprintf(v, "%s%s", strings.Repeat(" ", maxX-len(footer)-len(commands)), footer)

	return nil
}

func head(value string, n int) string {
	if len(value) <= n {
		return value
	}
	return value[:n]
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
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

func cursorUp(g *gocui.Gui, v *gocui.View) error {
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

func closeOutput(g *gocui.Gui, v *gocui.View) error {
	_ = g.DeleteView("output")
	return nil
}
