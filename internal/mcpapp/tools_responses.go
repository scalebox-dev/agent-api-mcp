package mcpapp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentapi "github.com/scalebox-dev/agent-api-sdk/go/agentapi"
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
	Memory             map[string]any `json:"memory,omitempty" jsonschema:"Memory options: enabled, read, write, tenant_search."`
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
		func(ctx context.Context, _ *mcp.CallToolRequest, in createResponseInput) (*mcp.CallToolResult, any, error) {
			params := agentapi.ResponseCreateParams{
				Input:              in.Input,
				Instructions:       in.Instructions,
				Preset:             in.Preset,
				Model:              in.Model,
				Models:             in.Models,
				ModelRouting:       in.ModelRouting,
				RoutingStrategy:    in.RoutingStrategy,
				PreviousResponseID: in.PreviousResponseID,
				VolumeID:           in.VolumeID,
				MaxOutputTokens:    in.MaxOutputTokens,
				MaxSteps:           in.MaxSteps,
				PlanModePreference: in.PlanModePreference,
				SubAgentPreference: in.SubAgentPreference,
				Metadata:           agentapi.Metadata(in.Metadata),
				PreferredSites:     in.PreferredSites,
			}
			if len(in.Memory) > 0 {
				memory, err := convertJSON[agentapi.MemoryOptions](in.Memory)
				if err != nil {
					return nil, nil, err
				}
				params.Memory = &memory
			}
			out, err := a.Client.Responses.Create(ctx, params)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_responses", Description: "List stored responses visible to the credential."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listResponsesInput) (*mcp.CallToolResult, any, error) {
			out, err := a.Client.Responses.List(ctx, agentapi.ListResponsesParams{
				Limit:            in.Limit,
				PageToken:        in.PageToken,
				SafetyIdentifier: in.SafetyIdentifier,
				UserID:           in.UserID,
			})
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_response", Description: "Retrieve a persisted Agent API response by id."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in getResponseInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.RetrieveWithParams(ctx, in.ResponseID, agentapi.RetrieveResponseParams{
				SafetyIdentifier: in.SafetyIdentifier,
			})
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_cancel_response", Description: "Best-effort cancellation for an in-flight response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.Cancel(ctx, in.ResponseID)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_response_events", Description: "List response audit/timeline events."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listResponseEventsInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.ListEvents(ctx, in.ResponseID, agentapi.ListEventsParams{
				AfterSequence: int64(in.AfterSequence),
				View:          in.View,
			})
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_list_child_responses", Description: "List delegated sub-agent runs for a parent response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.ListChildren(ctx, in.ResponseID)
			return JSONText(out), out, err
		})

	mcp.AddTool(server, &mcp.Tool{Name: "agent_api_get_response_volume", Description: "Resolve the durable agent volume associated with a response."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in responseIDInput) (*mcp.CallToolResult, any, error) {
			if err := require(in.ResponseID, "response_id"); err != nil {
				return nil, nil, err
			}
			out, err := a.Client.Responses.RetrieveVolume(ctx, in.ResponseID)
			return JSONText(out), out, err
		})
}
