# Bots Boilerplate Readme - 99.9% written by mushi bot

This is a Go program that interacts with OpenAI's GPT-3 API to enable conversations with virtual assistants. The program is capable of running in three different modes:
- `single`: In this mode, the user communicates with a single virtual assistant (bot) in a back-and-forth conversation. The bot's response is converted to speech via Google Cloud Text-to-Speech API.
- `blog`: In this mode, the user must specify the bot and a topic. The bot then generates a blog post in Markdown and a YAML metadata file. The generated content is printed on the console.
- `team`: In this mode, a group of bots maintain a circular conversation. This is useful for simulating a teacher/student, or a team working on a project with different members being different bots. The bot's response is again converted to speech via Google Cloud Text-to-Speech API.

## Prerequisites

- You need to have a valid GPT-3 OpenAI token.
- You may optionally have access to Google Cloud Text-to-Speech API.

## Installation
1. Clone this repository to your local machine.
2. Install dependencies by running the following command:
   ```
   go get https://github.com/ds0nt/bots
   ```

## Usage

1. Set your OpenAI token using the `CHATGPT_TOKEN` environment variable:
   ```
   export CHATGPT_TOKEN=<your_openai_token>
   ```
2. You can run the program by typing `go run main.go <mode>` in the terminal.

For example, running the program in single bot mode can be done using the following command:

```
go run main.go single BotName
```

3. Navigate to the `bots` folder to define new bots for use in the program.

## Modes

### Single
In this mode, users can interact with a single bot. The bot's responses will be converted to speech if Google Cloud Text-to-Speech API is available.

Usage example:

```
go run main.go single BotName
```

### Blog
In this mode, the user must specify the bot and a topic. The bot will generate a blog post in Markdown and a YAML metadata file. The content will be printed on the console.

Usage example:

```
go run main.go blog BotName "Topic"
```

### Team
In this mode, users can interact with a group of bots in a circular conversation. The bot's responses will be converted to speech if Google Cloud Text-to-Speech API is available.

Usage example:

```
go run main.go team BotName1 BotName2 BotName3
```

In the above command, `BotName1`, `BotName2`, and `BotName3` are the names of the bots that will participate in the conversation.

## Additional Flags

- `--no-tts` or `-T`: use this flag to disable text-to-speech conversion. Useful if you do not have access to Google Cloud Text-to-Speech API.

## Additional Notes

- The `getUserInput()` function allows users to input multiline text through `<<<` and `>>>` delimiters.
- The `tts()` function converts text to speech using the Google Cloud Text-to-Speech API. If this is not available, the function will do nothing.
- The `removeOld()` function removes messages from memory if the message count or total message length exceeds a threshold.
- It's a good idea to use [direnv](https://direnv.net/) to handle your environment variables, including the ChatGPT token.