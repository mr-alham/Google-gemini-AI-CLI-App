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

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	terminal "golang.org/x/term"
)

// getting configuration data from the json file
type Config struct {
	GeminiAPIKey      string `json:"GEMINI_API_KEY"`
	GeminiModel       string `json:"GEMINI_MODEL"`
	SystemInstruction string `json:"SYSTEM_INSTRUCTION"`
	GenerationConfig  struct {
		Temperature      float32 `json:"temperature"`
		TopP             float32 `json:"top_p"`
		TopK             int32   `json:"top_k"`
		MaxOutputTokens  *int32  `json:"max_output_tokens"`
		ResponseMimeType string  `json:"response_mime_type"`
	} `json:"GENERATION_CONFIG"`

	// you can get safety settings information from,
	// https://github.com/google/generative-ai-go/blob/v0.13.0/genai/generativelanguagepb_veneer.gen.go#L817
	SafetySettings []struct {
		Threshold string `json:"threshold"`
	} `json:"SAFETY_SETTINGS"`
}

func configFile() (string, error) {
	path1 := "/usr/local/etc/gemini.conf.d/keys.json"
	path2 := "gemini.conf.d/keys.json"

	if _, err := os.Stat(path1); err == nil {
		return path1, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error accessing %s: %v", path1, err)
	}

	if _, err := os.Stat(path2); err == nil {
		return path2, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error accessing %s: %v", path2, err)
	}

	return "", fmt.Errorf("configuration file not found")
}

func main() {
	err := welcomeBanner()
	if err != nil {
		fmt.Println("Error:")
	}

	config, err := loadConfig()
	if err != nil {
		log.Panic("Error loading config:")
	}

	ctx := context.Background()

	client, err := createClient(ctx, config.GeminiAPIKey)
	if err != nil {
		log.Panic("Error creating client:", err)
	}

	defer client.Close()

	model, err := createAndConfigureClient(ctx, config)
	if err != nil {
		log.Panic("Error Creating and Configuring client:")
	}

	model.SafetySettings = configureSafetySettings(config.SafetySettings)

	if len(os.Args) > 1 {
		handleArgs(ctx, model, os.Args)
	} else {
		err := generateTextFromPrompt(ctx, model)
		if err != nil {
			fmt.Println("Error:")
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
			fmt.Println()
			err := generateTextFromPrompt(ctx, model)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}

		imgData, err := os.ReadFile(pathToImage)
		if err != nil {
			fmt.Println("Error reading image file: ", err)
			continue
		}

		fmt.Print("\033[0;1;38;5;28mPrompt: \033[0;38;5;254m")
		scanner.Scan()
		userPrompt = scanner.Text()

		if Err := scanner.Err(); Err != nil {
			fmt.Println("Error scanning prompt: ", Err)
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
		fmt.Print("\033[0;1;38;5;28mPrompt: \033[0;38;5;254m")
		scanner.Scan()
		userPrompt = scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error scanning prompt: ", err)
			continue

		} else if userPrompt == "" {
			fmt.Println("The Prompt is empty.")
			continue
		} else if strings.ToLower(userPrompt) == "image mode" {
			fmt.Println()
			err := generateTextFromImage(ctx, model)
			if err != nil {
				fmt.Println("Error:")
			}
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
			Format: "\n🝔------------------------------🝔\n",
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

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithEmoji(), glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(width),
		glamour.WithStyles(customStylingConfig),
	)

	if err != nil {
		fmt.Println("Error rendering glamour output", err)
	}

	fmt.Print("\033[0;1;38;5;28m\nResponse: \033[0;38;5;254m")

	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				switch p := part.(type) {
				case genai.Text:
					out, err := r.Render(string(p))
					fmt.Print(out)
					if err != nil {
						log.Panic("error writing response: ", err)
					}
				case genai.Blob:
					out, err := r.Render(fmt.Sprintf("Blob: MIMEType=%s, DataLength=%d", p.MIMEType, len(p.Data)))
					if err != nil {
						log.Panic("error writing response: ", err)
					}
					fmt.Print(out)
				}
			}
		}
	}
	fmt.Println("\033[0;1;2;95m"+strings.Repeat("─", width-3), "\033[0;37m")
}

func welcomeBanner() error {
	width := terminalWidth()
	titleA := "---------------🝔---------------"
	title := "⚙️  Gemini-AI on Terminal ⚙️"

	spaceA := strings.Repeat(" ", ((width - 31) / 2))
	spaceTitle := strings.Repeat(" ", ((width - 27) / 2))

	fmt.Println("\033[1;38;5;178m" + spaceA + titleA + "\033[0;37m")
	fmt.Println("\033[1;38;5;38m" + spaceTitle + title + "\033[0m")
	fmt.Println("\033[1;38;5;178m" + spaceA + titleA + "\033[0;37m")

	return nil
}

