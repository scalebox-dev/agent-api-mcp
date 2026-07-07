package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/scalebox-dev/agent-api-mcp/internal/agentapi"
)

type createResponseInput struct {
	Input              any            `json:"input" jsonschema:"Agent API input payload."`
	Instructions       string         `json:"instructions,omitempty"`
	Preset             string         `json:"preset,omitempty" jsonschema:"Managed preset name such as pro-search or deep-research."`
	Model              string         `json:"model,omitempty" jsonschema:"Single model id in vendor/model form."`
	Models             []string       `json:"models,omitempty" jsonschema:"Ordered model candidates."`
	ModelRouting       string         `json:"model_routing,omitempty"`
	RoutingStrategy    string         `json:"routing_strategy,omitempty"`
	PreviousResponseID string         `json:"previous_response_id,omitempty"`
	VolumeID           string         `json:"volume_id,omitempty"`
	MaxOutputTokens    int            `json:"max_output_tokens,omitempty"`
	MaxSteps           int            `json:"max_steps,omitempty"`
	PlanModePreference string         `json:"plan_mode_preference,omitempty"`
	SubAgentPreference string         `json:"sub_agent_preference,omitempty"`
	Memory             map[string]any `json:"memory,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	PreferredSites     []string       `json:"preferred_sites,omitempty"`
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
	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_create_response", Description: "Create a non-streaming Agent API response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createResponseInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.post(ctx, "/v1/responses", map[string]any{
				"input":                in.Input,
				"instructions":         in.Instructions,
				"preset":               in.Preset,
				"model":                in.Model,
				"models":               in.Models,
				"model_routing":        in.ModelRouting,
				"routing_strategy":     in.RoutingStrategy,
				"previous_response_id": in.PreviousResponseID,
				"volume_id":            in.VolumeID,
				"max_output_tokens":    in.MaxOutputTokens,
				"max_steps":            in.MaxSteps,
				"plan_mode_preference": in.PlanModePreference,
				"sub_agent_preference": in.SubAgentPreference,
				"memory":               in.Memory,
				"metadata":             in.Metadata,
				"preferred_sites":      in.PreferredSites,
				"stream":               false,
			})
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_responses", Description: "List stored responses visible to the credential."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listResponsesInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := a.get(ctx, "/v1/responses", agentapi.Query(map[string]any{
				"limit":             in.Limit,
				"page_token":        in.PageToken,
				"safety_identifier": in.SafetyIdentifier,
				"user_id":           in.UserID,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_response", Description: "Retrieve a persisted Agent API response by id."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in getResponseInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			endpoint := agentapi.Join("v1", "responses", agentapi.Segment(in.ResponseID))
			out, err := a.get(ctx, endpoint, agentapi.Query(map[string]any{"safety_identifier": in.SafetyIdentifier}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_cancel_response", Description: "Best-effort cancellation for an in-flight response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.post(ctx, agentapi.Join("v1", "responses", agentapi.Segment(in.ResponseID), "cancel"), map[string]any{})
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_response_events", Description: "List response audit/timeline events."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listResponseEventsInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.get(ctx, agentapi.Join("v1", "responses", agentapi.Segment(in.ResponseID), "events"), agentapi.Query(map[string]any{
				"after_sequence": in.AfterSequence,
				"view":           in.View,
			}))
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_child_responses", Description: "List delegated sub-agent runs for a parent response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.get(ctx, agentapi.Join("v1", "responses", agentapi.Segment(in.ResponseID), "children"), nil)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_response_volume", Description: "Resolve the durable agent volume associated with a response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, map[string]any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.get(ctx, agentapi.Join("v1", "responses", agentapi.Segment(in.ResponseID), "volume"), nil)
			return JSONText(out), out, err
		})
}
