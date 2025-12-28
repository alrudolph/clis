package main

import (
	"os"

	"github.com/go-faster/jx"
)

type Command struct {
	Name    string
	Command string
}

type Package struct {
	title    string
	commands []Command
}

func (p Package) nCommands() int {
	return len(p.commands)
}

func loadPackageJsonCommands(path string) (Package, error) {
	b, err := os.ReadFile(path)

	if err != nil {
		return Package{}, err
	}

	var title string

	var commands []Command

	d := jx.DecodeBytes(b)

	err = d.Obj(func(d *jx.Decoder, key string) error {
		switch key {
		case "name":
			if title, err = d.Str(); err != nil {
				return err
			}
		case "scripts":
			return d.Obj(func(d *jx.Decoder, key string) error {
				command, err := d.Str()

				if err != nil {
					return err
				}

				commands = append(commands, Command{
					Name:    key,
					Command: command,
				})

				return nil
			})
		default:
			return d.Skip()
		}

		return nil
	})

	if err != nil {
		return Package{}, err
	}

	return Package{title: title, commands: commands}, nil
}
