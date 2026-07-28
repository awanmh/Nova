package context

// Compress prunes or truncates lower-ranked items to ensure the total token count
// fits within maxTokens budget.
func Compress(items []Item, maxTokens int) *Bundle {
	if maxTokens <= 0 {
		maxTokens = 8192 // fallback default
	}

	bundle := &Bundle{
		Items:       make([]Item, 0, len(items)),
		TotalTokens: 0,
	}

	for _, item := range items {
		if bundle.TotalTokens+item.TokenCnt <= maxTokens {
			bundle.Items = append(bundle.Items, item)
			bundle.TotalTokens += item.TokenCnt
			continue
		}

		// Check if we can partially include the item
		remaining := maxTokens - bundle.TotalTokens
		if remaining >= 50 {
			// Approximate 4 chars per token
			maxChars := (remaining - 15) * 4
			if maxChars > len(item.Content) {
				maxChars = len(item.Content)
			}
			truncatedContent := item.Content[:maxChars] + "\n... [TRUNCATED TO FIT TOKEN BUDGET]"
			newTokens := EstimateTokens(truncatedContent)
			if bundle.TotalTokens+newTokens <= maxTokens {
				item.Content = truncatedContent
				item.TokenCnt = newTokens
				bundle.Items = append(bundle.Items, item)
				bundle.TotalTokens += newTokens
			}
		}
		// Budget full, stop adding lower-ranked items
		break
	}

	return bundle
}
