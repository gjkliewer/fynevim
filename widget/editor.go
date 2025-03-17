package widget

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

const nvimRows = 10
const nvimCols = 25

type FyneStyleTable map[int]widget.TextGridStyle

type logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type noopLogger struct{}

func (n noopLogger) Debug(msg string, args ...any) {}
func (n noopLogger) Info(msg string, args ...any)  {}
func (n noopLogger) Error(msg string, args ...any) {}
func (n noopLogger) Warn(msg string, args ...any)  {}

type Editor struct {
	widget.BaseWidget

	log  logger
	Nvim *nvim.Nvim

	// graphical elements
	content         *widget.TextGrid
	cmdline         *widget.TextGrid
	cursor          cursor
	markdownPreview *widget.RichText

	// state
	previewMode bool
	controlHeld bool

	// nvim ui-linegrid events
	gridCursorGoto    *GridCursorGoto
	winViewport       WinViewport
	hlTable           HightlightTable
	styleTable        FyneStyleTable
	gridLineUpdates   []GridLine
	currentMode       ModeChange
	gridScrollUpdates []GridScroll
	modeInfoSet       ModeInfoSet
	gridResize        *GridResize
	gridClear         int

	// nvim ui-cmdline events
	cmdlineShow CmdlineShow
}

// Tappable interface
func (e *Editor) Tapped(pe *fyne.PointEvent) {
	e.debug("Tapped")
	e.debug("focusing editor")
	fyne.CurrentApp().Driver().CanvasForObject(e).Focus(e)
}

// Focusable interface
func (e *Editor) FocusGained() {
	// e.debug("focus gained")
}

// Focusable interface
func (e *Editor) FocusLost() {
	// e.debug("focus lost")
}

func (e *Editor) KeyDown(fke *fyne.KeyEvent) {
	if e.controlHeld {
		input := fmt.Sprintf("<C-%s>", fke.Name)
		_, err := e.Nvim.Input(input)
		if err != nil {
			e.debug("error in nvim.Input", "error", err)
		}
		return
	}

	switch fke.Name {
	case desktop.KeyControlLeft, desktop.KeyControlRight:
		e.controlHeld = true
		e.debug("control down")
	default:
		e.debug("unhandled key down, ignoring")
		return
	}
}

func (e *Editor) KeyUp(fke *fyne.KeyEvent) {
	switch fke.Name {
	case desktop.KeyControlLeft, desktop.KeyControlRight:
		e.controlHeld = false
		e.debug("control up")
	default:
		e.debug("unhandled key up, ignoring")
		return
	}
}

// handle normal key input
// Focusable interface
func (e *Editor) TypedRune(r rune) {
	if e.previewMode {
		return
	}

	input := string(r)

	if input == "<" {
		input = "<LT>"
	}

	e.debug("received typed rune input", "rune", input)

	_, err := e.Nvim.Input(input)
	if err != nil {
		e.debug("error in nvim.Input", "error", err)
	}
}

// handle special key input
// Focusable interface
func (e *Editor) TypedKey(fke *fyne.KeyEvent) {
	e.debug("received typed key input", "fke", fke)

	var vimKeycode string
	switch fke.Name {
	case fyne.KeyReturn:
		vimKeycode = "<CR>"
	case fyne.KeyBackspace:
		vimKeycode = "<BS>"
	case fyne.KeyEscape:
		vimKeycode = "<Esc>"
		if e.previewMode {
			e.ExitPreviewMode()
		}
	case fyne.KeyTab:
		vimKeycode = "<Tab>"
	default:
		e.debug("unhandled key input, ignoring")
		return
	}
	e.debug("mapped fyne key event to vim keycode", "fke", fke, "vimKeycode", vimKeycode)

	_, err := e.Nvim.Input(vimKeycode)
	if err != nil {
		e.debug("error in nvim.Input", "error", err)
	}
}

func (e *Editor) debug(msg string, args ...any) {
	e.log.Debug("fynevim/widget/editor "+msg, args...)
}
func (e *Editor) info(msg string, args ...any) {
	e.log.Info("fynevim/widget/editor "+msg, args...)
}

type cursor struct {
	row   int
	col   int
	image *canvas.Rectangle
	text  *canvas.Text
}

