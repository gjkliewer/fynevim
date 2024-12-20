package main

import (
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"fmt"

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
	forkFlag := flag.Bool("fork", false, "Spawn a child process")
	flag.Parse()

	if *forkFlag {
		// Create a command that will execute the current program
		cmd := exec.Command(os.Args[0])

		// Inherit the current process's environment variables
		cmd.Env = os.Environ()

		// Start the child process
		if err := cmd.Start(); err != nil {
			panic(fmt.Sprintf("Error forking: %v", err))
		}
	} else {
		startApp()
	}
}

func startApp() {
	log = initLogger()
	a := app.New()
	window := a.NewWindow("Neovim")
	window.SetPadded(false)
	window.Resize(fyne.NewSize(defaultWinWidth, defaultWinHeight))

	dir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Error getting working directory: %v", err))
	}

	editor = fynevim.NewEditor(
		log,
		[]nvim.ChildProcessOption{
			nvim.ChildProcessCommand("nvim"), // nvim must be in PATH
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
