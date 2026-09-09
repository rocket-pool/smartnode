package units

import (
	"fmt"
	"io"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	_ fmt.Formatter = Eth{}
	_ fmt.Formatter = Gwei{}
	_ fmt.Formatter = MilliEth{}
	_ fmt.Formatter = Wei{}
)

func (e Eth) Format(f fmt.State, verb rune) {
	formatDecimal(f, verb, e.Decimal, "units.Eth")
}

func (g Gwei) Format(f fmt.State, verb rune) {
	formatDecimal(f, verb, g.Decimal, "units.Gwei")
}

func (m MilliEth) Format(f fmt.State, verb rune) {
	formatDecimal(f, verb, m.Decimal, "units.MilliEth")
}

func (w Wei) Format(f fmt.State, verb rune) {
	switch verb {
	case 'f', 'F':
		formatDecimal(f, verb, w.Decimal, "units.Wei")
	case 'd', 's', 'v':
		if verb == 'v' && f.Flag('#') {
			fmt.Fprintf(f, "units.Wei(%s)", w.String())
			return
		}
		writeFormatted(f, w.String())
	default:
		fmt.Fprintf(f, "%%!%c(units.Wei=%s)", verb, w.String())
	}
}

func formatDecimal(f fmt.State, verb rune, d decimal.Decimal, typ string) {
	switch verb {
	case 'f', 'F':
		prec, ok := f.Precision()
		if !ok {
			prec = 6
		}
		writeFormatted(f, d.StringFixed(int32(prec)))
	case 'e', 'E', 'g', 'G':
		// Scientific / compact forms go through float64; prefer %f for exact display.
		prec, ok := f.Precision()
		if !ok {
			if verb == 'e' || verb == 'E' {
				prec = 6
			} else {
				prec = -1
			}
		}
		fmt.Fprintf(f, fmtVerb(f, verb), prec, d.InexactFloat64())
	case 's', 'v':
		if verb == 'v' && f.Flag('#') {
			fmt.Fprintf(f, "%s(%s)", typ, d.String())
			return
		}
		writeFormatted(f, d.String())
	default:
		fmt.Fprintf(f, "%%!%c(%s=%s)", verb, typ, d.String())
	}
}

// fmtVerb rebuilds a width/precision/flags format for float64 fallback verbs.
func fmtVerb(f fmt.State, verb rune) string {
	var b strings.Builder
	b.WriteByte('%')
	for _, flag := range "+-# 0" {
		if f.Flag(int(flag)) {
			b.WriteRune(flag)
		}
	}
	if width, ok := f.Width(); ok {
		fmt.Fprintf(&b, "%d", width)
	}
	b.WriteString(".*")
	b.WriteRune(verb)
	return b.String()
}

func writeFormatted(f fmt.State, s string) {
	width, hasWidth := f.Width()
	if !hasWidth || width <= len(s) {
		_, _ = io.WriteString(f, s)
		return
	}
	pad := width - len(s)
	padding := strings.Repeat(" ", pad)
	if f.Flag('0') && !f.Flag('-') {
		padding = strings.Repeat("0", pad)
		if strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
			_, _ = io.WriteString(f, s[:1])
			_, _ = io.WriteString(f, strings.Repeat("0", pad))
			_, _ = io.WriteString(f, s[1:])
			return
		}
	}
	if f.Flag('-') {
		_, _ = io.WriteString(f, s)
		_, _ = io.WriteString(f, padding)
		return
	}
	_, _ = io.WriteString(f, padding)
	_, _ = io.WriteString(f, s)
}
