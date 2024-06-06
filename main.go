package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/google/generative-ai-go/genai"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/atotto/clipboard"
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
	width := terminalWidth()
	titleA := "---------------🝔---------------"
	title := "Gemini-AI on Terminal"

	spaceA := strings.Repeat(" ", ((width - 31) / 2))
	spaceTitle := strings.Repeat(" ", ((width - 22) / 2))

	fmt.Println("\033[1;38;5;178m" + spaceA + titleA + "\033[0;37m")
	fmt.Println("\033[1;38;5;38m" + spaceTitle + title + "\033[0m")
	fmt.Println("\033[1;38;5;178m" + spaceA + titleA + "\033[0;37m")

	// if you intend to use a different file for json specify it here
	const configFile = "Gemini_Ai_Config/keys.json"

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
	model.MaxOutputTokens = config.GENERATION_CONFIG.Max_output_tokens
	model.ResponseMIMEType = config.GENERATION_CONFIG.Response_mime_type
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
	width := terminalWidth()

	fmt.Println("\033[2;1;93mYou are currently using Multi Mode Model")
	fmt.Println("Enter `Text Mode` instead of image path, To switch to text-to-text model\033[0;1;2;95m")
	fmt.Println(strings.Repeat("─", width-3), "\033[0;37m")

	var pathToImage string
	var userPrompt string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\033[0;1;38;5;28mPath to Image: \033[0;38;5;254m")
		scanner.Scan()
		pathToImage = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning file: ", err)
			continue
		}

		if pathToImage == "" {
			clipboardContent, err := clipboard.ReadAll()
			if err != nil {
				fmt.Println("Error Reading Clipboard: ", err)
				continue
			}
			pathToImage = clipboardContent
		}

		if strings.ToLower(pathToImage) == "text mode" {
			generateTextFromPrompt(ctx, model)
		}

		imgData, err := os.ReadFile(pathToImage)
		if err != nil {
			fmt.Println("Error reading image file: ", err)
			continue
		}

		fmt.Print("\033[0;1;38;5;28mThe Prompt: \033[0;38;5;254m")
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
	width := terminalWidth()

	fmt.Println("\033[2;1;93mYou are currently using Text-to-Text Model")
	fmt.Println("Enter `Image Mode` to switch to multi mode model\033[0;1;2;95m")
	fmt.Println(strings.Repeat("─", width-3), "\033[0;37m")

	var userPrompt string
	scanner := bufio.NewScanner(os.Stdin)

	cs := model.StartChat()
	cs.History = []*genai.Content{}

	for {
		fmt.Print("\033[0;1;38;5;28mThe Prompt: \033[0;38;5;254m")
		scanner.Scan()
		userPrompt = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning prompt: ", err)
			continue

		} else if userPrompt == "" {
			fmt.Println("The Prompt is empty.")
			continue
		} else if strings.ToLower(userPrompt) == "image mode" {
			generateTextFromImage(ctx, model)
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
	width := terminalWidth()

	const (
		defaultMargin     = 2
		defaultListIndent = 2
	)

	// DarkStyleConfig is the default dark style.
	customStylingConfig := ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr("250"),
			},
			Margin: uintPtr(defaultMargin),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
			Indent:         uintPtr(2),
			IndentToken:    stringPtr("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: defaultListIndent,
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       stringPtr("39"),
				Bold:        boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          "🞙 ",
				Suffix:          " 🞙",
				Color:           stringPtr("39"),
				BackgroundColor: stringPtr("63"),
				Bold:            boolPtr(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "🞘🞘 ",
				Suffix: " 🞘🞘",
				Color:  stringPtr("38"),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "🞗🞗🞗 ",
				Suffix: " 🞗🞗🞗",
				Color:  stringPtr("40"),
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "🞗🞗🞗🞗 ",
				Color:  stringPtr("165"),
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "🞗🞗🞗🞗🞗 ",
				Suffix: " 🞗🞗🞗🞗🞗",
				Color:  stringPtr("124"),
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "🞗🞗🞗🞗🞗🞗 ",
				Suffix: " 🞗🞗🞗🞗🞗🞗",
				Color:  stringPtr("35"),
				Bold:   boolPtr(false),
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  stringPtr("245"),
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Bold:  boolPtr(true),
			Color: stringPtr("214"),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  stringPtr("240"),
			Format: "\n----🝔----\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "🟆 ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{},
			Ticked:         "[✓] ",
			Unticked:       "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     stringPtr("5"),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: stringPtr("30"),
			Bold:  boolPtr(true),
		},
		Image: ansi.StylePrimitive{
			Color:     stringPtr("212"),
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  stringPtr("243"),
			Format: "Image: {{.text}} →",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           stringPtr("203"),
				BackgroundColor: stringPtr("236"),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr("244"),
				},
				Margin: uintPtr(defaultMargin),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: stringPtr("#C4C4C4"),
				},
				Error: ansi.StylePrimitive{
					Color:           stringPtr("#F1F1F1"),
					BackgroundColor: stringPtr("#F05B5B"),
				},
				Comment: ansi.StylePrimitive{
					Color: stringPtr("#676767"),
				},
				CommentPreproc: ansi.StylePrimitive{
					Color: stringPtr("#FF875F"),
				},
				Keyword: ansi.StylePrimitive{
					Color: stringPtr("#00AAFF"),
				},
				KeywordReserved: ansi.StylePrimitive{
					Color: stringPtr("#FF5FD2"),
				},
				KeywordNamespace: ansi.StylePrimitive{
					Color: stringPtr("#FF5F87"),
				},
				KeywordType: ansi.StylePrimitive{
					Color: stringPtr("#6E6ED8"),
				},
				Operator: ansi.StylePrimitive{
					Color: stringPtr("#EF8080"),
				},
				Punctuation: ansi.StylePrimitive{
					Color: stringPtr("#E8E8A8"),
				},
				Name: ansi.StylePrimitive{
					Color: stringPtr("#C4C4C4"),
				},
				NameBuiltin: ansi.StylePrimitive{
					Color: stringPtr("#FF8EC7"),
				},
				NameTag: ansi.StylePrimitive{
					Color: stringPtr("#B083EA"),
				},
				NameAttribute: ansi.StylePrimitive{
					Color: stringPtr("#7A7AE6"),
				},
				NameClass: ansi.StylePrimitive{
					Color:     stringPtr("#F1F1F1"),
					Underline: boolPtr(true),
					Bold:      boolPtr(true),
				},
				NameDecorator: ansi.StylePrimitive{
					Color: stringPtr("#FFFF87"),
				},
				NameFunction: ansi.StylePrimitive{
					Color: stringPtr("#00D787"),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: stringPtr("#6EEFC0"),
				},
				LiteralString: ansi.StylePrimitive{
					Color: stringPtr("#C69669"),
				},
				LiteralStringEscape: ansi.StylePrimitive{
					Color: stringPtr("#AFFFD7"),
				},
				GenericDeleted: ansi.StylePrimitive{
					Color: stringPtr("#FD5B5B"),
				},
				GenericEmph: ansi.StylePrimitive{
					Italic: boolPtr(true),
				},
				GenericInserted: ansi.StylePrimitive{
					Color: stringPtr("#00D787"),
				},
				GenericStrong: ansi.StylePrimitive{
					Bold: boolPtr(true),
				},
				GenericSubheading: ansi.StylePrimitive{
					Color: stringPtr("#777777"),
				},
				Background: ansi.StylePrimitive{
					BackgroundColor: stringPtr("#373737"),
				},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n🠶 ",
		},
	}

	r, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithEmoji(), glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(width),
		glamour.WithStyles(customStylingConfig),
	)

	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				switch p := part.(type) {
				case genai.Text:
					out, err := r.Render(string(p))
					fmt.Print(out)
					if err != nil {
						log.Fatal("error writing response: ", err)
					}
				case genai.Blob:
					out, err := r.Render(fmt.Sprintf("Blob: MIMEType=%s, DataLength=%d", p.MIMEType, len(p.Data)))
					if err != nil {
						log.Fatal("error writing response: ", err)
					}
					fmt.Print(out)
				}
			}
		}
	}
	// fmt.Println(strings.Repeat("-", width-3))
	fmt.Println(strings.Repeat("─", width-3), "\033[0;37m")
	// fmt.Println(r.Render(string("---")))
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint) *uint {
	return &u
}

func terminalWidth() int {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 100
	}
	return width
}
