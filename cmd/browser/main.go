package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/jyotishmoy12/go-browser/internal/css"
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"github.com/jyotishmoy12/go-browser/internal/layout"
	"github.com/jyotishmoy12/go-browser/internal/network"
	"github.com/jyotishmoy12/go-browser/internal/parser"
	"github.com/jyotishmoy12/go-browser/internal/style"
)

type BrowserApp struct {
	layoutRoot *layout.LayoutBox
	rawInput   string
	scrollY    float64
}

func (g *BrowserApp) Update() error {
	chars := ebiten.AppendInputChars(nil)
	for _, c := range chars {
		g.rawInput += string(c)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(g.rawInput) > 0 {
		g.rawInput = g.rawInput[:len(g.rawInput)-1]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.navigate(g.rawInput)
	}

	_, wheelY := ebiten.Wheel()
	g.scrollY += wheelY * 20

	return nil
}

func (g *BrowserApp) navigate(url string) {
	log.Printf("Navigating to: %s", url)

	raw, err := network.Fetch(url)
	if err != nil {
		log.Printf("Fetch error: %v", err)
		return
	}
	domRoot := parser.New(raw).Parse()
	sheet := css.StyleSheet{
		Rules: []css.Rule{
			{Selectors: []string{"h1"}, Declarations: []css.Declaration{{Property: "color", Value: "red"}}},
		},
	}
	styledRoot := style.CreateStyledTree(domRoot, sheet)
	newLayoutRoot := buildLayoutTree(styledRoot)
	newLayoutRoot.Layout(layout.Rect{Width: 800})
	g.layoutRoot = newLayoutRoot
}
func (g *BrowserApp) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	if g.layoutRoot != nil {
		g.paint(g.layoutRoot, screen, 40)
	}
	ebitenutil.DrawRect(screen, 0, 0, 800, 35, color.RGBA{45, 45, 45, 255})
	ebitenutil.DebugPrintAt(screen, "URL: "+g.rawInput, 10, 10)
}

func (g *BrowserApp) paint(box *layout.LayoutBox, screen *ebiten.Image, offsetY float64) {
	if box.StyledNode.Node.Type == dom.ElementNode {
		c := color.RGBA{200, 200, 200, 255}
		if box.StyledNode.Specified["color"] == "red" {
			c = color.RGBA{255, 0, 0, 255}
		}
		ebitenutil.DrawRect(screen,
			float64(box.Dimensions.X),
			float64(box.Dimensions.Y)+offsetY,
			float64(box.Dimensions.Width),
			float64(box.Dimensions.Height),
			c)
	} else {
		ebitenutil.DebugPrintAt(screen, box.StyledNode.Node.Text, int(box.Dimensions.X), int(box.Dimensions.Y)+int(offsetY))
	}
	for _, child := range box.Children {
		g.paint(child, screen, offsetY)
	}
}

func (g *BrowserApp) Layout(w, h int) (int, int) { return w, h }

func main() {
	initialURL := "http://httpbin.org/html"

	app := &BrowserApp{
		rawInput: initialURL,
	}

	app.navigate(initialURL)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Go Browser Engine - Built From Scratch")
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
func buildLayoutTree(sNode *style.StyledNode) *layout.LayoutBox {
	box := &layout.LayoutBox{StyledNode: sNode}
	for _, child := range sNode.Children {
		box.Children = append(box.Children, buildLayoutTree(child))
	}
	return box
}
