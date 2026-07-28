package agent

import (
	"fmt"
	"sync"
)

// Guard prevents runaway loops by capping maximum iterations and detecting repeating tool call arguments.
type Guard struct {
	mu            sync.Mutex
	maxIterations int
	iteration     int
	recentCalls   []string // format: "toolName:arguments"
}

// NewGuard creates a loop guard with the specified maximum iterations (default 10 if <= 0).
func NewGuard(maxIter int) *Guard {
	if maxIter <= 0 {
		maxIter = 10
	}
	return &Guard{
		maxIterations: maxIter,
		recentCalls:   make([]string, 0, 5),
	}
}

// CheckIteration increments the step counter and returns an error if maxIterations is exceeded.
func (g *Guard) CheckIteration() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.iteration++
	if g.iteration > g.maxIterations {
		return fmt.Errorf("loop safety abort: maximum iterations (%d) exceeded", g.maxIterations)
	}
	return nil
}

// CheckInfiniteLoop records a tool call and returns an error if the exact same tool call has repeated 3 times in a row.
func (g *Guard) CheckInfiniteLoop(toolName, argsJSON string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := fmt.Sprintf("%s:%s", toolName, argsJSON)
	g.recentCalls = append(g.recentCalls, key)

	n := len(g.recentCalls)
	if n >= 3 {
		if g.recentCalls[n-1] == key && g.recentCalls[n-2] == key && g.recentCalls[n-3] == key {
			return fmt.Errorf("loop safety abort: infinite loop detected (tool '%s' called 3 times consecutively with identical arguments)", toolName)
		}
	}
	return nil
}

// Iteration returns the current step count.
func (g *Guard) Iteration() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.iteration
}
