package tui

import (
	"fmt"
	"regexp"
	"strings"

	"bradenkennedy.com/compass-cli/internal/config"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	BorderColor       lipgloss.TerminalColor // i got lazy, this should be replaced with styles when its used instead of putting the styles in the files
	MainStyle         lipgloss.Style
	HeaderStyle       lipgloss.Style
	TitleStyle        lipgloss.Style
	SubtitleStyle     lipgloss.Style
	SidebarStyle      lipgloss.Style
	ContentStyle      lipgloss.Style
	ActiveTabStyle    lipgloss.Style
	InactiveTabStyle  lipgloss.Style
	TabGapStyle       lipgloss.Style
	DbStyle           lipgloss.Style
	CollStyle         lipgloss.Style
	CollSelectedStyle lipgloss.Style
	CardStyle         lipgloss.Style
	CardActiveStyle   lipgloss.Style
	SuccessStyle      lipgloss.Style
	ErrorStyle        lipgloss.Style
	InfoStyle         lipgloss.Style
	WarningStyle      lipgloss.Style
	KeyStyle          lipgloss.Style
	ValStrStyle       lipgloss.Style
	ValNumStyle       lipgloss.Style
	ValBoolStyle      lipgloss.Style
	SymbolStyle       lipgloss.Style
	FooterStyle       lipgloss.Style
	LabelStyle        lipgloss.Style
	InputLabelStyle   lipgloss.Style
	TextMutedStyle    lipgloss.Style
}

func InitStyles(theme config.ThemeConfig) Styles {
	s := Styles{
		BorderColor: lipgloss.Color(theme.BorderColor),
	}

	s.MainStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor)).
		Background(lipgloss.Color(theme.BackgroundColor))

	s.HeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.BackgroundColor)).
		Background(lipgloss.Color(theme.PrimaryColor)).
		Bold(true).
		Padding(0, 1)

	s.TitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Bold(true)

	s.SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedTextColor)).
		Bold(true)

	s.SidebarStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color(theme.BorderColor)).
		Padding(0, 1)

	s.ContentStyle = lipgloss.NewStyle().Padding(0, 1)

	s.ActiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.PrimaryColor)).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(theme.PrimaryColor)).
		Padding(0, 2).
		Bold(true)

	s.InactiveTabStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedTextColor)).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(theme.BorderColor)).
		Padding(0, 2)

	s.TabGapStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color(theme.BorderColor))

	s.DbStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedTextColor)).
		Bold(true)

	s.CollStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextColor)).
		PaddingLeft(2)

	s.CollSelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.BackgroundColor)).
		Background(lipgloss.Color(theme.PrimaryColor)).
		PaddingLeft(2).
		Bold(true)

	s.CardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.BorderColor)).
		Padding(1)

	s.CardActiveStyle = s.CardStyle.Copy().
		BorderForeground(lipgloss.Color(theme.PrimaryColor))

	s.SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SuccessColor))
	s.ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DangerColor))
	s.InfoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.InfoColor))
	s.WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.WarningColor))

	s.KeyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Syntax.ColorPurple)).Bold(true)
	s.ValStrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Syntax.ColorTeal))
	s.ValNumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Syntax.ColorAmber))
	s.ValBoolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Syntax.ColorBlue))
	s.SymbolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.MutedTextColor))

	s.FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MutedTextColor)).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color(theme.BorderColor))

	s.LabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.PrimaryColor))
	s.InputLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextColor)).Bold(true)
	s.TextMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.MutedTextColor))

	return s
}

//const (
//	ColorGreen    = "#00ED64"
//	ColorSlateBg  = "#1C1D1F"
//	ColorSlateSel = "#323338"
//	ColorBorder   = "#43464E"
//	ColorText     = "#E3E4E6"
//	ColorMuted    = "#888B94"
//	ColorAmber    = "#F59E0B"
//	ColorRed      = "#EF4444"
//	ColorTeal     = "#0D9488"
//	ColorPurple   = "#A855F7"
//	ColorBlue     = "#3B82F6"
//	ColorIndigo   = "#4F46E5"
//)

