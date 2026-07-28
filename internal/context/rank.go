package context

import (
	"sort"
	"strings"
)

// Rank orders items by relevance score in descending order, boosting items that match
// target paths or keywords in the query string.
func Rank(items []Item, query string) []Item {
	if len(items) == 0 {
		return nil
	}

	lowerQuery := strings.ToLower(strings.TrimSpace(query))
	keywords := strings.Fields(lowerQuery)

	ranked := make([]Item, len(items))
	copy(ranked, items)

	for i := range ranked {
		item := &ranked[i]
		lowerPath := strings.ToLower(item.Path)
		lowerContent := strings.ToLower(item.Content)

		// Boost for matching target keywords in path
		for _, kw := range keywords {
			if strings.Contains(lowerPath, kw) {
				item.Score += 0.3
			}
			if strings.Contains(lowerContent, kw) {
				item.Score += 0.1
			}
		}

		// Ensure score does not exceed 2.0
		if item.Score > 2.0 {
			item.Score = 2.0
		}
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	return ranked
}
