package main

import (
	"log/slog"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/neovim/go-client/nvim"

	fynevim "github.com/gjkliewer/fynevim/widget"
)

const defaultWinWidth = 1000
const defaultWinHeight = 800

var editor *fynevim.Editor
var richText *widget.RichText
var log *slog.Logger

func main() {
	log = initLogger()

	a := app.New()
	window := a.NewWindow("Neovim")
	window.SetPadded(false)
	window.Resize(fyne.NewSize(defaultWinWidth, defaultWinHeight))

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	editor = fynevim.NewEditor(
		log,
		[]nvim.ChildProcessOption{
			// TODO Did this because can't find command 'nvim' when started from macos Finder,
			//      need to make this work for more than just nvim homebrew install
			nvim.ChildProcessCommand("/opt/homebrew/bin/nvim"),
			nvim.ChildProcessArgs(
				"--embed",
			),
			nvim.ChildProcessDir(dir),
		},
	)
	defer editor.Nvim.Close()

	window.SetContent(editor)
	log.Debug("starting window")
	window.ShowAndRun()
}

func logLevel() slog.Level {
	switch level := os.Getenv("FYNEVIM_LOG_LEVEL"); level {
	case "DEBUG":
		return slog.LevelDebug
	default:
		return slog.LevelError
	}
}

func initLogger() *slog.Logger {
	var programLevel = new(slog.LevelVar)
	programLevel.Set(logLevel())
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	l := slog.New(h)
	return l
}
