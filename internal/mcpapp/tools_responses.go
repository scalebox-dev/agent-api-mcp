package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
)

type createResponseInput struct {
	Input              any                             `json:"input" jsonschema:"Agent API input payload. Usually a string, message list, or structured continuation item."`
	Instructions       string                          `json:"instructions,omitempty" jsonschema:"Additional system/developer guidance for this response."`
	LanguagePreference string                          `json:"language_preference,omitempty" jsonschema:"Preferred response language, such as en-US or zh-CN."`
	CallerContext      *agentapi.CallerContext         `json:"caller_context,omitempty" jsonschema:"Optional caller locale, timezone, locality, and extra request context."`
	Preset             string                          `json:"preset,omitempty" jsonschema:"Managed preset name such as pro-search or deep-research."`
	Model              string                          `json:"model,omitempty" jsonschema:"Single model id in vendor/model form."`
	Models             []string                        `json:"models,omitempty" jsonschema:"Ordered model candidates."`
	ModelRouting       string                          `json:"model_routing,omitempty" jsonschema:"Model routing policy when multiple models are available."`
	RoutingStrategy    string                          `json:"routing_strategy,omitempty" jsonschema:"Routing strategy for model/preset execution."`
	PreviousResponseID string                          `json:"previous_response_id,omitempty" jsonschema:"Previous response id when continuing an Agent API thread."`
	VolumeID           string                          `json:"volume_id,omitempty" jsonschema:"Durable volume id to attach to this response."`
	MaxOutputTokens    int                             `json:"max_output_tokens,omitempty" jsonschema:"Upper bound for generated output tokens."`
	MaxSteps           int                             `json:"max_steps,omitempty" jsonschema:"Upper bound for agent/tool execution steps."`
	Reasoning          *agentapi.ReasoningConfig       `json:"reasoning,omitempty" jsonschema:"Reasoning controls such as effort."`
	ResponseFormat     *agentapi.ResponseFormat        `json:"response_format,omitempty" jsonschema:"Structured output format request, including json schema when needed."`
	Tools              []agentapi.Tool                 `json:"tools,omitempty" jsonschema:"Inline tool definitions available to the response."`
	ToolChoice         any                             `json:"tool_choice,omitempty" jsonschema:"Tool choice policy or explicit tool selection."`
	ParallelToolCalls  *bool                           `json:"parallel_tool_calls,omitempty" jsonschema:"Whether compatible tool calls may run in parallel."`
	Store              *bool                           `json:"store,omitempty" jsonschema:"Whether to persist the response when supported by the credential and policy."`
	PlanModePreference string                          `json:"plan_mode_preference,omitempty" jsonschema:"Preference for plan-mode behavior."`
	SubAgentPreference string                          `json:"sub_agent_preference,omitempty" jsonschema:"Preference for delegated sub-agent behavior."`
	Memory             *agentapi.MemoryOptions         `json:"memory,omitempty" jsonschema:"Memory options: enabled, read, write, tenant_search."`
	Metadata           map[string]any                  `json:"metadata,omitempty" jsonschema:"Caller-supplied metadata for tracing and application state."`
	PreferredSites     []string                        `json:"preferred_sites,omitempty" jsonschema:"Preferred sites/domains for web or retrieval-capable presets."`
	Skills             []agentapi.SkillReference       `json:"skills,omitempty" jsonschema:"Workspace skills to make available to the response."`
	LocalSkills        []agentapi.LocalSkillDescriptor `json:"local_skills,omitempty" jsonschema:"Local skill descriptors available to skill tooling."`
	SkillTool          *agentapi.SkillToolOptions      `json:"skill_tool,omitempty" jsonschema:"Controls for skill tool availability and tenant search."`
	PromptCacheKey     string                          `json:"prompt_cache_key,omitempty" jsonschema:"Stable cache key for prompt caching when supported."`
	SafetyIdentifier   string                          `json:"safety_identifier,omitempty" jsonschema:"Application/user safety partition identifier."`
	User               string                          `json:"user,omitempty" jsonschema:"End-user identifier for attribution and policy."`
}