func (e *Editor) resizeContent(newSize fyne.Size) {
	cellSize := e.cellSize()
	cols := int(newSize.Width / cellSize.Width)
	rows := int(newSize.Height / cellSize.Height)
	e.debug("resizeContent", "newSize", newSize, "cellSize", cellSize, "row", rows, "cols", cols)
	e.Nvim.TryResizeUI(cols, rows)
	e.content.Resize(newSize)
}

func (e *Editor) resizeMarkdownPreview(newSize fyne.Size) {
	e.markdownPreview.Resize(newSize)
}

func (e *Editor) resizeCmdLine(newSize fyne.Size) {
	cellSize := e.cellSize()
	e.debug("resizeCmdLine", "newSize", newSize, "cellSize", cellSize)
	e.cmdline.Resize(newSize)
}

func (e *Editor) cellSize() fyne.Size {
	size := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true})
	size.Width = float32(math.Round(float64(size.Width)))
	size.Height = float32(math.Round(float64(size.Height)))
	return size
}

// WidgetRenderer interface
type renderer struct {
	e *Editor
}

func (r *renderer) Layout(s fyne.Size) {
	// cellSize := r.e.cellSize()
	// cmdLineSize := fyne.NewSize(s.Width, cellSize.Height)
	// set aside some room for the cmdline
	// contentSize := fyne.NewSize(s.Width, s.Height-cmdLineSize.Height)
	r.e.resizeContent(s)
	r.e.resizeMarkdownPreview(s)
	// r.e.resizeCmdLine(cmdLineSize)
	// r.e.cmdline.Move(fyne.NewPos(0, contentSize.Height))
}

func (r *renderer) MinSize() fyne.Size {
	cellSize := r.e.cellSize()
	return fyne.NewSize(60*cellSize.Width, 30*cellSize.Height)
}

func (r *renderer) Refresh() {
	r.e.content.Refresh() // this is needed to redraw screen when a new line is added or removed
	r.e.drawCursor()
}

func (r *renderer) Objects() []fyne.CanvasObject {
	o := []fyne.CanvasObject{
		// r.e.cmdline,
	}

	if r.e.previewMode {
		o = append(o, r.e.markdownPreview)
	} else {
		o = append(o, r.e.content)
		o = append(o, r.e.cursor.image)
		o = append(o, r.e.cursor.text)
	}

	return o
}

func (r *renderer) Destroy() {
}

func (e *Editor) CreateRenderer() fyne.WidgetRenderer {
	e.cursor.image = canvas.NewRectangle(theme.ErrorColor()) // TODO error color
	e.cursor.text = canvas.NewText("", theme.ErrorColor())   // TODO error color
	e.cursor.text.TextStyle = fyne.TextStyle{Monospace: true}
	e.cursor.image.Resize(e.cellSize())

	return &renderer{e: e}
}

func (e *Editor) drawCursor() {
	modeInfo := e.modeInfoSet.modeInfo[e.currentMode.ModeIdx]
	row := e.content.Row(e.cursor.row)
	cell := row.Cells[e.cursor.col]

	var fgColor color.Color
	var bgColor color.Color
	if modeInfo.AttrId > 0 {
		fgColor = e.hlTable[modeInfo.AttrId].Foreground
		bgColor = e.hlTable[modeInfo.AttrId].Background
	} else {
		// if modeInfo is 0 then we should invert the cell colors
		var cellStyle = cell.Style
		if cellStyle == nil {
			cellStyle = defaultStyle()
		}
		fgColor = cellStyle.BackgroundColor()
		bgColor = cellStyle.TextColor()
	}

	cellSize := e.cellSize()
	cursorPos := fyne.NewPos(cellSize.Width*float32(e.cursor.col), cellSize.Height*float32(e.cursor.row))
	var cursorSize fyne.Size
	var text string
	switch modeInfo.CursorShape {
	case "block":
		cursorSize = cellSize
		text = string(cell.Rune)
		e.cursor.text.Move(cursorPos)
		e.cursor.text.Color = fgColor
		e.cursor.text.Text = text
		e.cursor.text.Resize(cellSize)
		e.cursor.text.Show()
	case "horizontal":
		cursorSize = fyne.NewSize(cellSize.Width, 2)
		cursorPos = cursorPos.AddXY(0, cellSize.Height)
		e.cursor.text.Hide()
	case "vertical":
		cursorSize = fyne.NewSize(2, cellSize.Height)
		e.cursor.text.Hide()
	default:
		panic(fmt.Sprintf("unexpected cursor %v", modeInfo.CursorShape))
	}
	e.cursor.image.Resize(cursorSize)
	e.cursor.image.Move(cursorPos)
	e.cursor.image.FillColor = bgColor
	e.cursor.image.Refresh()
	e.cursor.text.Refresh()
}

