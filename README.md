# Google Gemini-AI CLI App

GeminiAI-Terminal is a powerful command-line interface (CLI) application that leverages Google's Gemini AI to generate text based on user prompts. This allows users to interact with advanced AI capabilities directly from their terminal, making it an essential tool for developers, and AI enthusiasts who prefer a terminal workflow.

[![Super-Linter](https://github.com/mr-alham/Google-gemini-AI-CLI-App/actions/workflows/linter.yaml/badge.svg)](https://github.com/marketplace/actions/super-linter) [![CodeQl](https://github.com/mr-alham/Google-gemini-AI-CLI-App/actions/workflows/codeQl.yaml/badge.svg)](https://github.com/marketplace/actions/codeql) [![Super-Linter](https://github.com/mr-alham/Google-gemini-AI-CLI-App/actions/workflows/release.yaml/badge.svg)](https://github.com/marketplace/actions/super-linter)

<p align="center">
  <img src="https://i.ibb.co/p37NzwH/AI-Generated-Spam-Love-Letter.png" alt="Demo Screenshot of Using Text-to-Text model">
</p>

## Table of Contents

* [Features](#features)

* [Prerequisites](#prerequisites)

* [Installation](#installation)
  * [Execute-Directly](#run-the-binary)
  * []()

* [Configuration](#configuration)
  * [Configuration Explanation](#configuration-explanation)
  * [Safety Settings](#safety-settings)

* [Usage](#usage)

* [Command-Line Arguments](#command-line-arguments)

* [Examples](#examples)

* [License](#license)

* [Author](#author)

## Features

* **Text-to-Text Model** - Engage in conversations with Gemini-AI using text Prompts
* **Multi-Mode Model** - Get output from given inputs text prompt and image
* **Clipboard Integration** - Instead of manually typing the file path, copy path to the clipboard and press enter
* **Flexible Customizations** - Flexibly configure/modify configurations from a single JSON file
* **Colorful Engaging Terminal Outputs** - Enjoy colorful engaging outputs including Code highlighting

## Prerequisites

* **Gemini API Key**
To use this app, you'll need an API key. If you don't already have one, create a key in Google AI Studio
Create an API key from [Google AI Studio](https://aistudio.google.com/app/apikey)

* **Go**
version 1.20 or higher

* **For best experience, use Gemini AI with Yakuake terminal to feel the productivity boost** (optional)
  ```sh
  sudo pacman -S yakuake
  ```

## Installation

### Run the Binary

Use the app directly withour Compiling or downloading dependencies,

Download latest version of the app from [GitHub Release](https://github.com/mr-alham/Google-gemini-AI-CLI-App/releases/), Extract the zip file and Execute as shown below.

#### For Linux users,

```sh
# first navigate to the directory where the zip file is
unzip Gemini-AI-CLI-App.zip
chmod +x gemini
./gemini
```

You have to setup the API key [As shown here](#configuration-explanation)

Follow these steps on topic [#Usage Method 3](#usage), to use gemini as a regular Linux command.

>! Only tested on Linux

---
#### Compile/Build the program

1. **Clone the Repository**

    ```sh
    git clone https://github.com/mr-alham/Google-gemini-AI-CLI-App.git
    cd Google-gemini-AI-CLI-App
    ```

2. **Install Dependencies**

    * Go Generative-AI SDK Package

        ```sh
        go get github.com/google/generative-ai-go
        ```

    * Options to configure Google API client

        ```sh
        go get google.golang.org/api/option
        ```

    * Clipboard

        ```sh
        go get github.com/atotto/clipboard
        ```

    * Glamour

        ```sh
        go get github.com/charmbracelet/glamour
        ```

    * Utilities to Interact with Terminal

        ```sh
        go get golang.org/x/term
        ```

3. **Build the Application**

    ```sh
    go build -o gemini
    ```

## Configuration

The application can be configured with a single configuration file `keys.json`. Below is the configuration file with explanations for each configuration,

```json
{
    "GEMINI_API_KEY": "your gemini api key",
    "GEMINI_MODEL": "gemini-1.5-pro-latest",
    "SYSTEM_INSTRUCTION": "",
    "GENERATION_CONFIG": {
        "Temperature": 0.9,
        "top_p": 0.95,
        "top_k": 100,
        "max_output_tokens": 8192,
        "response_mime_type": "text/plain"
    },
    "SAFETY_SETTINGS": [
        {
            "category": "HARM_CATEGORY_HARASSMENT",
            "threshold": "BLOCK_MEDIUM_AND_ABOVE"
        },
        {
            "category": "HARM_CATEGORY_HATE_SPEECH",
            "threshold": "BLOCK_MEDIUM_AND_ABOVE"
        },
        {
            "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
            "threshold": "BLOCK_MEDIUM_AND_ABOVE"
        },
        {
            "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
            "threshold": "BLOCK_MEDIUM_AND_ABOVE"
        }
    ]
}
```

### Configuration Explanation

* **GEMINI_API_KEY**

  * Your API key for accessing the Gemini AI services. This is required to authenticate your requests.
    Replace `your gemini api key` with your API key

    ```json
    { "GEMINI_API_KEY": "your gemini api key" }
    ```

* **GEMINI_Model**

  * Generative AI models are able to create content from varying types of data input including text, images and audio.

  | Model                  | Input                     | Optimized for                                                                                                         |
  |------------------------|---------------------------|-----------------------------------------------------------------------------------------------------------------------|
  |gemini-1.5-pro (Default)|Audio, Image,Video and Text|Complex reasoning tasks such as code and text generation, text editing, problem solving, data extraction and generation|
  |gemini-1.5-flash        |Audio, Image,Video and Text|Fast and versatile performance across a diverse variety of tasks                                                       |
  |gemini-1.0-pro          |Text                       |Natural language tasks, multi-turn text and code chat, and code generation                                             |
  |gemini-pro-vision       |Audio, Image,Video and Text|Visual-related tasks, like generating image descriptions or identifying objects in images                              |

* **SYSTEM_INSTRUCTION**

  * System instructions enable users to steer the behavior of the model based on their specific needs and use cases

      ```json
      "SYSTEM_INSTRUCTION": "You are a cat. Your name is Neko."
      ```

* **Temperature**

  * The temperature controls the degree of randomness in token selection
  The temperature can be changed in the range of 0 to 1

* **top_p**

  * model alters the way tokens are selected for output. It involves selecting tokens from the most probable to the least probable until their cumulative probability equals the topP value.
  For example, if tokens A, B, and C have probabilities of 0.3, 0.2, and 0.1, and the topP value is 0.5, the model picks either A or B as the next token using temperature sampling, excluding C. The default topP value is 0.95. This parameter helps filter the most probable tokens and exclude less probable ones.

* **top_K**

  * determines the number of the most probable tokens to be considered for output selection. A topK value of 1 (greedy decoding) selects the most probable token, while a topK of 3 selects from the top 3 probabilities. This helps narrow down the options for the next token during the generation process.

* **max_output_tokens**

  * Specifies the maximum number of tokens that can be generated in the response.
  A token is approximately four characters. 100 tokens correspond to roughly 60-80 words

#### Safety Settings

  | Safety Category   | Description                                                                 |
  |-------------------|-----------------------------------------------------------------------------|
  | Harassment        | Negative or harmful comments targeting identity and/or protected attributes.|
  | Hate Speech       | Content that is rude, disrespectful, or profane.                            |
  | Sexually explicit | Contains references to sexual acts or other lewd content.                   |
  | Dangerous         | Promotes, facilitates, or encourages harmful acts.                          |

  | Threshold                        | Description                                                  |
  |----------------------------------|--------------------------------------------------------------|
  | BLOCK_NONE                       | Always show regardless of probability of unsafe content      |
  | BLOCK_ONLY_HIGH                  | Block when high probability of unsafe content                |
  | BLOCK_MEDIUM_AND_ABOVE           | Block when medium or high probability of unsafe content      |
  | BLOCK_LOW_AND_ABOVE              | Block when low, medium or high probability of unsafe content |
  | HARM_BLOCK_THRESHOLD_UNSPECIFIED | Threshold is unspecified, block using default threshold      |

  > Note: You have to give safety setting in the order of the json file, Do not change the given order of safety category. Just change the Threshold

## Usage

1. Run the script directly

    ```sh
    go run main.go
    ```

    > You should be in the same directory as the file main.go

2. Run the binary, Which is [compiled](#installation)

    ```sh
    ./gemini
    ```

3. If you are a Linux/Unix User you can move the gemini binary to `/usr/local/bin/` and access gemini as a standard Command

    ```sh
    chmod +x gemini
    sudo mv gemini /usr/local/bin
    sudo mv Gemini /usr/local/etc  # Move the directory which has the config file to /usr/local/etc
    ```

    Now you can use the app as an ordinary terminal application

    ```sh
    gemini
    ```

## Command-Line Arguments

* **--text**
  * Start Conversation with the Text-to-Text Mode. (Default)

* **--image**
  * Start Conversation with the Multi-Mode Model.
    This model will allow you to give inputs as Text as well as Image
    * **When asking for the path of the image you can,**
      * Type the path (absolute path)
      * Copy the image path to the Clipboard and hit Enter
      * Drag and drop the image to the Terminal

* **-h**, **--help**
  * Shows help menu and exit

## Examples

  ```sh
  go run main.go
  ```

<p align="center">
    <img src="https://i.ibb.co/vcXGHHR/Gemini-AI-on-Terminal.png" alt="screenshot of running gemini code directly without compiling">
</p>

---
  ```sh
  go build -o gemini
  ./gemini
  ```

<p align="center">
    <img src="https://i.ibb.co/qgH9CPN/Gemini-AI-Terminal-Session.png" alt="Screenshot of building gemini and executing the executable">
</p>

---
  ```sh
  gemini
  ```

---
  ```sh
  gemini --help
  ```

---
  ```sh
  gemini --image
  ```

  <p align="center">
    <img src="https://i.ibb.co/KwFF4CB/Gemini-AI-Ready-for-Image-Mode.png" alt="Screenshot of building gemini and executing the executable">
</p>

***The same project in python can be found at: [mr-alham/Google-Gemini-AI-on-the-Terminal](https://github.com/mr-alham/Google-Gemini-AI-on-the-Terminal), But it is deprecated.***

## License

This project is licensed under the MIT License.

## Author

Written and maintaining by Alham

* GitHub : [mr-alham](https://github.com/mr-alham)
* Twitter (X) : [@alham__aa](https://twitter.com/@alham__aa)
* Email : [alham@duck.com](mailto:alham@duck.com?subject=Github%20Google-gemini-AI-CLI-App&body=Hello!)
