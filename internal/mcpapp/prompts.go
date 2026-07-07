package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (a *App) registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "research_with_agent",
		Description: "Plan and run a research task through Agent API.",
		Arguments: []*mcp.PromptArgument{
			{Name: "topic", Description: "Research topic.", Required: true},
			{Name: "depth", Description: "fast, standard, or deep.", Required: false},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		topic := req.Params.Arguments["topic"]
		depth := req.Params.Arguments["depth"]
		if depth == "" {
			depth = "standard"
		}
		return promptResult("Use Agent API to research this topic: " + topic + "\n\nChoose an appropriate preset for depth=" + depth + ", create a response, then inspect events and artifacts if needed."), nil
	})

	server.AddPrompt(&mcp.Prompt{
		Name:        "continue_agent_thread",
		Description: "Continue an existing Agent API response thread.",
		Arguments: []*mcp.PromptArgument{
			{Name: "previous_response_id", Description: "Previous Agent API response id.", Required: true},
			{Name: "instruction", Description: "Continuation instruction.", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return promptResult("Continue Agent API thread " + req.Params.Arguments["previous_response_id"] + " with this instruction:\n\n" + req.Params.Arguments["instruction"] + "\n\nUse previous_response_id when creating the next response."), nil
	})

	server.AddPrompt(&mcp.Prompt{
		Name:        "debug_agent_response",
		Description: "Inspect a response, timeline events, child runs, and associated volume.",
		Arguments: []*mcp.PromptArgument{
			{Name: "response_id", Description: "Agent API response id.", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return promptResult("Debug Agent API response " + req.Params.Arguments["response_id"] + ". Retrieve the response, list response events, list child responses, and resolve the response volume if present. Summarize status, errors, usage, tool activity, and likely next actions."), nil
	})
}

func promptResult(text string) *mcp.GetPromptResult {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{{
			Role:    "user",
			Content: &mcp.TextContent{Text: text},
		}},
	}
}
