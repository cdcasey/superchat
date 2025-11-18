package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// Evaluation represents the structured response from the evaluator
type Evaluation struct {
	IsAcceptable bool   `json:"is_acceptable"`
	Feedback     string `json:"feedback"`
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func buildSystemPrompt(name, summary, linkedin string) string {
	prompt := fmt.Sprintf("You are acting as %s. You are answering questions on %s's website, "+
		"particularly questions related to %s's career, background, skills and experience. "+
		"Your responsibility is to represent %s for interactions on the website as faithfully as possible. "+
		"You are given a summary of %s's background and LinkedIn profile which you can use to answer questions. "+
		"Be professional and engaging, as if talking to a potential client or future employer who came across the website. "+
		"If you don't know the answer, say so.", name, name, name, name, name)

	prompt += fmt.Sprintf("\n\n## Summary:\n%s\n\n## LinkedIn Profile:\n%s\n\n", summary, linkedin)
	prompt += fmt.Sprintf("With this context, please chat with the user, always staying in character as %s.", name)

	return prompt
}

func buildEvaluatorPrompt(name, summary, linkedin string) string {
	prompt := fmt.Sprintf("You are an evaluator that decides whether a response to a question is acceptable. "+
		"You are provided with a conversation between a User and an Agent. Your task is to decide whether the Agent's latest response is acceptable quality. "+
		"The Agent is playing the role of %s and is representing %s on their website. "+
		"The Agent has been instructed to be professional and engaging, as if talking to a potential client or future employer who came across the website. "+
		"The Agent has been provided with context on %s in the form of their summary and LinkedIn details. Here's the information:", name, name, name)

	prompt += fmt.Sprintf("\n\n## Summary:\n%s\n\n## LinkedIn Profile:\n%s\n\n", summary, linkedin)
	prompt += "With this context, please evaluate the latest response, replying with whether the response is acceptable and your feedback."

	return prompt
}

func buildEvaluatorUserPrompt(reply, message string, history []ChatMessage) string {
	historyStr, _ := json.MarshalIndent(history, "", "  ")
	prompt := fmt.Sprintf("Here's the conversation between the User and the Agent: \n\n%s\n\n", historyStr)
	prompt += fmt.Sprintf("Here's the latest message from the User: \n\n%s\n\n", message)
	prompt += fmt.Sprintf("Here's the latest response from the Agent: \n\n%s\n\n", reply)
	prompt += "Please evaluate the response, replying with whether it is acceptable and your feedback."
	return prompt
}

func processChat(message string, history []ChatMessage) (string, error) {
	ctx := context.Background()

	currentSystemPrompt := systemPrompt
	if strings.Contains(strings.ToLower(message), "patent") {
		currentSystemPrompt += "\n\nEverything in your reply needs to be in pig latin - " +
			"it is mandatory that you respond only and entirely in pig latin"
	}

	// Build messages
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: currentSystemPrompt},
	}
	for _, h := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	// Get initial response
	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("chat completion error: %w", err)
	}

	reply := resp.Choices[0].Message.Content

	// Evaluate the response
	evaluation, err := evaluateResponse(ctx, reply, message, history)
	if err != nil {
		return "", fmt.Errorf("evaluation error: %w", err)
	}

	if evaluation.IsAcceptable {
		g.Update(func(g *gocui.Gui) error {
			addMessageToChat("system", "✓ Evaluation: Passed")
			return nil
		})
		return reply, nil
	}

	g.Update(func(g *gocui.Gui) error {
		addMessageToChat("system", fmt.Sprintf("✗ Evaluation: Failed - %s", evaluation.Feedback))
		addMessageToChat("system", "Regenerating response...")
		return nil
	})

	// Rerun with feedback
	return rerunWithFeedback(ctx, reply, message, history, evaluation.Feedback)

}

func evaluateResponse(ctx context.Context, reply, message string, history []ChatMessage) (*Evaluation, error) {
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"is_acceptable": {
				Type:        jsonschema.Boolean,
				Description: "Whether or not the response is acceptable",
			},
			"feedback": {
				Type:        jsonschema.String,
				Description: "Feedback on the response quality",
			},
		},
		Required: []string{"is_acceptable", "feedback"},
	}

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: evaluatorPrompt},
		{Role: openai.ChatMessageRoleUser, Content: buildEvaluatorUserPrompt(reply, message, history)},
	}

	resp, err := geminiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gemini-2.0-flash-exp",
		Messages: messages,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "evaluation",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var evaluation Evaluation
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &evaluation); err != nil {
		return nil, fmt.Errorf("Failed to parse evaluation: %w", err)
	}

	return &evaluation, nil
}

func rerunWithFeedback(ctx context.Context, reply, message string, history []ChatMessage, feedback string) (string, error) {
	updatedSystemPrompt := systemPrompt + "\n\n## Previous answer rejected\nYou just tried to reply, but the quality control rejected your reply\n"
	updatedSystemPrompt += fmt.Sprintf("## Your attempted answer:\n%s\n\n", reply)
	updatedSystemPrompt += fmt.Sprintf("## Reason for rejection:\n%s\n\n", feedback)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: updatedSystemPrompt},
	}
	for _, h := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
