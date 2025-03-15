package main

import "os"
import "github.com/charmbracelet/log"
import "github.com/charmbracelet/lipgloss"

func GetLogger() *log.Logger {
	styles := log.DefaultStyles()
	styles.Keys["ToonName"] = lipgloss.NewStyle().Foreground(lipgloss.Color("201"))
	styles.Values["ToonName"] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("201"))
	l := log.New(os.Stdout)
	l.SetStyles(styles)
	return l
}
