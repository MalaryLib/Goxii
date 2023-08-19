package main

import (
	"github.com/fatih/color"
)

func SubtleTextIndent(str string, indent bool) {
	if indent {
		color.New(color.Faint).Printf("\t%s", str)
	} else {
		color.New(color.Faint).Printf("%s", str)
	}
}

func SubtleText(str string) {
	color.New(color.Faint).Printf("\t%s", str)
}

func H1Print(str string) {
	PrintColor(str, color.FgHiYellow)
}

func WarningText(str string) {
	PrintColor(str, color.FgHiRed)
}

func PrintColor(str string, c color.Attribute) {
	color.New(c).Printf("%s", str)
}
