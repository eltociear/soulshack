package main

import (
	"encoding/json"
	"fmt"
	"log"

	ai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type IrcOpTool struct {
}

func RegisterIrcTools(registry *ToolRegistry) {
	log.Println("registering irc tools")
	registry.RegisterTool("irc_mode", &IrcOpTool{})
	registry.RegisterTool("irc_kick", &IrcKickTool{})
	registry.RegisterTool("irc_topic", &IrcTopicTool{})
}

func (t *IrcOpTool) GetTool() (ai.Tool, error) {
	return ai.Tool{
		Type: ai.ToolTypeFunction,
		Function: &ai.FunctionDefinition{
			Name:        "irc_mode",
			Description: "grants or removes irc ops to a nick",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"nick": {
						Type:        jsonschema.String,
						Description: "the irc nickname of the user to grant or revoke ops",
					},
					"op": {
						Type:        jsonschema.Boolean,
						Description: "determines if it is a grant or revoke of ops",
					},
				},
				Required: []string{"nick", "op"},
			},
		}}, nil
}

func (t *IrcOpTool) Execute(ctx ChatContext, tool ai.ToolCall) (ai.ChatCompletionMessage, error) {

	type kickRequest struct {
		Nick string `json:"nick"`
		Op   bool   `json:"op"`
	}
	var req kickRequest

	err := json.Unmarshal([]byte(tool.Function.Arguments), &req)

	if err != nil {
		return ai.ChatCompletionMessage{
			Role:       ai.ChatMessageRoleTool,
			ToolCallID: tool.ID,
			Content:    "failed to unmarshal arguments" + err.Error(),
			Name:       tool.Function.Name}, err
	}

	if !ctx.IsAdmin() {
		return ai.ChatCompletionMessage{
			Role:       ai.ChatMessageRoleTool,
			ToolCallID: tool.ID,
			Name:       tool.Function.Name,
			Content:    req.Nick + "doesn't have admin permission to perform this action."}, fmt.Errorf("unauthorized")
	}

	// set opcmd to the appropriate value
	opcmd := "-o"
	if req.Op {
		opcmd = "+o"
	}

	ctx.Client.Cmd.Mode(BotConfig.Channel, opcmd, req.Nick)

	return ai.ChatCompletionMessage{
		Role:       ai.ChatMessageRoleTool,
		Content:    "success",
		Name:       tool.Function.Name,
		ToolCallID: tool.ID,
	}, nil
}

type IrcKickTool struct {
}

func (t *IrcKickTool) GetTool() (ai.Tool, error) {
	return ai.Tool{
		Type: ai.ToolTypeFunction,
		Function: &ai.FunctionDefinition{
			Name:        "irc_kick",
			Description: "kicks a nick from the channel",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"nick": {
						Type:        jsonschema.String,
						Description: "the irc nickname to kick",
					},
					"reason": {
						Type:        jsonschema.String,
						Description: "the reason for the kick",
					},
				},
				Required: []string{"nick", "reason"},
			},
		}}, nil
}

func (t *IrcKickTool) Execute(ctx ChatContext, tool ai.ToolCall) (ai.ChatCompletionMessage, error) {
	type kickRequest struct {
		Nick   string `json:"nick"`
		Reason string `json:"reason"`
	}
	var req kickRequest
	err := json.Unmarshal([]byte(tool.Function.Arguments), &req)

	if err != nil {
		return ai.ChatCompletionMessage{
			Role:       ai.ChatMessageRoleTool,
			ToolCallID: tool.ID,
			Name:       tool.Function.Name,
			Content:    "failed to unmarshal arguments" + err.Error(),
		}, err
	}

	if !ctx.IsAdmin() {
		return ai.ChatCompletionMessage{
			Role:       ai.ChatMessageRoleTool,
			Name:       tool.Function.Name,
			ToolCallID: tool.ID,
			Content:    ctx.Event.Source.Name + "doesn't have admin permission to perform this action.",
		}, fmt.Errorf("unauthorized")
	}

	ctx.Client.Cmd.Kick(BotConfig.Channel, req.Nick, req.Reason)

	return ai.ChatCompletionMessage{
		Role:       ai.ChatMessageRoleTool,
		Content:    "success",
		ToolCallID: tool.ID,
		Name:       tool.Function.Name,
	}, nil
}

type IrcTopicTool struct {
}

func (t *IrcTopicTool) GetTool() (ai.Tool, error) {
	return ai.Tool{
		Type: ai.ToolTypeFunction,
		Function: &ai.FunctionDefinition{
			Name:        "irc_topic",
			Description: "sets the channel topic",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"topic": {
						Type:        jsonschema.String,
						Description: "the new channel topic",
					},
				},
				Required: []string{"topic"},
			},
		}}, nil
}

func (t *IrcTopicTool) Execute(ctx ChatContext, tool ai.ToolCall) (ai.ChatCompletionMessage, error) {

	type topicRequest struct {
		Topic string `json:"topic"`
	}

	var req topicRequest
	err := json.Unmarshal([]byte(tool.Function.Arguments), &req)

	if err != nil {
		return ai.ChatCompletionMessage{
			Role:       ai.ChatMessageRoleTool,
			ToolCallID: tool.ID,
			Name:       tool.Function.Name,
			Content:    "failed to unmarshal arguments" + err.Error(),
		}, err
	}

	if !ctx.IsAdmin() {
		return ai.ChatCompletionMessage{
			Role:       ai.ChatMessageRoleTool,
			ToolCallID: tool.ID,
			Name:       tool.Function.Name,
			Content:    ctx.Event.Source.Name + " has no admin permission to perform this action.",
		}, fmt.Errorf("unauthorized")
	}

	ctx.Client.Cmd.Topic(BotConfig.Channel, req.Topic)
	return ai.ChatCompletionMessage{
		Role:       ai.ChatMessageRoleTool,
		Content:    "success",
		ToolCallID: tool.ID,
		Name:       tool.Function.Name,
	}, nil
}

func checkBotOp(_ ChatContext) error {
	return nil
}