func loadConfig() (*Config, error) {
	// if you intend to use a different file for json specify it here
	dataFile, err := configFile()

	if err != nil {
		log.Panic("Error couldn't find the config file")
	}

	file, err := os.Open(dataFile)
	if err != nil {
		log.Panic("Error opening config file: ", err)
	}
	defer file.Close()

	var config Config

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)

	if err != nil {
		log.Panic("Error decoding config JSON: ", err)
	}

	if config.GeminiAPIKey == "your gemini api key" {
		fmt.Println("To use this chatbot you need a API key,")
		fmt.Println("If you don't posses a Gemini-API key, get one from `https://aistudio.google.com/app/apikey`")
		log.Panic("Then paste the key in `keys.json` as showed in the documentation.")
	}

	return &config, nil
}

func createAndConfigureClient(ctx context.Context, config *Config) (*genai.GenerativeModel, error) {

	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GeminiAPIKey))
	if err != nil {
		log.Panic("Error creating client:", err)
	}

	model := client.GenerativeModel(config.GeminiModel)
	model.SetTemperature(config.GenerationConfig.Temperature)
	model.SetTopP(config.GenerationConfig.TopP)
	model.SetTopK(config.GenerationConfig.TopK)
	model.MaxOutputTokens = config.GenerationConfig.MaxOutputTokens
	model.ResponseMIMEType = config.GenerationConfig.ResponseMimeType
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(config.SystemInstruction)},
	}

	return model, nil
}

func configureSafetySettings(safetySettings []struct {
	Threshold string `json:"threshold"`
}) []*genai.SafetySetting {

	var thresholds = []uint8{2, 2, 2, 2}
	for index, t := range safetySettings {
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

	return []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockThreshold(thresholds[0])},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockThreshold(thresholds[1])},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockThreshold(thresholds[2])},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockThreshold(thresholds[3])},
	}
}

func handleArgs(ctx context.Context, model *genai.GenerativeModel, args []string) {

	switch strings.ToLower(args[1]) {
	case "--image":
		err := generateTextFromImage(ctx, model)
		if err != nil {
			fmt.Println("Error:")
		}
	case "--text":
		err := generateTextFromPrompt(ctx, model)
		if err != nil {
			fmt.Println("Error:")
		}
	case "--help":
		help()
	case "-h":
		help()
	default:
		fmt.Println("Error `" + args[1] + "` not identified.")
		fmt.Println("Redirecting to Text-to-Text Model...")
		fmt.Println()
		err := generateTextFromPrompt(ctx, model)
		if err != nil {
			fmt.Println("Error:")
		}
	}
}

func help() {
	width := terminalWidth()

	fmt.Println("\033[0;1;2;95m"+strings.Repeat("─", width-3), "\033[0;37m")
	fmt.Println("\033[0;1;37mSYNOPSIS")
	fmt.Println("\033[0;32m\tgemini\033[0;37m [OPTIONS]")
	fmt.Println()
	fmt.Println("\033[0;1;37mDESCRIPTION")
	fmt.Println("\033[0;32m\tThe `gemini` allows you to interact with the Gemini AI directly from your terminal.")
	fmt.Println("\tversatile tool supports both text-to-text and multimodal (text-and-image) interactions,")
	fmt.Println("\tincluding colorful engaging output with code syntax highlighting")
	fmt.Println("\tenabling you to generate content based on text prompts or image inputs.")
	fmt.Println()
	fmt.Println("\033[0;1;37mOPTIONS")
	fmt.Println("\033[0;1;37m\t--text")
	fmt.Println("\033[0;32m\t\tStart Conversation with the Text-to-Text Mode. (Default)")
	fmt.Println("\033[0;1;37m\t--image")
	fmt.Println("\033[0;32m\t\tStart Conversation with the Multi-Mode Model.")
	fmt.Println("\033[0;32m\t\tThis model will allow you to give inputs as Text as well as Image")
	fmt.Println("\033[0;1;32m\t\t- When asking for the path of the image you can,")
	fmt.Println("\033[0;32m\t\t\t- Type the path (absolute path)")
	fmt.Println("\033[0;32m\t\t\t- Copy the image path to the Clipboard and hit Enter")
	fmt.Println("\033[0;32m\t\t\t- Drag and drop the image to the Terminal")
	fmt.Println("\033[0;1;37m\t-h, --help")
	fmt.Println("\033[0;32m\t\tShows this help message and exit.")
	fmt.Println()
	fmt.Println("\033[0;1;37mDocumentation")
	fmt.Println("\033[0;32m\tTo see the documentation visit, \033[0;3;35mhttps://github.com/mr-alham/Google-gemini-AI-CLI-App")
	fmt.Println()
	fmt.Println("\033[0;1;37mAUTHOR")
	fmt.Println("\033[0;32m\tWritten by: \033[0;37mALHAM")
	fmt.Println("\033[0;32m\tGithub: \033[0;3;35mhttps://github.com/mr-alham/")
	fmt.Println("\033[0;32m\tTwitter:\033[0;3;35mhttps://www.twitter.com/@alham__aa")
	fmt.Println("\033[0;32m\tEmail:  \033[0;37malham@duck.com")
	fmt.Println()
	fmt.Println("\033[0;1;2;95m"+strings.Repeat("─", width-3), "\033[0;37m")

}

func boolPtr(b bool) *bool { return &b }

func stringPtr(s string) *string { return &s }

func uintPtr(u uint) *uint { return &u }

func terminalWidth() int {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 100
	}
	return width
}
