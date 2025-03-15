package widget

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/neovim/go-client/nvim"
)

// HightlightTable is used to keep track of nvims highlight ids.
type HightlightTable map[int]HLAttribute

func (hlt HightlightTable) GetTextGridStyle(hlID int) *widget.CustomTextGridStyle {
	attr := hlt[hlID]
	fg := attr.Foreground
	bg := attr.Background
	if fg == nil {
		fg = hlt[0].Foreground
	}
	if bg == nil {
		bg = hlt[0].Background
	}
	if attr.Reverse {
		fg = attr.Background
		bg = attr.Foreground
	}
	return &widget.CustomTextGridStyle{
		FGColor: fg,
		BGColor: bg,
		TextStyle: fyne.TextStyle{
			Bold:      attr.Bold,
			Italic:    attr.Italic,
			Underline: attr.Underline,
			Monospace: true,
			// Symbol:    false,
			// TabWidth: 2,
		},
	}
}

func NewHLAttribute(eventData any) (int, HLAttribute) {
	d := eventData.([]any)
	id := toi(d[0])
	attributes := d[1].(map[string]any)

	var foreground, background, special color.Color
	fg, ok := attributes["foreground"]
	if ok {
		foreground = NewColor(toi(fg))
	}
	bg, ok := attributes["background"]
	if ok {
		background = NewColor(toi(bg))
	}
	sp, ok := attributes["special"]
	if ok {
		special = NewColor(toi(sp))
	}
	reverse, _ := attributes["reverse"].(bool)
	italic, _ := attributes["italic"].(bool)
	bold, _ := attributes["bold"].(bool)
	strikethrough, _ := attributes["strikethrough"].(bool)
	underline, _ := attributes["underline"].(bool)
	undercurl, _ := attributes["undercurl"].(bool)
	underdouble, _ := attributes["underdouble"].(bool)
	underdotted, _ := attributes["underdotted"].(bool)
	underdashed, _ := attributes["underdashed"].(bool)
	blend := toi(attributes["blend"])

	// info := data[3].([]any)

	return id, HLAttribute{
		Foreground:    foreground,
		Background:    background,
		Special:       special,
		Reverse:       reverse,
		Italic:        italic,
		Bold:          bold,
		Strikethrough: strikethrough,
		Underline:     underline,
		Undercurl:     undercurl,
		Underdouble:   underdouble,
		Underdotted:   underdotted,
		Underdashed:   underdashed,
		Blend:         blend,
	}
}

type HLAttribute struct {
	Foreground    color.Color
	Background    color.Color
	Special       color.Color
	Reverse       bool
	Italic        bool
	Bold          bool
	Strikethrough bool
	Underline     bool
	Undercurl     bool
	Underdouble   bool
	Underdotted   bool
	Underdashed   bool
	Blend         int
}

func (attr *HLAttribute) String() string {
	var str []string
	if attr.Foreground != nil {
		str = append(str, fmt.Sprintf("FG:%v", attr.Foreground))
	}
	if attr.Background != nil {
		str = append(str, fmt.Sprintf("Background:%v", attr.Background))
	}
	if attr.Special != nil {
		str = append(str, fmt.Sprintf("Special:%v", attr.Special))
	}
	if attr.Reverse {
		str = append(str, fmt.Sprintf("Reverse:%v", attr.Reverse))
	}
	if attr.Italic {
		str = append(str, fmt.Sprintf("Italic:%v", attr.Italic))
	}
	if attr.Bold {
		str = append(str, fmt.Sprintf("Bold:%v", attr.Bold))
	}
	if attr.Strikethrough {
		str = append(str, fmt.Sprintf("Striket:%v", attr.Strikethrough))
	}
	if attr.Underline {
		str = append(str, fmt.Sprintf("Underline:%v", attr.Underline))
	}
	if attr.Undercurl {
		str = append(str, fmt.Sprintf("Undercurl:%v", attr.Undercurl))
	}
	if attr.Underdotted {
		str = append(str, fmt.Sprintf("Underdotted:%v", attr.Underdotted))
	}
	if attr.Underdouble {
		str = append(str, fmt.Sprintf("Underdouble:%v", attr.Underdotted))
	}
	if attr.Underdashed {
		str = append(str, fmt.Sprintf("Underdashed:%v", attr.Underdashed))
	}
	str = append(str, fmt.Sprintf("Blend:%v", attr.Blend))

	return strings.Join(str, " ")
}

type GridScroll struct {
	Grid  int
	Top   int
	Bot   int
	Left  int
	Right int
	Rows  int
	Cols  int
}

type GridResize struct {
	Grid   int
	Width  int
	Height int
}

type GridCursorGoto struct {
	Grid   int
	Row    int
	Column int
}

func NewGridCursorGoto(eventArgs any) *GridCursorGoto {
	ea := eventArgs.([]any)
	return &GridCursorGoto{
		Grid:   toi(ea[0]),
		Row:    toi(ea[1]),
		Column: toi(ea[2]),
	}
}

type WinViewport struct {
	Grid    int
	Win     nvim.Window
	Topline int
	Botline int
	Curline int
	Curcol  int
}

type GridLine struct {
	Grid     int
	Row      int
	ColStart int
	Cells    []Cell
	Wrap     bool
}

type Cell struct {
	Text        rune
	HighlightID int
	Repeat      int
}

type ModeInfoSet struct {
	cursorStyleEnabled bool
	modeInfo           []ModeInfo // current mode is given by the mode_idx field of the mode_change event
}

type ModeInfo struct {
	CursorShape    string
	CellPercentage int
	BlinkWait      int
	BlinkOn        int
	BlinkOff       int
	AttrId         int
	AttrIdLm       int
	ShortName      string
	Name           string
	MouseShape     int
}

type ModeChange struct {
	Mode    string
	ModeIdx int
}

type CmdlineShow struct {
	content []CmdlineContent
	pos     int
	firstc  string
	prompt  string
	indent  int
	level   int
}

type CmdlineContent struct {
	attrs   HLAttribute
	content string
}

// ["cmdline_show", content, pos, firstc, prompt, indent, level]
// content: List of [attrs, string] [[{}, "t"], [attrs, "est"], ...]
func NewCmlineShow(eventData any) (c CmdlineShow) {
	d := eventData.([]any)

	contentList := d[0].([]any)
	for _, content := range contentList {
		cell := content.([]any)
		// _, attrs := NewHLAttribute(cell)
		c.content = append(c.content, CmdlineContent{content: cell[1].(string)})
	}
	c.pos = toi(d[1])
	c.firstc = d[2].(string)
	c.prompt = d[3].(string)
	c.indent = toi(d[4])
	c.level = toi(d[5])

	return
}

// toi resolves interface containing uint64 or int64 to int
func toi(i any) int {
	if i == nil {
		return 0
	}

	i1, ok := i.(int64)
	if ok {
		return int(i1)
	}

	i2, ok := i.(uint64)
	if ok {
		return int(i2)
	}

	panic(fmt.Sprintf("toi: unable to convert %T %v to int", i, i))
}

func NewColor(c int) color.Color {
	return &color.NRGBA{
		R: uint8((c >> 16) & 0xFF),
		G: uint8((c >> 8) & 0xFF),
		B: uint8(c & 0xFF),
		A: 255,
	}
}

func defaultStyle() *widget.CustomTextGridStyle {
	return &widget.CustomTextGridStyle{
		FGColor: theme.ForegroundColor(),
		BGColor: theme.BackgroundColor(),
	}
}
