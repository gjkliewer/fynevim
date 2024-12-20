# fynevim
Fynevim is both a [Fyne](https://fyne.io/) widget that can be included in other Fyne apps as well as a standalone GUI for [Neovim](https://neovim.io/).

**Warning:** fynevim is still under development, there may be breaking changes until v1 is released.

## Standalone editor installation
1. Ensure go is installed
2. Run `make install`

## Library usage
Here's a minimal example of creating a text editor with fynevim:
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
