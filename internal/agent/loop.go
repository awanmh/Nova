package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/memory"
	"github.com/awanmh/Nova/internal/tools"
)

// Runner implements the autonomous iterative NOVA agent execution loop.
type Runner struct {
	mu          sync.RWMutex
	state       State
	provider    llm.Provider
	modelName   string
	executor    *tools.Executor
	store       memory.SessionStore
	sessionID   string
	personaName string
	systemRule  string
	maxIter     int
	cancelFunc  context.CancelFunc
}

// NewRunner creates a new agent runner configured with LLM provider, tool executor, memory store, and persona instructions.
func NewRunner(
	provider llm.Provider,
	modelName string,
	executor *tools.Executor,
	store memory.SessionStore,
	sessionID string,
	personaName string,
	systemRule string,
	maxIter int,
) *Runner {
	if maxIter <= 0 {
		maxIter = 10
	}
	return &Runner{
		state:       StateIdle,
		provider:    provider,
		modelName:   modelName,
		executor:    executor,
		store:       store,
		sessionID:   sessionID,
		personaName: personaName,
		systemRule:  systemRule,
		maxIter:     maxIter,
	}
}

// State returns the current execution state of the agent.
func (r *Runner) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *Runner) setState(s State) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = s
}

// Cancel aborts any running execution.
func (r *Runner) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancelFunc != nil {
		r.cancelFunc()
	}
	r.state = StateCancelled
}

// Run executes the autonomous reasoning and tool-calling loop for a given prompt.
func (r *Runner) Run(ctx context.Context, prompt string) error {
	runCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.cancelFunc = cancel
	r.mu.Unlock()
	defer cancel()

	r.setState(StateUnderstand)

	// Initialize session if not present
	if r.store != nil {
		_, err := r.store.GetSession(runCtx, r.sessionID)
		if err != nil {
			_, _ = r.store.CreateSession(runCtx, r.sessionID, "Agent Session "+r.sessionID, r.personaName)
		}
		_ = r.store.AppendMessage(runCtx, r.sessionID, llm.Message{
			Role:    "user",
			Content: prompt,
		})
	}

	guard := NewGuard(r.maxIter)

	for {
		if err := guard.CheckIteration(); err != nil {
			r.setState(StateFailed)
			return err
		}

		select {
		case <-runCtx.Done():
			r.setState(StateCancelled)
			return runCtx.Err()
		default:
		}

		r.setState(StatePlanning)

		messages, err := r.buildMessages(runCtx, prompt)
		if err != nil {
			r.setState(StateFailed)
			return err
		}

		resp, err := r.provider.Chat(runCtx, &llm.ChatRequest{
			Model:       r.modelName,
			Messages:    messages,
			Temperature: 0.2,
		})
		if err != nil {
			r.setState(StateFailed)
			return fmt.Errorf("LLM chat error: %w", err)
		}

		// Store assistant response in history
		if r.store != nil {
			_ = r.store.AppendMessage(runCtx, r.sessionID, resp.Message)
		}

		// If no tool calls requested -> task completed
		if len(resp.Message.ToolCalls) == 0 {
			r.setState(StateCompleted)
			return nil
		}

		// Execute tool calls
		r.setState(StateExecuting)
		for _, call := range resp.Message.ToolCalls {
			if err := guard.CheckInfiniteLoop(call.Name, call.Arguments); err != nil {
				r.setState(StateFailed)
				return err
			}

			toolResp := r.executor.ExecuteTool(runCtx, call)

			content := toolResp.Content
			if toolResp.Error != "" {
				content = fmt.Sprintf("TOOL ERROR: %s", toolResp.Error)
			}

			toolMsg := llm.Message{
				Role:    "tool",
				Content: content,
				ToolID:  call.ID,
			}
			if r.store != nil {
				_ = r.store.AppendMessage(runCtx, r.sessionID, toolMsg)
			}
		}

		r.setState(StateObserving)
	}
}

// Step performs one turn of the agent loop (for manual stepping/testing).
func (r *Runner) Step(ctx context.Context) (State, error) {
	// Simple step wrapper
	return r.State(), nil
}

func (r *Runner) buildMessages(ctx context.Context, prompt string) ([]llm.Message, error) {
	var messages []llm.Message
	if r.systemRule != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: r.systemRule,
		})
	}

	if r.store != nil {
		history, err := r.store.GetHistory(ctx, r.sessionID)
		if err == nil && len(history) > 0 {
			messages = append(messages, history...)
			return messages, nil
		}
	}

	// Fallback single user prompt if no memory store
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: prompt,
	})
	return messages, nil
}
