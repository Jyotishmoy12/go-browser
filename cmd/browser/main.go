package main

import (
	"image/color"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/jyotishmoy12/go-browser/internal/css"
	"github.com/jyotishmoy12/go-browser/internal/dom"
	"github.com/jyotishmoy12/go-browser/internal/javascript"
	"github.com/jyotishmoy12/go-browser/internal/layout"
	"github.com/jyotishmoy12/go-browser/internal/network"
	"github.com/jyotishmoy12/go-browser/internal/parser"
	"github.com/jyotishmoy12/go-browser/internal/style"
)

type BrowserApp struct {
	layoutRoot *layout.LayoutBox
	rawInput   string
	scrollY    float64
	URL        string
	history    []string
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
		g.navigate(g.rawInput, true)
	}

	_, wheelY := ebiten.Wheel()
	g.scrollY += wheelY * 20

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		adjustedY := float64(mouseY) - (40 + g.scrollY)
		log.Printf("Click at: %d, %d (Adjusted Y: %f)", mouseX, mouseY, adjustedY)
		targetLink := g.hitTest(g.layoutRoot, float64(mouseX), adjustedY)
		if targetLink != "" {
			log.Printf("Found Link! Navigating to: %s", targetLink)
			g.navigate(resolveURL(g.URL, targetLink), true)
		} else {
			log.Printf("No link found at this coordinate.")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && len(g.history) > 0 {
		lastIndex := len(g.history) - 1
		previousURL := g.history[lastIndex]
		g.history = g.history[:lastIndex]
		g.navigate(previousURL, false)
	}

	return nil
}
func (g *BrowserApp) hitTest(box *layout.LayoutBox, x, y float64) string {
	if x >= float64(box.Dimensions.X) && x <= float64(box.Dimensions.X+box.Dimensions.Width) &&
		y >= float64(box.Dimensions.Y) && y <= float64(box.Dimensions.Y+box.Dimensions.Height) {
		if box.LinkURL != "" {
			return box.LinkURL
		}
	}
	for _, child := range box.Children {
		if url := g.hitTest(child, x, y); url != "" {
			return url
		}
	}
	return ""
}

func resolveURL(current, target string) string {
	if strings.HasPrefix(target, "http") {
		return target
	}
	lastSlash := strings.LastIndex(current, "/")
	if lastSlash == -1 {
		return target
	}
	baseURL := current[:lastSlash+1]
	if strings.HasPrefix(target, "/") {
		parts := strings.Split(current, "/")
		return parts[0] + "//" + parts[2] + target
	}

	return baseURL + target
}

func (g *BrowserApp) navigate(url string, saveToHistory bool) {
	if saveToHistory && g.URL != "" {
		g.history = append(g.history, g.URL)
	}
	g.URL = url
	g.rawInput = url
	log.Printf("Navigating to: %s", url)

	raw, err := network.Fetch(url)
	if err != nil {
		log.Printf("Fetch error: %v", err)
		return
	}
	domRoot := parser.New(raw).Parse()
	js := javascript.NewRuntime(domRoot)
	scripts := findScripts(domRoot)
	for _, s := range scripts {
		log.Printf("Executing JS...")
		js.Execute(s)
	}
	sheet := css.StyleSheet{
		Rules: []css.Rule{
			{Selectors: []string{"h1"}, Declarations: []css.Declaration{{Property: "color", Value: "red"}}},
		},
	}
	styledRoot := style.CreateStyledTree(domRoot, sheet)
	newLayoutRoot := buildLayoutTree(styledRoot, "")
	newLayoutRoot.Layout(layout.Rect{Width: 800})
	g.layoutRoot = newLayoutRoot
}
func findScripts(n *dom.Node) []string {
	var scripts []string
	if n.TagName == "script" && len(n.Children) > 0 {
		scripts = append(scripts, n.Children[0].Text)
	}
	for _, child := range n.Children {
		scripts = append(scripts, findScripts(child)...)
	}
	return scripts
}
func (g *BrowserApp) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{250, 250, 250, 255})
	if g.layoutRoot != nil {
		g.paint(g.layoutRoot, screen, 50+g.scrollY)
	}
	vector.FillRect(screen, 0, 0, 800, 45, color.RGBA{33, 33, 33, 255}, false)
	vector.FillRect(screen, 0, 45, 800, 2, color.RGBA{0, 120, 215, 255}, false)
	vector.FillRect(screen, 10, 8, 780, 28, color.RGBA{60, 60, 60, 255}, false)
	displayURL := g.rawInput
	if displayURL == "" {
		displayURL = "Type a URL and press Enter..."
	}
	ebitenutil.DebugPrintAt(screen, "  "+displayURL, 15, 14)
	if len(g.history) > 0 {
		ebitenutil.DebugPrintAt(screen, "← Back [Esc]", 700, 50)
	}
}

func (g *BrowserApp) paint(box *layout.LayoutBox, screen *ebiten.Image, offsetY float64) {
	if box.LinkURL != "" {
		vector.FillRect(screen,
			float32(box.Dimensions.X),
			float32(float64(box.Dimensions.Y)+offsetY+float64(box.Dimensions.Height)-2),
			float32(box.Dimensions.Width),
			1,
			color.RGBA{0, 102, 204, 255}, false)
	}

	if box.StyledNode.Node.Type == dom.ElementNode {
		if box.StyledNode.Specified["color"] == "red" {
			vector.FillRect(screen,
				float32(box.Dimensions.X),
				float32(float64(box.Dimensions.Y)+offsetY),
				float32(box.Dimensions.Width),
				float32(box.Dimensions.Height),
				color.RGBA{231, 76, 60, 100}, false)
		}
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

	app.navigate(initialURL, true)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Go Browser Engine - Built From Scratch")
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
func buildLayoutTree(sNode *style.StyledNode, parentLink string) *layout.LayoutBox {
	box := &layout.LayoutBox{
		StyledNode: sNode,
		LinkURL:    parentLink,
	}
	if sNode.Node.TagName == "a" {
		log.Printf("Found A tag! Attrs: %v", sNode.Node.Attr)
		if val, ok := sNode.Node.Attr["href"]; ok {
			box.LinkURL = val
		}
	}
	for _, child := range sNode.Children {
		box.Children = append(box.Children, buildLayoutTree(child, box.LinkURL))
	}
	return box
}