//var (
//	MainStyle = lipgloss.NewStyle().
//			Foreground(lipgloss.Color(ColorText)).
//			Background(lipgloss.Color(ColorSlateBg))
//
//	HeaderStyle = lipgloss.NewStyle().
//			Foreground(lipgloss.Color(ColorSlateBg)).
//			Background(lipgloss.Color(ColorGreen)).
//			Bold(true).
//			Padding(0, 1)
//
//	TitleStyle = lipgloss.NewStyle().
//			Foreground(lipgloss.Color(ColorGreen)).
//			Bold(true)
//
//	SidebarStyle = lipgloss.NewStyle().
//			Border(lipgloss.NormalBorder(), false, true, false, false).
//			BorderForeground(lipgloss.Color(ColorBorder)).
//			Padding(0, 1)
//
//	ContentStyle = lipgloss.NewStyle().
//			Padding(0, 1)
//
//	ActiveTabStyle = lipgloss.NewStyle().
//			Foreground(lipgloss.Color(ColorGreen)).
//			Border(lipgloss.NormalBorder(), false, false, true, false).
//			BorderForeground(lipgloss.Color(ColorGreen)).
//			Padding(0, 2).
//			Bold(true)
//
//	InactiveTabStyle = lipgloss.NewStyle().
//				Foreground(lipgloss.Color(ColorMuted)).
//				Border(lipgloss.NormalBorder(), false, false, true, false).
//				BorderForeground(lipgloss.Color(ColorBorder)).
//				Padding(0, 2)
//
//	TabGapStyle = lipgloss.NewStyle().
//			Border(lipgloss.NormalBorder(), false, false, true, false).
//			BorderForeground(lipgloss.Color(ColorBorder))
//
//	DbStyle = lipgloss.NewStyle().
//		Foreground(lipgloss.Color(ColorMuted)).
//		Bold(true)
//
//	CollStyle = lipgloss.NewStyle().
//			Foreground(lipgloss.Color(ColorText)).
//			PaddingLeft(2)
//
//	CollSelectedStyle = lipgloss.NewStyle().
//				Foreground(lipgloss.Color(ColorSlateBg)).
//				Background(lipgloss.Color(ColorGreen)).
//				PaddingLeft(2).
//				Bold(true)
//
//	CardStyle = lipgloss.NewStyle().
//			Border(lipgloss.RoundedBorder()).
//			BorderForeground(lipgloss.Color(ColorBorder)).
//			Padding(1)
//
//	CardActiveStyle = CardStyle.Copy().
//			BorderForeground(lipgloss.Color(ColorGreen))
//
//	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))
//	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed))
//	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue))
//	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAmber))
//
//	KeyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPurple)).Bold(true)
//	ValStrStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal))
//	ValNumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAmber))
//	ValBoolStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue))
//	SymbolStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorMuted))
//)

func (s Styles) HighlightJSON(jsonStr string) string {
	keyRegex := regexp.MustCompile(`"([^"]+)"\s*:`)
	strValRegex := regexp.MustCompile(`:\s*"([^"]*)"`)
	boolValRegex := regexp.MustCompile(`:\s*(true|false|null)\b`)
	numValRegex := regexp.MustCompile(`:\s*(-?\d+(\.\d+)?([eE][+-]?\d+)?)\b`)
	extKeyRegex := regexp.MustCompile(`"\$([a-zA-Z0-9]+)"`)

	lines := strings.Split(jsonStr, "\n")
	highlighted := make([]string, len(lines))

	for i, line := range lines {
		line = keyRegex.ReplaceAllStringFunc(line, func(match string) string {
			key := keyRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s%s", s.KeyStyle.Render(fmt.Sprintf(`"%s"`, key)), s.SymbolStyle.Render(":"))
		})

		line = strValRegex.ReplaceAllStringFunc(line, func(match string) string {
			val := strValRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s %s", s.SymbolStyle.Render(":"), s.ValStrStyle.Render(fmt.Sprintf(`"%s"`, val)))
		})

		line = boolValRegex.ReplaceAllStringFunc(line, func(match string) string {
			val := boolValRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s %s", s.SymbolStyle.Render(":"), s.ValBoolStyle.Render(val))
		})

		line = numValRegex.ReplaceAllStringFunc(line, func(match string) string {
			val := numValRegex.FindStringSubmatch(match)[1]
			return fmt.Sprintf("%s %s", s.SymbolStyle.Render(":"), s.ValNumStyle.Render(val))
		})

		line = extKeyRegex.ReplaceAllStringFunc(line, func(match string) string {
			return s.InfoStyle.Render(match)
		})

		highlighted[i] = line
	}

	return strings.Join(highlighted, "\n")
}
