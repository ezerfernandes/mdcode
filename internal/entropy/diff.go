package entropy

// Drift holds the result of comparing two roadmap versions.
type Drift struct {
	Added        []string `json:"added"`
	Removed      []string `json:"removed"`
	Changed      []string `json:"changed"`
	EntropyScore float64  `json:"entropy_score"`
}

// Compare diffs two parsed roadmaps and returns a Drift report.
func Compare(old, new *Roadmap) Drift {
	oldMap := indexItems(old)
	newMap := indexItems(new)

	var d Drift

	for title := range newMap {
		if _, exists := oldMap[title]; !exists {
			d.Added = append(d.Added, title)
		}
	}

	for title := range oldMap {
		if _, exists := newMap[title]; !exists {
			d.Removed = append(d.Removed, title)
		}
	}

	for title, oldItem := range oldMap {
		if newItem, exists := newMap[title]; exists {
			if oldItem.Depth != newItem.Depth {
				d.Changed = append(d.Changed, title)
			}
		}
	}

	d.EntropyScore = entropyScore(old, d)

	return d
}

func indexItems(r *Roadmap) map[string]RoadmapItem {
	m := make(map[string]RoadmapItem, len(r.Items))

	for _, item := range r.Items {
		m[item.Title] = item
	}

	return m
}

func entropyScore(old *Roadmap, d Drift) float64 {
	changes := len(d.Added) + len(d.Removed) + len(d.Changed)
	if changes == 0 {
		return 0.0
	}

	base := len(old.Items)
	if base == 0 {
		if changes > 0 {
			return 1.0
		}

		return 0.0
	}

	score := float64(changes) / float64(base)
	if score > 1.0 {
		score = 1.0
	}

	return score
}
