package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Theme ThemeConfig `toml:"theme"`
}

type ThemeConfig struct {
	PrimaryColor    string            `toml:"primary_color"`
	SecondaryColor  string            `toml:"secondary_color"`
	SelectionColor  string            `toml:"selection_color"`
	BackgroundColor string            `toml:"background_color"`
	BorderColor     string            `toml:"border_color"`
	TextColor       string            `toml:"text_color"`
	MutedTextColor  string            `toml:"muted_text_color"`
	SuccessColor    string            `toml:"success_color"`
	WarningColor    string            `toml:"warning_color"`
	DangerColor     string            `toml:"danger_color"`
	InfoColor       string            `toml:"info_color"`
	Syntax          SyntaxThemeConfig `toml:"syntax_highlighting"`
}

type SyntaxThemeConfig struct {
	ColorDefault string `toml:"color_default"`
	ColorAmber   string `toml:"color_amber"`
	ColorRed     string `toml:"color_red"`
	ColorTeal    string `toml:"color_teal"`
	ColorPurple  string `toml:"color_purple"`
	ColorBlue    string `toml:"color_blue"`
	ColorIndigo  string `toml:"color_indigo"`
}

func GetConfigPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(userConfigDir, "compass-cli")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.toml"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	config := &Config{
		Theme: ThemeConfig{
			PrimaryColor:    "#00ED64",
			SecondaryColor:  "#A855F7",
			SelectionColor:  "#323338",
			BackgroundColor: "#1C1D1F",
			BorderColor:     "#43464E",
			TextColor:       "#E3E4E6",
			MutedTextColor:  "#888B94",
			SuccessColor:    "#519872",
			WarningColor:    "#F59E0B",
			DangerColor:     "#EF4444",
			InfoColor:       "#3B82F6",
			Syntax: SyntaxThemeConfig{
				ColorDefault: "#E3E4E6",
				ColorAmber:   "#F59E0B",
				ColorRed:     "#EF4444",
				ColorTeal:    "#0D9488",
				ColorPurple:  "#A855F7",
				ColorBlue:    "#3B82F6",
				ColorIndigo:  "#4F46E5",
			},
		},
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config, nil
		}
		return nil, err
	}

	err = toml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}
