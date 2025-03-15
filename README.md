# fynevim
Fynevim is both a [Fyne](https://fyne.io/) widget that can be included in other Fyne apps as well as a standalone GUI for [Neovim](https://neovim.io/).

**Warning:** fynevim is still under development, there may be breaking changes until v1 is released.

## Features
- The command `:Pre[view]` allows for rendered previews of markdown files. Press `<Esc>` to exit preview mode.

## Standalone editor installation
1. Install neovim.
2. Install go.
2. Run `make install`.

## Library usage
Here's a minimal example of embedding a fyenvim text editor in a fyne app:
```go
package main

import (
	"fyne.io/fyne/v2/app"
	fynevim "github.com/gjkliewer/fynevim/widget"
	"github.com/neovim/go-client/nvim"
)

func main() {
	a := app.New()
	window := a.NewWindow("Neovim")

	editor := fynevim.NewEditor(
		nil,
		[]nvim.ChildProcessOption{
			nvim.ChildProcessCommand("nvim"), // nvim must be in PATH
			nvim.ChildProcessArgs(
				"--embed",
			),
		},
	)
	defer editor.Nvim.Close()

	window.SetContent(editor)
	window.ShowAndRun()
}
```
