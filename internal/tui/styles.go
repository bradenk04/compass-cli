package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ColorGreen    = "#00ED64"
	ColorSlateBg  = "#1C1D1F"
	ColorSlateSel = "#323338"
	ColorBorder   = "#43464E"
	ColorText     = "#E3E4E6"
	ColorMuted    = "#888B94"
	ColorAmber    = "#F59E0B"
	ColorRed      = "#EF4444"
	ColorTeal     = "#0D9488"
	ColorPurple   = "#A855F7"
	ColorBlue     = "#3B82F6"
	ColorIndigo   = "#4F46E5"
)

var (
	MainStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorSlateBg))

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSlateBg)).
			Background(lipgloss.Color(ColorGreen)).
			Bold(true).
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Bold(true)

	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1)

	ContentStyle = lipgloss.NewStyle().
			Padding(0, 1)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color(ColorGreen)).
			Padding(0, 2).
			Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorMuted)).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Padding(0, 2)

	TabGapStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color(ColorBorder))

	DbStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Bold(true)

	CollStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			PaddingLeft(2)

	CollSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSlateBg)).
				Background(lipgloss.Color(ColorGreen)).
				PaddingLeft(2).
				Bold(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(1)

	CardActiveStyle = CardStyle.Copy().
			BorderForeground(lipgloss.Color(ColorGreen))

	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed))
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAmber))

	KeyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple)).Bold(true)
	ValStrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal))
	ValNumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAmber))
	ValBoolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue))
	SymbolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted))
)

func HighlightJSON(jsonStr string) string {
	keyRegex     := regexp.MustCompile(`"([^"]+)"\s*:`)
	strValRegex  := regexp.MustCompile(`:\s*"([^"]*)"`)
	boolValRegex := regexp.MustCompile(`:\s*(true|false|null)\b`)
	numValRegex  := regexp.MustCompile(`:\s*(-?\d+(\.\d+)?([eE][+-]?\d+)?)\b`)
	extKeyRegex  := regexp.MustCompile(`"\$([a-zA-Z0-9]+)"`)

	lines := strings.Split(jsonStr, "\n")
	highlighted := make([]string, len(lines))

	for i, line := range lines {
		line = keyRegex.ReplaceAllStringFunc(line, func(match string) string {
			key := keyRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s%s", KeyStyle.Render(fmt.Sprintf(`"%s"`, key)), SymbolStyle.Render(":"))
		})

		line = strValRegex.ReplaceAllStringFunc(line, func(match string) string {
			val := strValRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s %s", SymbolStyle.Render(":"), ValStrStyle.Render(fmt.Sprintf(`"%s"`, val)))
		})

		line = boolValRegex.ReplaceAllStringFunc(line, func(match string) string {
			val := boolValRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s %s", SymbolStyle.Render(":"), ValBoolStyle.Render(val))
		})

		line = numValRegex.ReplaceAllStringFunc(line, func(match string) string {
			val := numValRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s %s", SymbolStyle.Render(":"), ValNumStyle.Render(val))
		})

		line = extKeyRegex.ReplaceAllStringFunc(line, func(match string) string {
			return InfoStyle.Render(match)
		})

		highlighted[i] = line
	}

	return strings.Join(highlighted, "\n")
}
