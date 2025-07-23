package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sashabaranov/go-openai"
)

/*
getFixFromOpenAIWithMeta queries OpenAI's GPT-4 API using function calling to suggest a fix for a failed shell command.

Arguments:
- command: The shell command that was attempted.
- errorMsg: The error message received from running the command.
- meta: Optional additional context or notes to help OpenAI provide a better fix.

Returns:
- fix: A shell command string suggested by OpenAI to fix the error.
- explanation: A human-readable explanation of the fix.

This function constructs a prompt including the command, error, OS, and meta note, sends it to OpenAI using the
function calling (tool use) interface, and expects a structured JSON response with 'fix' and 'explanation' fields.
The use of OpenAI's ToolChoiceFunction ensures reliable and valid JSON output.
*/

func getFixFromOpenAIWithMeta(command, errorMsg, meta string) (string, string) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " 🧠 Thinking for a fix..."
	s.Start()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", "Missing OPENAI_API_KEY"
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	userOs := runtime.GOOS
	userContent := fmt.Sprintf(
		"I ran this command: %s\nIt failed with this error: %s\nI am on this OS: %s.",
		command, errorMsg, userOs,
	)
	if meta != "" {
		userContent += fmt.Sprintf("\nNote: %s", meta)
	}

	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "suggest_fix",
				Description: "Suggest a terminal command to fix a given error",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"fix": map[string]string{
							"type":        "string",
							"description": "Strictly just the terminal command to fix the issue",
						},
						"explanation": map[string]string{
							"type":        "string",
							"description": "Explanation of why this fix works",
						},
					},
					"required": []string{"fix", "explanation"},
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are a terminal assistant that always suggests shell command fixes."},
			{Role: openai.ChatMessageRoleUser, Content: userContent},
		},
		Tools: tools,
		ToolChoice: openai.ToolChoice(openai.ToolChoice{
			Type: openai.ToolTypeFunction,
			Function: openai.ToolFunction{
				Name: "suggest_fix",
			},
		}),
	})
	if err != nil {
		s.Stop()
		return "", fmt.Sprintf("Failed to contact OpenAI: %v", err)
	}

	toolCall := resp.Choices[0].Message.ToolCalls[0]
	var result struct {
		Fix         string `json:"fix"`
		Explanation string `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &result); err != nil {
		s.Stop()
		return "", fmt.Sprintf("Invalid OpenAI JSON format: %v", err)
	}

	s.Stop()
	return result.Fix, result.Explanation
}
