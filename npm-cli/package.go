package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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

	dec := json.NewDecoder(bytes.NewReader(b))

	// root must be object
	t, err := dec.Token()
	if err != nil {
		return Package{}, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return Package{}, fmt.Errorf("expected JSON object")
	}

	title := ""

	var out []Command
	for dec.More() {
		// read field name
		keyTok, err := dec.Token()
		if err != nil {
			return Package{}, err
		}
		key := keyTok.(string)

		if key == "name" {
			if err := dec.Decode(&title); err != nil {
				return Package{}, err
			}
			continue
		}

		if key == "scripts" {
			// scripts value must be an object
			t, err := dec.Token()
			if err != nil {
				return Package{}, err
			}
			if delim, ok := t.(json.Delim); !ok || delim != '{' {
				return Package{}, fmt.Errorf("scripts is not an object")
			}

			// iterate scripts object in order
			for dec.More() {
				kTok, err := dec.Token()
				if err != nil {
					return Package{}, err
				}
				name := kTok.(string)
				var cmd string
				if err := dec.Decode(&cmd); err != nil {
					return Package{}, err
				}
				out = append(out, Command{Name: name, Command: cmd})
			}
			// consume closing '}' of scripts
			if _, err := dec.Token(); err != nil {
				return Package{}, err
			}
		} else {
			// skip other values
			var skip interface{}
			if err := dec.Decode(&skip); err != nil {
				return Package{}, err
			}
		}
	}

	return Package{title: title, commands: out}, nil
}
