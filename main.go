package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	//"github.com/ledongthuc/pdf"
	openai "github.com/sashabaranov/go-openai"
)

var (
	openaiClient *openai.Client
	geminiClient *openai.Client
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	openaiClient = openai.NewClient(os.Getenv("OPENAPI_API_KEY"))

	geminiConfig := openai.DefaultConfig(os.Getenv("GOOGLE_API_KEY"))
	geminiConfig.BaseURL = "https://generativelanguage.googleapis.com/v1beta/openai"
	geminiClient = openai.NewClientWithConfig(geminiConfig)

	linkedinText, err := readPDF("./Profile.pdf")
	if err != nil {
		log.Fatalf("Error reading PDF: %v", err)
	}

	summaryBytes, err := os.ReadFile("./coverletter.txt")
	if err != nil {
		log.Fatalf("Error reading cover letter: %v", err)
	}
	summary := string(summaryBytes)

	fmt.Println(summary)

	fmt.Println(linkedinText)
	fmt.Printf("Hithere\n%v", geminiConfig.String())
}
