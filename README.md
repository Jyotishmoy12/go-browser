# Go Browser Engine

> A web browser engine built from scratch in Go.

## Overview

**Go Browser Engine** is a lightweight, browser engine implemented entirely in Go. It renders web pages by fetching raw HTML over a custom HTTP/HTTPS client, parsing it into a DOM tree, applying CSS styles, computing a layout, and finally painting the result to a native desktop window using the [Ebitengine](https://ebitengine.org/) 2D game engine.

The project is structured as a pipeline of independent, self-contained internal packages that mirror the core subsystems found in a real browser:

| Package | Responsibility |
|---|---|
| `internal/network` | Raw TCP/TLS HTTP client — fetches HTML without using `net/http` |
| `internal/parser` | HTML tokenizer and DOM tree builder |
| `internal/dom` | DOM node types (element, text) |
| `internal/css` | CSS rule and declaration representation |
| `internal/style` | Style resolution — matches CSS rules to DOM nodes |
| `internal/layout` | Box model layout engine — computes positions and dimensions |
| `internal/javascript` | JavaScript runtime (powered by [otto](https://github.com/robertkrimen/otto)) |
| `cmd/browser` | Main application — wires all subsystems together and runs the window |

### Key Features

- **Custom HTTP layer** — connects over raw TCP sockets, handles both plain HTTP and HTTPS (TLS), and decodes chunked transfer encoding manually.
- **HTML parsing** — builds a live DOM tree from raw HTML source.
- **CSS styling** — applies a stylesheet to produce a styled node tree.
- **Layout engine** — implements a basic block-level box model to position elements.
- **JavaScript execution** — runs inline `<script>` blocks against the DOM.
- **Link navigation** — click-based hit-testing resolves and follows hyperlinks.
- **Browser history** — navigate back with the `Esc` key.
- **Rendered with Ebitengine** — the final painted output is displayed in a native desktop window.

---

## Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- A desktop environment (Ebitengine opens a native window)

### Run

```bash
git clone https://github.com/jyotishmoy12/go-browser.git
cd go-browser
go run ./cmd/browser/
```

The browser opens and navigates to `http://httpbin.org/html` on startup.

### Controls

| Input | Action |
|---|---|
| Type in window | Edit the URL bar |
| `Backspace` | Delete last character in URL |
| `Enter` | Navigate to the typed URL |
| `Left Click` | Follow a hyperlink |
| `Mouse Wheel` | Scroll the page up / down |
| `Esc` | Go back to the previous page |

---

## Tech Stack

| Dependency | Purpose |
|---|---|
| [Ebitengine v2](https://ebitengine.org/) | 2D game engine used as the rendering backend |
| [otto](https://github.com/robertkrimen/otto) | Pure-Go JavaScript interpreter (ES5) |
| Go standard library only (net, crypto/tls, bufio…) | Everything else |

---

## Project Structure

```
go-browser/
├── cmd/
│   └── browser/
│       └── main.go          # Entry point — BrowserApp, navigation, paint loop
├── internal/
│   ├── network/
│   │   └── request.go       # Raw TCP/TLS HTTP fetcher
│   ├── parser/
│   │   └── html.go          # Recursive-descent HTML parser
│   ├── dom/
│   │   └── node.go          # DOM Node types (Element / Text)
│   ├── css/
│   │   ├── css.go           # StyleSheet / Rule / Declaration types
│   │   └── parser.go        # CSS text parser
│   ├── style/
│   │   └── style.go         # CSS rule matching → StyledNode tree
│   ├── layout/
│   │   └── layout.go        # Box model → LayoutBox tree (X/Y/W/H)
│   └── javascript/
│       └── runtime.go       # otto VM wrapper + console.log / document bindings
├── ARCHITECTURE.md           # Full pipeline diagrams and package internals
├── TESTING.md                # Functionality reference, test suite, and test procedures
├── README.md                 # This file
├── go.mod
└── go.sum
```

---

## Documentation

| Document | Description |
|---|---|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Detailed walkthrough of the entire request-to-pixel pipeline with Mermaid diagrams for every stage — network fetch, HTML parsing, DOM, CSS, style resolution, layout, JavaScript execution, and painting |
| [TESTING.md](./TESTING.md) | Complete functionality reference (what works / what doesn't), breakdown of every existing test, how to run them, manual testing procedures, and suggested future tests |

---

## How It Works (Pipeline Summary)

```
URL entered
    │
    ▼
[network] ── raw TCP/TLS socket ──▶ raw HTML string
    │
    ▼
[parser]  ── recursive descent ───▶ *dom.Node tree
    │
    ▼
[javascript] ── otto VM ──────────▶ DOM mutations (inline <script>)
    │
    ▼
[css]     ── hardcoded rules ─────▶ css.StyleSheet
    │
    ▼
[style]   ── selector matching ───▶ *StyledNode tree (computed props)
    │
    ▼
[layout]  ── block box model ─────▶ *LayoutBox tree (X / Y / W / H)
    │
    ▼
[Ebitengine] ── Draw() / paint() ─▶ pixels on screen
```

> For the full annotated version with sequence diagrams, state machines, and class diagrams, see [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Running Tests

```bash
# Run all unit tests
go test ./...

# Verbose output
go test -v ./...
```

> For a breakdown of what each test covers and suggestions for expanding coverage, see [TESTING.md](./TESTING.md).
