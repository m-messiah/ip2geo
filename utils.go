package main

import (
	"github.com/fatih/color"
	"os"
	"path"
	"strings"
	"unicode"
)

func remove_space(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func open_map_file(output_dir, filename string) *os.File {
	filepath := path.Join(output_dir, filename)
	f, err := os.Create(filepath)
	if err != nil {
		color.Red("[FAIL]\nCan't open %s: %s", filepath, err.Error())
		return nil
	}
	return f
}
