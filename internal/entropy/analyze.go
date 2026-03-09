// Package entropy detects roadmap scope creep and drift in markdown documents.
package entropy

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// RoadmapItem represents a heading or list item in a roadmap.
type RoadmapItem struct {
	Title     string
	Depth     int
	IsHeading bool
	Children  []RoadmapItem
	LineStart int
	LineEnd   int
}

// Roadmap represents a parsed markdown roadmap structure.
type Roadmap struct {
	Title string
	Items []RoadmapItem
}

// Metrics holds summary statistics about a roadmap.
type Metrics struct {
	SectionCount int
	ItemCount    int
	MaxDepth     int
	TotalWords   int
}

// Parse extracts a roadmap structure from markdown source using goldmark AST.
func Parse(src []byte) *Roadmap {
	parser := goldmark.DefaultParser()
	reader := text.NewReader(src)
	doc := parser.Parse(reader)

	roadmap := &Roadmap{}
	var items []RoadmapItem

	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Heading:
			title := extractText(n, src)
			item := RoadmapItem{
				Title:     title,
				Depth:     n.Level,
				IsHeading: true,
				LineStart: lineNumber(n, src),
				LineEnd:   lineEnd(n, src),
			}

			if roadmap.Title == "" && n.Level == 1 {
				roadmap.Title = title
			}

			items = append(items, item)

		case *ast.ListItem:
			title := extractText(n, src)
			item := RoadmapItem{
				Title:     title,
				Depth:     headingDepthAt(items) + listDepth(n),
				LineStart: lineNumber(n, src),
				LineEnd:   lineEnd(n, src),
			}

			items = append(items, item)
		}

		return ast.WalkContinue, nil
	})

	roadmap.Items = items

	return roadmap
}

// ComputeMetrics calculates summary statistics for a roadmap.
func ComputeMetrics(r *Roadmap) Metrics {
	m := Metrics{}

	for _, item := range r.Items {
		if item.IsHeading {
			m.SectionCount++
		} else {
			m.ItemCount++
		}

		if item.Depth > m.MaxDepth {
			m.MaxDepth = item.Depth
		}

		m.TotalWords += countWords(item.Title)
	}

	return m
}

func extractText(node ast.Node, src []byte) string {
	var buf strings.Builder

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			buf.Write(t.Segment.Value(src))
		} else {
			// Recurse for inline elements
			buf.WriteString(extractText(child, src))
		}
	}

	return strings.TrimSpace(buf.String())
}

func lineNumber(node ast.Node, _ []byte) int {
	if node.Lines().Len() > 0 {
		// Use segment position for approximate line counting
		return node.Lines().At(0).Start
	}

	return 0
}

func lineEnd(node ast.Node, _ []byte) int {
	if node.Lines().Len() > 0 {
		last := node.Lines().At(node.Lines().Len() - 1)
		return last.Stop
	}

	return 0
}

func headingDepthAt(items []RoadmapItem) int {
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].IsHeading {
			return items[i].Depth
		}
	}

	return 0
}

func listDepth(node ast.Node) int {
	depth := 0

	for p := node.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*ast.List); ok {
			depth++
		}
	}

	return depth
}

func countWords(s string) int {
	return len(strings.Fields(s))
}
