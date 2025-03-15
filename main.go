package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/neovim/go-client/nvim"

	fynevim "github.com/gjkliewer/fynevim/widget"
)

const defaultWinWidth = 1000
const defaultWinHeight = 800

var log *slog.Logger

func main() {
	forkFlag := flag.Bool("fork", false, "Spawn a child process")
	flag.Parse()

	if *forkFlag {
		// Create a command that will execute the current program
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		dir, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("Error getting working directory: %v", err))
		}
		cmd.Dir = dir

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

	dir := os.Getenv("PWD")

	childProcessArgs := append([]string{"--embed"}, os.Args[1:]...)

	editor := fynevim.NewEditor(
		log,
		[]nvim.ChildProcessOption{
			nvim.ChildProcessCommand("nvim"), // nvim must be in PATH
			nvim.ChildProcessArgs(childProcessArgs...),
			nvim.ChildProcessDir(dir),
			nvim.ChildProcessLogf(log.Debug),
		},
	)
	defer editor.Nvim.Close()

	cID := editor.Nvim.ChannelID()
	err := editor.Nvim.Command(fmt.Sprintf("autocmd VimLeave * call rpcnotify(%v, 'fynevim.VimLeave')", cID))
	if err != nil {
		panic(fmt.Sprintf("Could set autocmd: %v", err))
	}

	err = editor.Nvim.RegisterHandler("fynevim.VimLeave", func() error {
		window.Close()
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("Could not register handler: %v", err))
	}

	window.SetContent(editor)
	log.Debug("starting window")
	window.Canvas().Focus(editor)
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
