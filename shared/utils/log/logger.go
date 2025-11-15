package log

import (
	"log"

	"github.com/fatih/color"
)

// Logger with ANSI color output
type ColorLogger struct {
	Color       color.Attribute
	sprintFunc  func(a ...any) string
	sprintfFunc func(format string, a ...any) string
}

// Create new color logger
func NewColorLogger(colorAttr color.Attribute) ColorLogger {
	return ColorLogger{
		Color:       colorAttr,
		sprintFunc:  color.New(colorAttr).SprintFunc(),
		sprintfFunc: color.New(colorAttr).SprintfFunc(),
	}
}

// Print values
func (l *ColorLogger) Print(v ...any) {
	log.Print(l.sprintFunc(v...))
}

// Print values with a newline
func (l *ColorLogger) Println(v ...any) {
	log.Println(l.sprintFunc(v...))
}

// Print a formatted string
func (l *ColorLogger) Printf(format string, v ...any) {
	log.Print(l.sprintfFunc(format, v...))
}

// Print a formatted string with a newline
func (l *ColorLogger) Printlnf(format string, v ...any) {
	log.Println(l.sprintfFunc(format, v...))
}