func NewEditor(log logger, nvimProcessOptions []nvim.ChildProcessOption) *Editor {
	e := &Editor{
		log:             log,
		content:         widget.NewTextGrid(),
		cmdline:         widget.NewTextGrid(),
		markdownPreview: widget.NewRichText(),
	}
	e.ExtendBaseWidget(e)
	e.content.ShowLineNumbers = false
	e.content.ShowWhitespace = false
	e.markdownPreview.Scroll = container.ScrollVerticalOnly
	e.markdownPreview.Wrapping = fyne.TextWrapWord

	if e.log == nil {
		e.log = noopLogger{}
	}

	e.debug("starting nvim child process")
	var err error
	e.Nvim, err = nvim.NewChildProcess(nvimProcessOptions...)
	if err != nil {
		panic(err)
	}

	err = e.Nvim.SetClientInfo(
		"fynevim",
		nvim.ClientVersion{Major: 0, Minor: 0, Patch: 0},
		nvim.EmbedderClientType,
		map[string]*nvim.ClientMethod{},
		nvim.ClientAttributes{},
	)
	if err != nil {
		panic(err)
	}

	e.hlTable = HightlightTable{}

	// handle redraw events from nvim
	e.info("registering redraw handler")
	e.Nvim.RegisterHandler("redraw", e.handleNvimEvents)

	e.info("attaching ui to nvim")
	err = e.Nvim.AttachUI(nvimCols, nvimRows, map[string]any{
		"ext_linegrid": true,
		"ext_hlstate":  true,
		// "ext_cmdline":  true,
		// "ext_messages": true,
		// "ext_tabline":  true,
		// "term_background": "dark",
		// "ext_termcolors":  true,
	})
	if err != nil {
		panic(err)
	}

	// register preview command
	plug := plugin.New(e.Nvim)
	plug.HandleCommand(&plugin.CommandOptions{Name: "Preview", NArgs: "0"}, e.EnterPreviewMode)
	plug.RegisterForTests() // TODO this works but is this how these should be registered?

	return e
}

// EnterPreviewMode Renders the current buffer as markdown richtext
func (e *Editor) EnterPreviewMode() error {
	e.previewMode = true
	buf, err := e.Nvim.CurrentBuffer()
	if err != nil {
		return fmt.Errorf("Error getting buffer: %v", err)
	}
	bufLines, err := e.Nvim.BufferLines(buf, 0, -1, false)
	if err != nil {
		return fmt.Errorf("Error reading buffer lines: %v", err)
	}
	var contents string
	for _, line := range bufLines {
		contents += string(line) + "\n"
	}

	e.markdownPreview.ParseMarkdown(contents)
	return nil
}

func (e *Editor) ExitPreviewMode() {
	e.previewMode = false
	e.markdownPreview.ParseMarkdown("")
}

func (e *Editor) clearRow(row int) {
	for col := range e.content.Row(row).Cells {
		e.content.SetRune(row, col, ' ')
	}
}