type listResponsesInput struct {
	Limit            int    `json:"limit,omitempty"`
	PageToken        string `json:"page_token,omitempty"`
	SafetyIdentifier string `json:"safety_identifier,omitempty"`
	UserID           string `json:"user_id,omitempty"`
}

type responseIDInput struct {
	ResponseID string `json:"response_id" jsonschema:"Agent API response id."`
}

type getResponseInput struct {
	ResponseID       string `json:"response_id"`
	SafetyIdentifier string `json:"safety_identifier,omitempty"`
}

type listResponseEventsInput struct {
	ResponseID    string `json:"response_id"`
	AfterSequence int    `json:"after_sequence,omitempty"`
	View          string `json:"view,omitempty" jsonschema:"timeline or full."`
}

func (a *App) registerResponseTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "agent_api_create_response",
		Title:       "Create Agent Response",
		Description: "Create a non-streaming Agent API response. This is the primary write path for running an agent task, continuing a thread, attaching volumes, selecting models or presets, enabling memory, and providing tools or skills.",
		Annotations: mutatingTool(false, true),
	},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createResponseInput) (*mcp.CallToolResult, any, error) {
			params := agentapi.ResponseCreateParams{
				Input:              in.Input,
				Instructions:       in.Instructions,
				LanguagePreference: in.LanguagePreference,
				CallerContext:      in.CallerContext,
				Preset:             in.Preset,
				Model:              in.Model,
				Models:             in.Models,
				ModelRouting:       in.ModelRouting,
				RoutingStrategy:    in.RoutingStrategy,
				PreviousResponseID: in.PreviousResponseID,
				VolumeID:           in.VolumeID,
				MaxOutputTokens:    in.MaxOutputTokens,
				MaxSteps:           in.MaxSteps,
				Reasoning:          in.Reasoning,
				ResponseFormat:     in.ResponseFormat,
				Tools:              in.Tools,
				ToolChoice:         in.ToolChoice,
				ParallelToolCalls:  in.ParallelToolCalls,
				Store:              in.Store,
				PlanModePreference: in.PlanModePreference,
				SubAgentPreference: in.SubAgentPreference,
				Memory:             in.Memory,
				Metadata:           agentapi.Metadata(in.Metadata),
				PreferredSites:     in.PreferredSites,
				Skills:             in.Skills,
				LocalSkills:        in.LocalSkills,
				SkillTool:          in.SkillTool,
				PromptCacheKey:     in.PromptCacheKey,
				SafetyIdentifier:   in.SafetyIdentifier,
				User:               in.User,
			}
			out, err := a.Client.Responses.Create(ctx, params)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_responses", Title: "List Agent Responses", Description: "List stored responses visible to the credential.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listResponsesInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Responses.List(ctx, agentapi.ListResponsesParams{
				Limit:            in.Limit,
				PageToken:        in.PageToken,
				SafetyIdentifier: in.SafetyIdentifier,
				UserID:           in.UserID,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_response", Title: "Get Agent Response", Description: "Retrieve a persisted Agent API response by id.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in getResponseInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.RetrieveWithParams(ctx, in.ResponseID, agentapi.RetrieveResponseParams{
				SafetyIdentifier: in.SafetyIdentifier,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_cancel_response", Title: "Cancel Agent Response", Description: "Best-effort cancellation for an in-flight response.", Annotations: destructiveTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.Cancel(ctx, in.ResponseID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_response_events", Title: "List Response Events", Description: "List response audit/timeline events.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listResponseEventsInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.ListEvents(ctx, in.ResponseID, agentapi.ListEventsParams{
				AfterSequence: int64(in.AfterSequence),
				View:          in.View,
			})
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_child_responses", Title: "List Child Responses", Description: "List delegated sub-agent runs for a parent response.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.ListChildren(ctx, in.ResponseID)
			return ToolResult(out, err)
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_response_volume", Title: "Get Response Volume", Description: "Resolve the durable agent volume associated with a response.", Annotations: readOnlyTool(false)},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.RetrieveVolume(ctx, in.ResponseID)
			return ToolResult(out, err)
		})
}
