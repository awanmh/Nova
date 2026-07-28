package context

import (
	"bufio"
	"path/filepath"
	"strings"
)

// ExtractGoSymbols scans a Go source string and extracts high-level symbol signatures
// (type, struct, interface, function headers) for lightweight context injection.
func ExtractGoSymbols(relPath, content string) []Item {
	var symbols []Item
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentBlock strings.Builder
	inBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Top-level function declarations
		if strings.HasPrefix(trimmed, "func ") && strings.HasSuffix(trimmed, "{") {
			sig := strings.TrimSuffix(trimmed, "{")
			sig = strings.TrimSpace(sig)
			symbols = append(symbols, Item{
				Type:     ItemTypeSymbol,
				Path:     filepath.ToSlash(relPath),
				Content:  sig,
				Score:    0.8,
				TokenCnt: EstimateTokens(sig),
			})
			continue
		}

		// Top-level struct or interface definitions
		if strings.HasPrefix(trimmed, "type ") && (strings.Contains(trimmed, " struct") || strings.Contains(trimmed, " interface")) {
			inBlock = true
			currentBlock.Reset()
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
			if strings.HasSuffix(trimmed, "}") {
				// single-line type def
				inBlock = false
				content := strings.TrimSpace(currentBlock.String())
				symbols = append(symbols, Item{
					Type:     ItemTypeSymbol,
					Path:     filepath.ToSlash(relPath),
					Content:  content,
					Score:    0.85,
					TokenCnt: EstimateTokens(content),
				})
			}
			continue
		}

		if inBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
			if trimmed == "}" {
				inBlock = false
				content := strings.TrimSpace(currentBlock.String())
				symbols = append(symbols, Item{
					Type:     ItemTypeSymbol,
					Path:     filepath.ToSlash(relPath),
					Content:  content,
					Score:    0.85,
					TokenCnt: EstimateTokens(content),
				})
			}
		}
	}

	return symbols
}