func (e *Editor) handleNvimEvents(updates ...[]any) {
	for _, update := range updates {
		eventName := update[0].(string)
		eventData := update[1:]

		e.debug("nvim.redraw", "event", eventName) //, "data", eventData)

		switch eventName {
		case "mode_info_set":
			// modeInfoSet = []ModeInfo{}

			for _, d := range eventData {
				data := d.([]any)
				e.modeInfoSet.cursorStyleEnabled = data[0].(bool)
				modeInfoList := data[1].([]any)

				for _, mi := range modeInfoList {
					modeInfoMap := mi.(map[string]any)

					modeInfo := ModeInfo{}
					modeInfo.CursorShape, _ = modeInfoMap["cursor_shape"].(string)
					modeInfo.CellPercentage = toi(modeInfoMap["cell_percentage"])
					modeInfo.BlinkWait = toi(modeInfoMap["blinkwait"])
					modeInfo.BlinkOn = toi(modeInfoMap["blinkon"])
					modeInfo.BlinkOff = toi(modeInfoMap["blinkoff"])
					modeInfo.AttrId = toi(modeInfoMap["attr_id"])
					modeInfo.AttrIdLm = toi(modeInfoMap["attr_id_lm"])
					modeInfo.ShortName = modeInfoMap["short_name"].(string)
					modeInfo.Name = modeInfoMap["name"].(string)
					modeInfo.MouseShape = toi(modeInfoMap["mouse_shape"])

					// log.Debug("  mode_info_set: %+v", modeInfoMap)
					e.modeInfoSet.modeInfo = append(e.modeInfoSet.modeInfo, modeInfo)
				}
				// log.Debug("  mode_info_set: %+v", modeInfoSet)
			}

		case "mode_change":
			for _, d := range eventData {
				data := d.([]any)

				e.currentMode.Mode = data[0].(string)
				e.currentMode.ModeIdx = toi(data[1])
			}

		case "default_colors_set":
			for _, d := range eventData {
				data := d.([]any)
				fg := NewColor(toi(data[0]))
				bg := NewColor(toi(data[1]))
				sp := NewColor(toi(data[2]))
				e.hlTable[0] = HLAttribute{
					Foreground: fg,
					Background: bg,
					Special:    sp,
				}
			}

		case "hl_attr_define":
			for _, d := range eventData {
				id, attr := NewHLAttribute(d)

				e.hlTable[id] = attr
			}

		case "grid_resize":
			for _, d := range eventData {
				data := d.([]any)
				e.gridResize = &GridResize{
					Grid:   toi(data[0]),
					Width:  toi(data[1]),
					Height: toi(data[2]),
				}
				e.debug("grid_resize", "gridResize", e.gridResize)
			}

		case "grid_clear":
			for _, d := range eventData {
				data := d.([]any)
				e.gridClear = toi(data[0])
			}

		case "grid_cursor_goto":
			for _, d := range eventData {
				e.gridCursorGoto = NewGridCursorGoto(d)
			}

		case "grid_line":
			// [1 250 152 [[  0 5] [t] [o] [ ] [o] [p] [t] [i] [m] [i] [z] [e] [ ] [N] [v] [i] [m]]]
			//
			// root  [grid, row, col_start, cells, wrap]
			//       [   1,  250,      152, [...], false]
			//
			// cell  [text, hl_id (optional), repeat (optional)]
			//       [   a,                0,                 5]

			// e.debug("grid_line", "eventData", eventData)

			for _, l := range eventData {
				gridLineData := l.([]any)

				gl := GridLine{
					Grid:     toi(gridLineData[0]),
					Row:      toi(gridLineData[1]),
					ColStart: toi(gridLineData[2]),
					Wrap:     gridLineData[4].(bool),
				}

				cells := gridLineData[3].([]any)
				var highlightID int
				for _, c := range cells {
					cellData := c.([]any)
					cell := Cell{}

					t := cellData[0].(string)
					cell.Text = []rune(t)[0]

					if len(cellData) > 1 {
						highlightID = toi(cellData[1])
					}
					cell.HighlightID = highlightID

					if len(cellData) > 2 {
						cell.Repeat = toi(cellData[2])
					} else {
						cell.Repeat = 1
					}

					gl.Cells = append(gl.Cells, cell)
				}

				e.gridLineUpdates = append(e.gridLineUpdates, gl)
			}

		case "win_viewport":
			ed := eventData[0].([]any)

			e.winViewport = WinViewport{
				Grid:    toi(ed[0]),
				Win:     ed[1].(nvim.Window),
				Topline: toi(ed[2]),
				Botline: toi(ed[3]),
				Curline: toi(ed[4]),
				Curcol:  toi(ed[5]),
			}

		case "grid_scroll":
			ed := eventData[0].([]any)

			gridScroll := GridScroll{
				Grid:  toi(ed[0]),
				Top:   toi(ed[1]),
				Bot:   toi(ed[2]),
				Left:  toi(ed[3]),
				Right: toi(ed[4]),
				Rows:  toi(ed[5]),
				Cols:  toi(ed[6]),
			}
			e.gridScrollUpdates = append(e.gridScrollUpdates, gridScroll)

		case "flush":
			if e.gridClear > 0 {
				for r := range e.content.Rows {
					e.clearRow(r)
				}
				e.gridClear = 0
			}

			// grid resize
			if e.gridResize != nil {
				rows := e.gridResize.Height
				cols := e.gridResize.Width
				// add extra rows to the bottom of the grid if needed
				for r := len(e.content.Rows) - 1; r < rows; r++ {
					e.content.SetRow(r, widget.TextGridRow{})
				}

				// remove extra rows from the bottom of the grid if needed
				e.content.Rows = e.content.Rows[:rows]

				// add/remove columns
				for r, row := range e.content.Rows {
					// add empty columns on right of grid if needed
					for c := len(row.Cells) - 1; c < cols; c++ {
						row.Cells = append(row.Cells, widget.TextGridCell{Rune: ' '})
					}

					// remove extra columns columns on right of grid if needed
					for c := cols; c < len(row.Cells); c++ {
						e.content.SetRune(r, c, ' ') // zero out cell
					}
					row.Cells = row.Cells[:cols]
				}
				e.gridResize = nil
			}

			// grid scrolls
			for _, gridScroll := range e.gridScrollUpdates {
				e.debug("handling grid scroll", "rows", gridScroll.Rows)

				if gridScroll.Rows > 0 { // scroll down; move rows up
					for fromRow := gridScroll.Top + gridScroll.Rows; fromRow < gridScroll.Bot; fromRow++ {
						row := e.content.Row(fromRow)
						toRow := fromRow - gridScroll.Rows
						r := widget.TextGridRow{Style: row.Style}
						r.Cells = append(r.Cells, row.Cells...)
						e.content.SetRow(toRow, r)
					}
				} else if gridScroll.Rows < 0 { // scroll up; move rows down
					for fromRow := gridScroll.Bot - 1 + gridScroll.Rows; fromRow >= gridScroll.Top; fromRow-- {
						row := e.content.Row(fromRow)
						toRow := fromRow - gridScroll.Rows
						r := widget.TextGridRow{Style: row.Style}
						r.Cells = append(r.Cells, row.Cells...)
						e.content.SetRow(toRow, r)
					}
				}
			}
			e.gridScrollUpdates = nil

			// grid line updates
			for _, gl := range e.gridLineUpdates {
				col := gl.ColStart

				for _, cell := range gl.Cells {
					style := e.hlTable.GetTextGridStyle(cell.HighlightID)
					for repeat := 0; repeat < cell.Repeat; repeat++ {
						e.content.SetRune(gl.Row, col, cell.Text)
						e.content.SetStyle(gl.Row, col, style)

						col += 1
					}
					// e.info("render", "l", gl.Row, "t", lineText, "hl_id", cell.HighlightID, "attr", attr.String())
				}
			}
			e.gridLineUpdates = nil

			// update cursor position
			if e.gridCursorGoto != nil {
				e.cursor.row = e.gridCursorGoto.Row
				e.cursor.col = e.gridCursorGoto.Column
				e.gridCursorGoto = nil
			}

			e.debug("refreshing editor")
			e.Refresh()

		case "cmdline_show":
			for _, ed := range eventData {
				e.cmdlineShow = NewCmlineShow(ed)
			}
			row := widget.TextGridRow{
				Cells: []widget.TextGridCell{
					{
						Rune: []rune(e.cmdlineShow.firstc)[0],
					},
				},
			}
			e.debug("cmdline_show", "cmdlineShow", e.cmdlineShow)
			for _, content := range e.cmdlineShow.content {
				for _, c := range content.content {
					row.Cells = append(row.Cells, widget.TextGridCell{
						Rune: c,
					})
				}
			}

			e.cmdline.SetRow(0, row)
		}
	}
}
