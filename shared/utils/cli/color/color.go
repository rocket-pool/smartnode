package color

import (
	"fmt"
	"strings"
)

const colorReset string = "\033[0m"
const colorBold string = "\033[1m"
const colorRed string = "\033[31m"
const colorGreen string = "\033[32m"
const colorYellow string = "\033[33m"
const colorLightBlue string = "\033[36m"

func color(color, msg string) string {
	return fmt.Sprintf("%s%s%s", color, msg, colorReset)
}

func Red(msg string) string {
	return color(colorRed, msg)
}

func RedPrintln(msgs ...string) {
	fmt.Println(Red(strings.Join(msgs, " ")))
}

func RedPrintf(format string, a ...any) {
	fmt.Printf(Red(format), a...)
}

func RedSprintf(format string, a ...any) string {
	return color(colorRed, fmt.Sprintf(format, a...))
}

func Green(msg string) string {
	return color(colorGreen, msg)
}

func GreenPrintln(msgs ...string) {
	fmt.Println(Green(strings.Join(msgs, " ")))
}

func GreenPrintf(format string, a ...any) {
	fmt.Printf(Green(format), a...)
}

func GreenSprintf(format string, a ...any) string {
	return color(colorGreen, fmt.Sprintf(format, a...))
}

func Yellow(msg string) string {
	return color(colorYellow, msg)
}

func YellowPrintln(msgs ...string) {
	fmt.Println(Yellow(strings.Join(msgs, " ")))
}

func YellowPrintf(format string, a ...any) {
	fmt.Printf(Yellow(format), a...)
}

func YellowSprintf(format string, a ...any) string {
	return color(colorYellow, fmt.Sprintf(format, a...))
}

func LightBlue(msg string) string {
	return color(colorLightBlue, msg)
}

func LightBluePrintln(msgs ...string) {
	fmt.Println(LightBlue(strings.Join(msgs, " ")))
}

func LightBluePrintf(format string, a ...any) {
	fmt.Printf(LightBlue(format), a...)
}

func LightBlueSprintf(format string, a ...any) string {
	return color(colorLightBlue, fmt.Sprintf(format, a...))
}

func Bold(msg string) string {
	return color(colorBold, msg)
}

func BoldPrintln(msgs ...string) {
	fmt.Println(Bold(strings.Join(msgs, " ")))
}

func BoldPrintf(format string, a ...any) {
	fmt.Printf(Bold(format), a...)
}

func BoldSprintf(format string, a ...any) string {
	return color(colorBold, fmt.Sprintf(format, a...))
}
