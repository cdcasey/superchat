package main

import (
	"fmt"
	"log"
	"os"

	"github.com/awesome-gocui/gocui"
	"github.com/joho/godotenv"

	//"github.com/ledongthuc/pdf"
	openai "github.com/sashabaranov/go-openai"
)

var (
	openaiClient    *openai.Client
	geminiClient    *openai.Client
	systemPrompt    string
	evaluatorPrompt string
	name            = "Chris Casey"
	history         []ChatMessage
	g               *gocui.Gui
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	openaiClient = openai.NewClient(os.Getenv("OPENAPI_API_KEY"))

	geminiConfig := openai.DefaultConfig(os.Getenv("GOOGLE_API_KEY"))
	geminiConfig.BaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
	geminiClient = openai.NewClientWithConfig(geminiConfig)

	linkedinText, err := readPDF("./me/Profile.pdf")
	if err != nil {
		log.Fatalf("Error reading PDF: %v", err)
	}

	summaryBytes, err := os.ReadFile("./me/coverletter.txt")
	if err != nil {
		log.Fatalf("Error reading cover letter: %v", err)
	}
	summary := string(summaryBytes)

	systemPrompt = buildSystemPrompt(name, summary, linkedinText)
	evaluatorPrompt = buildEvaluatorPrompt(name, summary, linkedinText)

	var guiErr error
	g, guiErr = gocui.NewGui(gocui.OutputNormal, true)
	if guiErr != nil {
		log.Fatalf("Failed to create GUI: %v", guiErr)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Fatalf("Failed to set keybindings: %v", err)
	}

	addMessageToChat("system", fmt.Sprintf("Welcome! Chat with %s. Press Ctrl+C to quit.", name))

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalf("Main loop error: %v", err)
	}

}
