package entropy_test

import (
	"math"
	"testing"

	"github.com/ezerfernandes/mdcode/internal/entropy"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantTitle  string
		wantItems  int
	}{
		{
			name:      "single heading",
			input:     "# Roadmap\n",
			wantTitle: "Roadmap",
			wantItems: 1,
		},
		{
			name:      "headings and list items",
			input:     "# Roadmap\n## Phase 1\n- Task A\n- Task B\n## Phase 2\n- Task C\n",
			wantTitle: "Roadmap",
			wantItems: 6,
		},
		{
			name:      "no headings",
			input:     "Just some text\n",
			wantTitle: "",
			wantItems: 0,
		},
		{
			name:      "nested lists",
			input:     "# Plan\n- Item 1\n  - Sub item\n",
			wantTitle: "Plan",
			wantItems: 3,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := entropy.Parse([]byte(tc.input))
			if r.Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", r.Title, tc.wantTitle)
			}

			if len(r.Items) != tc.wantItems {
				t.Errorf("Items count = %d, want %d", len(r.Items), tc.wantItems)
			}
		})
	}
}

func TestComputeMetrics(t *testing.T) {
	t.Parallel()

	src := []byte("# Roadmap\n## Phase 1\n- Task A\n- Task B\n## Phase 2\n- Task C\n")
	r := entropy.Parse(src)
	m := entropy.ComputeMetrics(r)

	if m.SectionCount != 3 {
		t.Errorf("SectionCount = %d, want 3", m.SectionCount)
	}

	if m.ItemCount != 3 {
		t.Errorf("ItemCount = %d, want 3", m.ItemCount)
	}

	if m.TotalWords != 11 {
		t.Errorf("TotalWords = %d, want 11", m.TotalWords)
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		oldSrc       string
		newSrc       string
		wantAdded    int
		wantRemoved  int
		wantChanged  int
		wantEntropy  float64
	}{
		{
			name:        "no changes",
			oldSrc:      "# Roadmap\n- Task A\n",
			newSrc:      "# Roadmap\n- Task A\n",
			wantAdded:   0,
			wantRemoved: 0,
			wantChanged: 0,
			wantEntropy: 0.0,
		},
		{
			name:        "added items",
			oldSrc:      "# Roadmap\n- Task A\n",
			newSrc:      "# Roadmap\n- Task A\n- Task B\n- Task C\n",
			wantAdded:   2,
			wantRemoved: 0,
			wantChanged: 0,
			wantEntropy: 1.0,
		},
		{
			name:        "removed items",
			oldSrc:      "# Roadmap\n- Task A\n- Task B\n",
			newSrc:      "# Roadmap\n",
			wantAdded:   0,
			wantRemoved: 2,
			wantChanged: 0,
			wantEntropy: 2.0 / 3.0,
		},
		{
			name:        "complete rewrite",
			oldSrc:      "# Old\n- A\n- B\n",
			newSrc:      "# New\n- X\n- Y\n- Z\n",
			wantAdded:   4,
			wantRemoved: 3,
			wantChanged: 0,
			wantEntropy: 1.0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			oldR := entropy.Parse([]byte(tc.oldSrc))
			newR := entropy.Parse([]byte(tc.newSrc))
			d := entropy.Compare(oldR, newR)

			if len(d.Added) != tc.wantAdded {
				t.Errorf("Added = %d (%v), want %d", len(d.Added), d.Added, tc.wantAdded)
			}

			if len(d.Removed) != tc.wantRemoved {
				t.Errorf("Removed = %d (%v), want %d", len(d.Removed), d.Removed, tc.wantRemoved)
			}

			if len(d.Changed) != tc.wantChanged {
				t.Errorf("Changed = %d (%v), want %d", len(d.Changed), d.Changed, tc.wantChanged)
			}

			if math.Abs(d.EntropyScore-tc.wantEntropy) > 0.01 {
				t.Errorf("EntropyScore = %f, want %f", d.EntropyScore, tc.wantEntropy)
			}
		})
	}
}
