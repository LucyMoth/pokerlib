package main

import (
	"fmt"
	"strings"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

func PrintHeader(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s%s%s\n", Bold, Cyan, msg, Reset)
}

func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s✓ %s%s\n", Bold, Green, msg, Reset)
}

func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s✗ %s%s\n", Bold, Red, msg, Reset)
}

func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s⚠ %s%s\n", Bold, Yellow, msg, Reset)
}

func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%sℹ %s%s\n", Dim, White, msg, Reset)
}

func PrintPrompt(prompt string) {
	fmt.Printf("%s%s%s%s", Bold, BrightCyan, prompt, Reset)
}

func PrintLabel(label string, value interface{}) {
	fmt.Printf("  %s%s%-15s%s %v\n", Dim, White, label+":", Reset, value)
}

func PrintDivider() {
	fmt.Printf("%s%s%s%s\n", Dim, BrightBlack, strings.Repeat("─", 50), Reset)
}

func PrintTableHeader(columns []string, widths []int) {
	fmt.Print(Bold + Cyan)
	for i, col := range columns {
		fmt.Printf("%-*s ", widths[i], col)
	}
	fmt.Println(Reset)

	fmt.Print(Dim + BrightBlack)
	for _, w := range widths {
		fmt.Print(strings.Repeat("─", w) + " ")
	}
	fmt.Println(Reset)
}

func PrintTableRow(values []string, widths []int, highlight bool) {
	if highlight {
		fmt.Print(Bold + BrightGreen)
	} else {
		fmt.Print(White)
	}
	for i, val := range values {
		fmt.Printf("%-*s ", widths[i], val)
	}
	fmt.Println(Reset)
}

func PrintTableRowColored(values []string, widths []int, colors []string) {
	for i, val := range values {
		color := White
		if i < len(colors) && colors[i] != "" {
			color = colors[i]
		}
		fmt.Printf("%s%-*s%s ", color, widths[i], val, Reset)
	}
	fmt.Println()
}

func ColorizeNumber(n int, positive bool) string {
	if positive {
		if n > 0 {
			return fmt.Sprintf("%s%s+%d%s", Bold, Green, n, Reset)
		} else if n < 0 {
			return fmt.Sprintf("%s%s%d%s", Bold, Red, n, Reset)
		}
	}
	return fmt.Sprintf("%d", n)
}

func ColorizePercentage(pct float64) string {
	if pct >= 60 {
		return fmt.Sprintf("%s%s%.1f%%%s", Bold, Green, pct, Reset)
	} else if pct >= 40 {
		return fmt.Sprintf("%s%.1f%%%s", Yellow, pct, Reset)
	} else {
		return fmt.Sprintf("%s%.1f%%%s", Red, pct, Reset)
	}
}

func ColorizeChips(chips int, initial int) string {
	diff := chips - initial
	if diff > 0 {
		return fmt.Sprintf("%s%d %s(+%d)%s", White, chips, Green, diff, Reset)
	} else if diff < 0 {
		return fmt.Sprintf("%s%d %s(%d)%s", White, chips, Red, diff, Reset)
	}
	return fmt.Sprintf("%d", chips)
}

func ColorizeStrategy(name string) string {
	switch name {
	case "GTO":
		return fmt.Sprintf("%s%s%s%s", Bold, Cyan, name, Reset)
	case "TAG":
		return fmt.Sprintf("%s%s%s%s", Bold, Blue, name, Reset)
	case "LAG":
		return fmt.Sprintf("%s%s%s%s", Bold, Magenta, name, Reset)
	case "Fish":
		return fmt.Sprintf("%s%s%s%s", Bold, Yellow, name, Reset)
	case "Nit":
		return fmt.Sprintf("%s%s%s%s", Bold, BrightBlack, name, Reset)
	default:
		return name
	}
}

func ColorizeHandRank(rank string) string {
	switch rank {
	case "Royal Flush", "Straight Flush":
		return fmt.Sprintf("%s%s%s%s", Bold, BrightMagenta, rank, Reset)
	case "Four of a Kind", "Full House":
		return fmt.Sprintf("%s%s%s%s", Bold, BrightYellow, rank, Reset)
	case "Flush", "Straight":
		return fmt.Sprintf("%s%s%s%s", Bold, BrightCyan, rank, Reset)
	case "Three of a Kind", "Two Pair":
		return fmt.Sprintf("%s%s%s%s", Bold, BrightGreen, rank, Reset)
	case "One Pair":
		return fmt.Sprintf("%s%s%s%s", Dim, White, rank, Reset)
	default:
		return fmt.Sprintf("%s%s%s", Dim, rank, Reset)
	}
}

func ColorizeCard(card string) string {
	if len(card) < 2 {
		return card
	}
	suit := card[len(card)-1]
	switch suit {
	case 'h', 'H':
		return fmt.Sprintf("%s%s%s", Red, card, Reset)
	case 'd', 'D':
		return fmt.Sprintf("%s%s%s", BrightRed, card, Reset)
	case 'c', 'C':
		return fmt.Sprintf("%s%s%s", BrightGreen, card, Reset)
	case 's', 'S':
		return fmt.Sprintf("%s%s%s", BrightWhite, card, Reset)
	default:
		return card
	}
}

func PrintBox(title string, lines []string) {
	maxLen := len(title)
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}
	width := maxLen + 4

	fmt.Printf("%s%s╭%s╮%s\n", Bold, Cyan, strings.Repeat("─", width), Reset)
	fmt.Printf("%s%s│ %-*s │%s\n", Bold, Cyan, width-2, title, Reset)
	fmt.Printf("%s%s├%s┤%s\n", Dim, Cyan, strings.Repeat("─", width), Reset)

	for _, line := range lines {
		fmt.Printf("%s│%s %-*s %s│%s\n", Dim+Cyan, Reset, width-2, line, Dim+Cyan, Reset)
	}

	fmt.Printf("%s%s╰%s╯%s\n", Dim, Cyan, strings.Repeat("─", width), Reset)
}

func PrintProgressBar(current, total int, width int) {
	pct := float64(current) / float64(total)
	filled := int(pct * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Printf("\r%s%s[%s]%s %s%.1f%%%s (%d/%d)",
		Bold, Cyan, bar, Reset,
		BrightWhite, pct*100, Reset,
		current, total)

	if current == total {
		fmt.Println()
	}
}
