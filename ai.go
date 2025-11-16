package main

import (
	"encoding/json"
	"fmt"
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
	return "nil", nil
}

// func evaluateResponse(ctx context.Context, reply, message string, history []ChatMessage) (*Evaluation, error) {

// }

// func rerunWithFeedback(ctx context.Context, reply, message string, history []ChatMessage, feedback string) (string, error) {

// }

func addMessageToChat(role, content string) {
	v, err := g.View(ViewChat)
	if err != nil {
		return
	}

	var prefix string
	switch role {
	case "user":
		prefix = "\n[YOU]"
	case "assistant":
		prefix = "\n[ASSISTANT]"
	case "system":
		prefix = "\n[SYSTEM]"
	case "error":
		prefix = "\n[ERROR]"
	}

	fmt.Fprintf(v, "%s\n%s\n", prefix, content)
}

func updateStatus(message string) {
	v, err := g.View(ViewStatus)
	if err != nil {
		return
	}
	v.Clear()
	fmt.Fprintf(v, " %s | Enter: Send | Ctrl+C: Quit", message)
}
