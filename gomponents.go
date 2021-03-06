// Package gomponents provides declarative view components in Go, that can render to HTML.
// The primary interface is a Node, which has a single function Render, which should render
// the Node to a string. Furthermore, NodeFunc is a function which implements the Node interface
// by calling itself on Render.
// All DOM elements and attributes can be created by using the El and Attr functions.
// The package also provides a lot of convenience functions for creating elements and attributes
// with the most commonly used parameters. If they don't suffice, a fallback to El and Attr is always possible.
package gomponents

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

// Node is a DOM node that can Render itself to a string representation.
type Node interface {
	Render() string
}

// Placer can be implemented to tell Render functions where to place the string representation of a Node
// in the parent element.
type Placer interface {
	Place() Placement
}

// Placement is used with the Placer interface.
type Placement int

const (
	Outside = Placement(iota)
	Inside
)

// NodeFunc is render function that is also a Node.
type NodeFunc func() string

func (n NodeFunc) Render() string {
	return n()
}

func (n NodeFunc) Place() Placement {
	return Outside
}

// String satisfies fmt.Stringer.
func (n NodeFunc) String() string {
	return n.Render()
}

// El creates an element DOM Node with a name and child Nodes.
// Use this if no convenience creator exists.
func El(name string, children ...Node) NodeFunc {
	return func() string {
		var b, inside, outside strings.Builder

		b.WriteString("<")
		b.WriteString(name)

		if len(children) == 0 {
			b.WriteString(" />")
			return b.String()
		}

		for _, c := range children {
			renderChild(c, &inside, &outside)
		}

		b.WriteString(inside.String())

		if outside.Len() == 0 {
			b.WriteString(" />")
			return b.String()
		}

		b.WriteString(">")
		b.WriteString(outside.String())
		b.WriteString("</")
		b.WriteString(name)
		b.WriteString(">")
		return b.String()
	}
}

func renderChild(c Node, inside, outside *strings.Builder) {
	if g, ok := c.(group); ok {
		for _, groupC := range g.children {
			renderChild(groupC, inside, outside)
		}
		return
	}
	if p, ok := c.(Placer); ok {
		switch p.Place() {
		case Inside:
			inside.WriteString(c.Render())
		case Outside:
			outside.WriteString(c.Render())
		}
		return
	}
	// If c doesn't implement Placer, default to outside
	outside.WriteString(c.Render())
}

// Attr creates an attr DOM Node.
// If one parameter is passed, it's a name-only attribute (like "required").
// If two parameters are passed, it's a name-value attribute (like `class="header"`).
// More parameter counts make Attr panic.
// Use this if no convenience creator exists.
func Attr(name string, value ...string) Node {
	switch len(value) {
	case 0:
		return attr{name: name}
	case 1:
		return attr{name: name, value: &value[0]}
	default:
		panic("attribute must be just name or name and value pair")
	}
}

type attr struct {
	name  string
	value *string
}

func (a attr) Render() string {
	if a.value == nil {
		return fmt.Sprintf(" %v", a.name)
	}
	return fmt.Sprintf(` %v="%v"`, a.name, *a.value)
}

func (a attr) Place() Placement {
	return Inside
}

// String satisfies fmt.Stringer.
func (a attr) String() string {
	return a.Render()
}

// Text creates a text DOM Node that Renders the escaped string t.
func Text(t string) NodeFunc {
	return func() string {
		return template.HTMLEscapeString(t)
	}
}

// Textf creates a text DOM Node that Renders the interpolated and escaped string t.
func Textf(format string, a ...interface{}) NodeFunc {
	return func() string {
		return template.HTMLEscapeString(fmt.Sprintf(format, a...))
	}
}

// Raw creates a raw Node that just Renders the unescaped string t.
func Raw(t string) NodeFunc {
	return func() string {
		return t
	}
}

// Write to the given io.Writer, returning any error.
func Write(w io.Writer, n Node) error {
	_, err := w.Write([]byte(n.Render()))
	return err
}

type group struct {
	children []Node
}

func (g group) Render() string {
	panic("cannot render group")
}

// Group multiple Nodes into one Node. Useful for concatenation of Nodes in variadic functions.
// The resulting Node cannot Render directly, trying it will panic.
// Render must happen through a parent element created with El or a helper.
func Group(children []Node) Node {
	return group{children: children}
}
