# Go Browser Engine — Architecture

This document describes the complete internal architecture of the Go Browser Engine, tracing the full data flow from a URL entered by the user all the way to pixels rendered on screen.

---

## Table of Contents

1. [High-Level Pipeline](#1-high-level-pipeline)
2. [Package Dependency Graph](#2-package-dependency-graph)
3. [Stage 1 — Network: Fetching HTML](#3-stage-1--network-fetching-html)
4. [Stage 2 — Parser: Building the DOM](#4-stage-2--parser-building-the-dom)
5. [Stage 3 — DOM: Node Model](#5-stage-3--dom-node-model)
6. [Stage 4 — JavaScript Execution](#6-stage-4--javascript-execution)
7. [Stage 5 — CSS: Parsing Stylesheets](#7-stage-5--css-parsing-stylesheets)
8. [Stage 6 — Style: Computing Styles](#8-stage-6--style-computing-styles)
9. [Stage 7 — Layout: Box Model](#9-stage-7--layout-box-model)
10. [Stage 8 — Paint: Rendering to Screen](#10-stage-8--paint-rendering-to-screen)
11. [User Interaction Loop](#11-user-interaction-loop)
12. [Data Structures Cheat Sheet](#12-data-structures-cheat-sheet)

---

## 1. High-Level Pipeline

The browser engine is a sequential pipeline. Each stage transforms its input into a richer representation that the next stage consumes.

```mermaid
flowchart TD
    A([🌐 User types URL + presses Enter]) --> B

    subgraph NET["📦 internal/network"]
        B["Fetch(url)\nOpen raw TCP/TLS socket\nSend HTTP/1.1 GET request\nRead & decode response"]
    end

    B -->|"raw HTML string"| C

    subgraph PARSE["📦 internal/parser"]
        C["parser.New(html).Parse()\nTokenize HTML character-by-character\nBuild recursive DOM tree"]
    end

    C -->|"*dom.Node (tree)"| D

    subgraph JS["📦 internal/javascript"]
        D["NewRuntime(domRoot)\nFind all &lt;script&gt; tags\nExecute each script via otto VM"]
    end

    D -->|"mutated *dom.Node (tree)"| E

    subgraph CSS["📦 internal/css"]
        E["css.StyleSheet{Rules}\n(currently hardcoded)\n css.Parser available for\n parsing external CSS text"]
    end

    E -->|"css.StyleSheet"| F

    subgraph STYLE["📦 internal/style"]
        F["CreateStyledTree(domRoot, sheet)\nMatch CSS rules to each DOM node\nProduce computed property map"]
    end

    F -->|"*style.StyledNode (tree)"| G

    subgraph LAYOUT["📦 internal/layout"]
        G["buildLayoutTree(styledRoot)\nLayoutBox.Layout(containingBlock)\nCompute X, Y, Width, Height for every box"]
    end

    G -->|"*layout.LayoutBox (tree)"| H

    subgraph PAINT["📦 cmd/browser (Ebitengine)"]
        H["BrowserApp.Draw(screen)\npaint() walks LayoutBox tree\nDraws rectangles + text via ebiten"]
    end

    H --> I([🖥️ Pixels on screen])

    style NET fill:#1a3a5c,color:#fff,stroke:#4a90d9
    style PARSE fill:#1a4a2e,color:#fff,stroke:#4caf50
    style JS fill:#4a2e00,color:#fff,stroke:#ff9800
    style CSS fill:#3a1a4a,color:#fff,stroke:#9c27b0
    style STYLE fill:#1a3a4a,color:#fff,stroke:#00bcd4
    style LAYOUT fill:#4a1a1a,color:#fff,stroke:#f44336
    style PAINT fill:#2e2e1a,color:#fff,stroke:#cddc39
```

---

## 2. Package Dependency Graph

Shows which packages import which. The dependency flow is strictly top-down — no circular imports.

```mermaid
graph LR
    CMD["cmd/browser\nmain.go"]

    NET["internal/network\nrequest.go"]
    PARSE["internal/parser\nhtml.go"]
    DOM["internal/dom\nnode.go"]
    CSS["internal/css\ncss.go · parser.go"]
    STYLE["internal/style\nstyle.go"]
    LAYOUT["internal/layout\nlayout.go"]
    JS["internal/javascript\nruntime.go"]

    OTTO["github.com/robertkrimen/otto\n(JS VM)"]
    EBITEN["github.com/hajimehoshi/ebiten/v2\n(2D rendering)"]

    CMD --> NET
    CMD --> PARSE
    CMD --> CSS
    CMD --> STYLE
    CMD --> LAYOUT
    CMD --> JS
    CMD --> DOM
    CMD --> EBITEN

    PARSE --> DOM
    STYLE --> DOM
    STYLE --> CSS
    LAYOUT --> DOM
    LAYOUT --> STYLE
    JS --> DOM
    JS --> OTTO

    style CMD fill:#263238,color:#eceff1,stroke:#90a4ae
    style NET fill:#1a3a5c,color:#fff,stroke:#4a90d9
    style PARSE fill:#1a4a2e,color:#fff,stroke:#4caf50
    style DOM fill:#33691e,color:#fff,stroke:#8bc34a
    style CSS fill:#3a1a4a,color:#fff,stroke:#9c27b0
    style STYLE fill:#1a3a4a,color:#fff,stroke:#00bcd4
    style LAYOUT fill:#4a1a1a,color:#fff,stroke:#f44336
    style JS fill:#4a2e00,color:#fff,stroke:#ff9800
    style OTTO fill:#424242,color:#fff,stroke:#9e9e9e
    style EBITEN fill:#424242,color:#fff,stroke:#9e9e9e
```

---

## 3. Stage 1 — Network: Fetching HTML

**Package:** `internal/network` · **File:** `request.go`  
**Entry point:** `Fetch(url string) (string, error)`

The engine does **not** use Go's standard `net/http` package. Instead it opens a raw socket connection and manually constructs an HTTP/1.1 request.

### Flow

```mermaid
sequenceDiagram
    participant App as cmd/browser
    participant Net as network.Fetch()
    participant Sock as Raw TCP/TLS Socket
    participant Server as Remote HTTP Server

    App->>Net: Fetch("https://example.com/page")
    Note over Net: Detect scheme (http vs https)
    Note over Net: Parse host and path from URL

    alt HTTPS
        Net->>Sock: tls.Dial("tcp", host+":443", InsecureSkipVerify)
    else HTTP
        Net->>Sock: net.Dial("tcp", host+":80")
    end

    Net->>Sock: Write raw HTTP/1.1 GET request
    Note over Sock: GET /page HTTP/1.1\nHost: example.com\nUser-Agent: Go-Browser-Project/1.0\nConnection: close

    Server-->>Sock: HTTP response (headers + body)
    Sock-->>Net: Buffered response via bufio.Reader

    Note over Net: Read & discard status line
    loop Header lines
        Net->>Net: Read each header line
        Note over Net: Detect "Transfer-Encoding: chunked"
    end

    alt Chunked Transfer Encoding
        loop Chunks
            Net->>Net: Read hex chunk size
            Net->>Net: ReadFull(chunk bytes)
            Net->>Net: Discard CRLF trailer
        end
    else Content-Length / Close
        Net->>Net: io.Copy entire remaining body
    end

    Net-->>App: Return raw HTML string
```

### Key implementation details

| Detail | Implementation |
|---|---|
| Protocol detection | `strings.HasPrefix(url, "https://")` |
| TLS | `tls.Dial` with `InsecureSkipVerify: true` |
| HTTP version | HTTP/1.1 only |
| Header parsing | Line-by-line with `bufio.Reader.ReadString('\n')` |
| Chunked decoding | Manual hex-size → `io.ReadFull` loop |
| Body (non-chunked) | Single `io.Copy` into a `strings.Builder` |

---

## 4. Stage 2 — Parser: Building the DOM

**Package:** `internal/parser` · **File:** `html.go`  
**Entry point:** `New(html).Parse() *dom.Node`

The HTML parser is a hand-written **recursive descent** parser. It consumes the raw HTML string one character at a time and constructs a tree of `dom.Node` objects.

### Parsing State Machine

```mermaid
stateDiagram-v2
    [*] --> ParseNodes : Parse() called
    ParseNodes --> ConsumeWhitespace : loop start
    ConsumeWhitespace --> CheckEOF

    CheckEOF --> Done : EOF or closing tag detected
    CheckEOF --> ParseNode : more input remains

    ParseNode --> ParseElement : current char == '<'
    ParseNode --> ParseText : current char != '<'

    ParseElement --> ConsumeTagName : skip '<'
    ConsumeTagName --> ParseAttributes
    ParseAttributes --> ParseAttribute : more attrs
    ParseAttribute --> ParseAttributes : loop
    ParseAttributes --> ConsumeClosingBracket : char == '>'

    ConsumeClosingBracket --> ParseNodes : recurse for children
    ParseNodes --> ConsumeClosingTag : found '</'
    ConsumeClosingTag --> ParseNodes : return to parent

    ParseText --> ParseNodes : return text node

    Done --> [*] : return root Node
```

### Call Stack (Recursive Descent)

```mermaid
flowchart TD
    A["Parse()"] --> B["parseNodes() — loop"]
    B --> C{"next char?"}
    C -->|"'<'"| D["parseElement()"]
    C -->|"other"| E["parseText()\n→ dom.NewText(rawText)"]

    D --> F["consumeIdentifier() → tagName"]
    D --> G["parseAttributes() loop\n→ parseAttribute()\n  consumeIdentifier() = name\n  skip '='\n  read quoted value"]
    D --> H["skip '>'\nrecurse: parseNodes() → children"]
    D --> I["skip closing tag '</' + tagName + '>'"]
    D --> J["dom.NewElement(tagName, attrs)\n.AddChildren(children)"]

    style A fill:#1a4a2e,color:#fff
    style B fill:#1a4a2e,color:#fff
    style D fill:#1a4a2e,color:#fff
    style E fill:#2e3a1a,color:#fff
    style J fill:#2e3a1a,color:#fff
```

### Example transformation

```
Input HTML:
  <h1 class="title">Hello</h1>

Parsed DOM:
  Node{Type: ElementNode, TagName: "h1", Attr: {"class": "title"}}
    └── Node{Type: TextNode, Text: "Hello"}
```

> **Note:** The parser does not handle self-closing tags (e.g., `<br/>`, `<img/>`), `DOCTYPE`, HTML comments, or malformed markup. It is intentionally minimal.

---

## 5. Stage 3 — DOM: Node Model

**Package:** `internal/dom` · **File:** `node.go`

The DOM is the shared data structure that flows through every subsequent stage. It is a simple recursive tree.

```mermaid
classDiagram
    class NodeType {
        <<enumeration>>
        ElementNode = 0
        TextNode    = 1
    }

    class Node {
        +NodeType Type
        +string TagName
        +map[string]string Attr
        +string Text
        +[]*Node Children
        +NewElement(tagname, attrs) *Node
        +NewText(content) *Node
        +AddChildren(children) *Node
    }

    Node --> NodeType : has
    Node --> Node : Children (recursive)
```

**ElementNode** carries `TagName` (e.g., `"div"`, `"a"`) and `Attr` (e.g., `{"href": "/page"}`).  
**TextNode** carries only `Text` — the raw string content between tags.

---

## 6. Stage 4 — JavaScript Execution

**Package:** `internal/javascript` · **File:** `runtime.go`  
**Entry point:** `NewRuntime(root *dom.Node) *JSRuntime`

JavaScript runs **before** CSS styling and layout, allowing scripts to mutate the DOM tree first.

```mermaid
sequenceDiagram
    participant Nav as navigate()
    participant Find as findScripts()
    participant JSR as javascript.NewRuntime()
    participant Otto as otto.Otto VM

    Nav->>Find: Walk DOM tree for &lt;script&gt; nodes
    Note over Find: Collect Children[0].Text from every\nnode where TagName == "script"
    Find-->>Nav: []string of script source code

    Nav->>JSR: NewRuntime(domRoot)
    Note over JSR: Creates otto.Otto VM
    JSR->>Otto: vm.Set("console", {log: fn})
    Note over Otto: console.log → fmt.Printf to stdout
    JSR->>Otto: vm.Set("document", {title: "Go Browser Engine"})
    JSR-->>Nav: *JSRuntime

    loop For each script string
        Nav->>JSR: Execute(scriptSource)
        JSR->>Otto: vm.Run(scriptSource)
        Otto-->>JSR: (result, error)
        JSR-->>Nav: error (logged if non-nil)
    end
```

### JS API surface (current)

| JS global | Go binding | Behaviour |
|---|---|---|
| `console.log(msg)` | `fmt.Printf` | Prints to stdout |
| `document.title` | Static string | Always `"Go Browser Engine"` |

> Full `document.getElementById`, `innerHTML` mutations, and event listeners are **not yet implemented**.

---

## 7. Stage 5 — CSS: Parsing Stylesheets

**Package:** `internal/css` · **Files:** `css.go`, `parser.go`

### Data Model

```mermaid
classDiagram
    class StyleSheet {
        +[]Rule Rules
    }
    class Rule {
        +[]string Selectors
        +[]Declaration Declarations
    }
    class Declaration {
        +string Property
        +string Value
    }

    StyleSheet "1" --> "*" Rule : contains
    Rule "1" --> "*" Declaration : contains
```

### CSS Parser FSM

```mermaid
stateDiagram-v2
    [*] --> ParseSheet
    ParseSheet --> ConsumeWS
    ConsumeWS --> CheckEOF2
    CheckEOF2 --> Done2 : EOF
    CheckEOF2 --> ParseRule : more input

    ParseRule --> ParseSelectors
    ParseSelectors --> ReadSelector : consumeIdentifier()
    ReadSelector --> CheckComma
    CheckComma --> ReadSelector : ',' found (another selector)
    CheckComma --> ParseDeclarations : '{' found

    ParseDeclarations --> ReadDeclaration
    ReadDeclaration --> ReadProp : consumeIdentifier()
    ReadProp --> SkipColon
    SkipColon --> ReadValue : scan until ';' or '}'
    ReadValue --> ReadDeclaration : ';' found, loop
    ReadDeclaration --> ParseSheet : '}' found, rule done

    Done2 --> [*]
```

### Example

```css
/* Input CSS string */
h1, h2 { color: red; font-size: 24px; }

/* Resulting StyleSheet */
StyleSheet{
  Rules: [
    Rule{
      Selectors:    ["h1", "h2"],
      Declarations: [
        {Property: "color",     Value: "red"},
        {Property: "font-size", Value: "24px"},
      ],
    },
  ],
}
```

> **Current usage:** The `css.Parser` is available but the `navigate()` function currently defines the stylesheet **inline** in Go code (hardcoded `h1 { color: red }`). Parsing external `<link>` stylesheets or `<style>` blocks is not yet wired up.

---

## 8. Stage 6 — Style: Computing Styles

**Package:** `internal/style` · **File:** `style.go`  
**Entry point:** `CreateStyledTree(root *dom.Node, sheet css.StyleSheet) *StyledNode`

This stage **marries** the DOM tree with the CSS stylesheet, producing a mirror tree where every node carries its computed CSS properties.

```mermaid
flowchart TD
    A["CreateStyledTree(node, sheet)"] --> B["Create StyledNode\n{Node: node, Specified: {}}"]
    B --> C["Loop over sheet.Rules"]
    C --> D{"matches(node, rule)?"}
    D -->|"yes — tagName equals\none of rule.Selectors"| E["Copy all Declarations\ninto node.Specified map"]
    D -->|"no"| F["Skip rule"]
    E --> G
    F --> G["Loop over node.Children"]
    G --> H["Recurse: CreateStyledTree(child, sheet)"]
    H --> I["Append *StyledNode to sNode.Children"]
    I --> J["Return *StyledNode"]

    style A fill:#1a3a4a,color:#fff
    style B fill:#1a3a4a,color:#fff
    style D fill:#0d2b3a,color:#fff
    style E fill:#0a3a2a,color:#fff
    style J fill:#1a3a4a,color:#fff
```

### Data Model

```mermaid
classDiagram
    class StyledNode {
        +*dom.Node Node
        +PropertyMap Specified
        +[]*StyledNode Children
    }
    class PropertyMap {
        <<type alias>>
        map[string]string
    }

    StyledNode --> PropertyMap : Specified
    StyledNode --> StyledNode : Children (recursive)
    StyledNode --> Node : wraps
```

**Selector matching** is currently tag-name only — `node.TagName == selector`. Class, ID, attribute, and pseudo selectors are not yet implemented.

---

## 9. Stage 7 — Layout: Box Model

**Package:** `internal/layout` · **File:** `layout.go`  
**Entry point:** `LayoutBox.Layout(containingBlock Rect)`

Layout converts the styled tree into a **positioned box tree**. Every box gets concrete `X`, `Y`, `Width`, `Height` float32 values.

### Algorithm

```mermaid
flowchart TD
    A["Layout(containingBlock Rect)"] --> B{"Node type?"}

    B -->|"ElementNode AND\ntagName in\nh1/div/root/html/body/a"| C["Width = containingBlock.Width\n(full-width block element)"]
    B -->|"TextNode"| D["Width = len(Text) × 9px\n(monospace approximation)"]
    B -->|"Other element"| E["Width = 100px (fallback)"]

    C --> F["X = containingBlock.X\nY = containingBlock.Y"]
    D --> F
    E --> F

    F --> G["cursorY = Dimensions.Y"]
    G --> H{"Any children?"}
    H -->|"yes"| I["For each child:\nchild.Layout(Rect{X, Y: cursorY, Width})\ncursorY += child.Height"]
    I --> J{"Has LinkURL and\nHeight < 24?"}
    J -->|"yes"| K["Height = 24 (min clickable size)"]
    J -->|"no"| L["Height = cursorY − Dimensions.Y"]
    H -->|"no leaf node"| M["Height = 24 (default line height)"]

    K --> N["Return"]
    L --> N
    M --> N

    style A fill:#4a1a1a,color:#fff
    style B fill:#3a1010,color:#fff
    style C fill:#2a0a0a,color:#fff
    style I fill:#2a0a0a,color:#fff
```

### Data Model

```mermaid
classDiagram
    class Rect {
        +float32 X
        +float32 Y
        +float32 Width
        +float32 Height
    }
    class LayoutBox {
        +Rect Dimensions
        +*StyledNode StyledNode
        +[]*LayoutBox Children
        +string LinkURL
        +Layout(containingBlock Rect)
    }

    LayoutBox --> Rect : Dimensions
    LayoutBox --> StyledNode : wraps
    LayoutBox --> LayoutBox : Children (recursive)
```

**Link propagation:** When `buildLayoutTree` encounters an `<a>` element, it stores the `href` attribute in `LinkURL`. This value is propagated down to all descendant boxes so hit-testing can find the URL anywhere within the link's subtree.

---

## 10. Stage 8 — Paint: Rendering to Screen

**Package:** `cmd/browser` · handled inside `BrowserApp.Draw()` and `BrowserApp.paint()`

The final stage walks the `LayoutBox` tree and issues draw calls to Ebitengine.

```mermaid
flowchart TD
    A["Draw(screen)"] --> B["screen.Fill(#fafafa)\n(white background)"]
    B --> C{"layoutRoot != nil?"}
    C -->|"yes"| D["paint(layoutRoot, screen, offsetY=50+scrollY)"]
    C -->|"no"| E["Skip page content"]

    D --> F["Walk each LayoutBox recursively"]
    F --> G{"box.LinkURL != empty?"}
    G -->|"yes"| H["Draw 1px blue underline\nbelow the link box"]
    G -->|"no"| I["No underline"]

    H --> J{"box.StyledNode.Node.Type?"}
    I --> J

    J -->|"ElementNode AND\nSpecified[color] == red"| K["FillRect with semi-transparent\nred overlay (ERR highlight)"]
    J -->|"TextNode"| L["ebitenutil.DebugPrintAt\n(text, x, y + scrollOffset)"]
    J -->|"ElementNode, no special style"| M["No fill (transparent)"]

    L --> N["Recurse into box.Children"]
    K --> N
    M --> N

    N --> O["Draw chrome UI on top\n(always rendered last)"]
    O --> P["Dark toolbar bar (y=0 to 45)"]
    P --> Q["Blue accent line (y=45, h=2)"]
    Q --> R["URL bar rect (grey, y=8)"]
    R --> S["DebugPrintAt URL text in bar"]
    S --> T{"history not empty?"}
    T -->|"yes"| U["DebugPrintAt '← Back Esc'"]
    T -->|"no"| V["Done"]

    style A fill:#2e2e1a,color:#fff
    style O fill:#1a1a2e,color:#fff
```

### Paint colour logic

| Condition | Visual output |
|---|---|
| `TextNode` | Plain text drawn with `ebitenutil.DebugPrintAt` |
| `ElementNode` + `color: red` style | Semi-transparent red `FillRect` highlight |
| Box has `LinkURL` | 1 px blue underline drawn at bottom of box |
| Toolbar | Dark grey rect + blue accent bar |
| URL input | Darker grey rect with typed URL text |

---

## 11. User Interaction Loop

`BrowserApp.Update()` is called every frame by Ebitengine (~60 FPS). It handles all user input.

```mermaid
stateDiagram-v2
    [*] --> Idle : App starts, initial URL loaded

    Idle --> TypingURL : User presses a key
    TypingURL --> TypingURL : More keys pressed\n(rawInput grows)
    TypingURL --> TypingURL : Backspace pressed\n(rawInput shrinks)
    TypingURL --> Navigating : Enter pressed

    Idle --> HitTest : Left mouse button clicked
    HitTest --> Navigating : Link found at cursor coords
    HitTest --> Idle : No link found

    Idle --> NavigatingBack : Esc pressed AND history not empty
    NavigatingBack --> Navigating : Pop previous URL from history stack

    Navigating --> FetchHTML : network.Fetch(url)
    FetchHTML --> ParseDOM : parser.New(raw).Parse()
    ParseDOM --> RunJS : javascript runtime executes scripts
    RunJS --> ApplyCSS : css.StyleSheet applied
    ApplyCSS --> BuildLayout : buildLayoutTree + Layout()
    BuildLayout --> Idle : layoutRoot updated, next Draw() shows result

    Navigating --> Idle : Fetch error (logged, page unchanged)
```

### URL Resolution

When a link is clicked, `resolveURL(currentURL, target)` determines the absolute URL to navigate to:

```mermaid
flowchart LR
    A["resolveURL(current, target)"] --> B{"target starts\nwith 'http'?"}
    B -->|"yes"| C["Return target as-is\n(already absolute)"]
    B -->|"no"| D{"target starts\nwith '/'?"}
    D -->|"yes (root-relative)"| E["Extract scheme+host\nfrom current URL\nAppend target path"]
    D -->|"no (relative)"| F["Take current URL\nup to last '/'\nAppend target"]
```

---

## 12. Data Structures Cheat Sheet

This shows how data transforms at each stage boundary:

```mermaid
flowchart LR
    S1["string\n(raw URL)"]
    S2["string\n(raw HTML body)"]
    S3["*dom.Node\n(tree)"]
    S4["css.StyleSheet\n(rules + declarations)"]
    S5["*style.StyledNode\n(tree + computed props)"]
    S6["*layout.LayoutBox\n(tree + x/y/w/h)"]
    S7["Ebitengine screen\n(pixel buffer)"]

    S1 -->|"network.Fetch()"| S2
    S2 -->|"parser.Parse()"| S3
    S3 -->|"javascript.Execute()  mutates in-place"| S3
    S3 -->|"style.CreateStyledTree()"| S5
    S4 -->|"input to CreateStyledTree()"| S5
    S5 -->|"buildLayoutTree()\n+ Layout()"| S6
    S6 -->|"BrowserApp.paint()"| S7
```

---

*This document reflects the current state of the codebase as of March 2026. As new features are added (redirect handling, CSS class/ID selectors, full DOM API, etc.), this document should be updated accordingly.*
