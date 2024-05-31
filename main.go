package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)


type SafetySetting struct {
	Category genai.HarmCategory `json:"category"`
	Threshold genai.HarmBlockThreshold `json:"threshold"`
}

type Config struct {
	GEMINI_API_KEY string `json:"GEMINI_API_KEY"`
	GEMINI_MODEL   string `json:"GEMINI_MODEL"`
}

func main() {
	const configFile = "keys.json"

	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Error opening config file: ", err)
	}

	defer file.Close()
	var config Config

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)


	if err != nil {
		log.Fatal("Error decoding config JSON: ", err)
	}

	apiKey := config.GEMINI_API_KEY
	generativeModel := config.GEMINI_MODEL

	ctx := context.Background()
	client, err := createClient(ctx, apiKey)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(generativeModel)
	// model.SafetySettings = config.SAFETY_SETTINGS

	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "--image" {
		err := generateTextFromImage(ctx, model)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := generateTextFromPrompt(ctx, model)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	return genai.NewClient(ctx, option.WithAPIKey(apiKey))
}

// Generate text from text-and-image input (multimodal)
func generateTextFromImage(ctx context.Context, model *genai.GenerativeModel) error {
	var pathToImage string
	var userPrompt string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Path to The Image (OR `Text Mode` to switch): ")
		scanner.Scan()
		pathToImage = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning file: ", err)
			continue
		}

		if strings.ToLower(pathToImage) == "text mode" {
			generateTextFromPrompt(ctx, model)
		}

		imgData, err := os.ReadFile(pathToImage)
		if err != nil {
			fmt.Println("Error reading image file: ", err)
			continue
		}

		fmt.Print("The Prompt: ")
		scanner.Scan()
		userPrompt = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning prompt: ", err)
			continue
		}

		prompt := []genai.Part{
			genai.ImageData("jpeg", imgData),
			genai.Text(userPrompt),
		}

		resp, err := model.GenerateContent(ctx, prompt...)
		if err != nil {
			fmt.Println("Error generating content: ", err)
			continue
		}
		printResponse(resp)
	}
}

// Generate text from text-only input
func generateTextFromPrompt(ctx context.Context, model *genai.GenerativeModel) error {
	var userPrompt string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("The Prompt(Or `Image Mode` to switch): ")
		scanner.Scan()
		userPrompt = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning prompt: ", err)
			continue

		} else if userPrompt == "" {
			fmt.Println("The Prompt is empty.")
			continue
		} else if strings.ToLower(userPrompt) == "image mode" {
			generateTextFromPrompt(ctx, model)
		}

		cs := model.StartChat()
		cs.History = []*genai.Content{}

		resp, err := cs.SendMessage(ctx, genai.Text(userPrompt))
		if err != nil {
			fmt.Println("Error sending message: ", err)
			continue
		}

		printResponse(resp)
	}
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println()
}
