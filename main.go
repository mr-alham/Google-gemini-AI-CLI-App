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

type Config struct {
	GEMINI_API_KEY     string `json:"GEMINI_API_KEY"`
	GEMINI_MODEL       string `json:"GEMINI_MODEL"`
	SYSTEM_INSTRUCTION string `json:"SYSTEM_INSTRUCTION"`
	GENERATION_CONFIG  struct {
		Temperature        float32 `json:"temperature"`
		Top_p              float32 `json:"top_p"`
		Top_k              int32   `json:"top_k"`
		Max_output_tokens  *int32  `json:"max_output_tokens"`
		Response_mime_type string  `json:"response_mime_type"`
	} `json:"GENERATION_CONFIG"`

	// you can get safety settings information from,
	// https://github.com/google/generative-ai-go/blob/v0.13.0/genai/generativelanguagepb_veneer.gen.go#L817
	SAFETY_SETTINGS []struct {
		Threshold string `json:"threshold"`
	} `json:"SAFETY_SETTINGS"`
}

func main() {
	// if you intend to use a different file for json specify it here
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

	ctx := context.Background()
	client, err := createClient(ctx, config.GEMINI_API_KEY)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(config.GEMINI_MODEL)
	model.SetTemperature(config.GENERATION_CONFIG.Temperature)
	model.SetTopP(config.GENERATION_CONFIG.Top_p)
	model.SetTopK(config.GENERATION_CONFIG.Top_k)
	model.ResponseMIMEType = config.GENERATION_CONFIG.Response_mime_type
	model.MaxOutputTokens = config.GENERATION_CONFIG.Max_output_tokens
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(config.SYSTEM_INSTRUCTION)},
	}

	var thresholds = []uint8{2, 2, 2, 2}
	for index, t := range config.SAFETY_SETTINGS {
		switch {
		case t.Threshold == "HarmBlockUnspecified" || t.Threshold == "HARM_BLOCK_THRESHOLD_UNSPECIFIED":
			thresholds[index] = 0
		case t.Threshold == "HarmBlockLowAndAbove" || t.Threshold == "BLOCK_LOW_AND_ABOVE":
			thresholds[index] = 1
		case t.Threshold == "HarmBlockMediumAndAbove" || t.Threshold == "BLOCK_MEDIUM_AND_ABOVE":
			thresholds[index] = 2
		case t.Threshold == "HarmBlockOnlyHigh" || t.Threshold == "BLOCK_ONLY_HIGH":
			thresholds[index] = 3
		case t.Threshold == "HarmBlockNone" || t.Threshold == "BLOCK_NONE":
			thresholds[index] = 4
		}
	}

	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockThreshold(thresholds[0]),
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockThreshold(thresholds[1]),
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockThreshold(thresholds[2]),
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockThreshold(thresholds[3]),
		},
	}

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

	cs := model.StartChat()
	cs.History = []*genai.Content{}

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
	fmt.Println("****************************************************************")
}
