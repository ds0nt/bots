package main

import (
	"bots/chatgpt"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"bots/bots"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/alecthomas/chroma/quick"
	"github.com/spf13/cobra"
)

var token = os.Getenv("CHATGPT_TOKEN")

var rootCmd = &cobra.Command{
	Use: "bots",
}
var logFile = "bots.log"

var middlewares = []MessageMiddleware{
	LogToFileMiddleware(logFile),
	SyntaxHighlightCodeBlocksMiddleware,
}

func main() {
	rootCmd.PersistentFlags().BoolP("no-tts", "T", false, "Disable text to speech")
	rootCmd.PersistentFlags().BoolP("no-logfile", "L", false, "Disable saving log to speech")

	singleCmd := &cobra.Command{
		Use:   "single",
		Short: "Run in single bot mode",
		RunE:  singleCmd,
	}
	blogCmd := &cobra.Command{
		Use:   "blog",
		Short: "Run bot in blog mode",
		RunE:  blogCmd,
	}

	teamCmd := &cobra.Command{
		Use:   "team",
		Short: "Run in team bot mode",
		RunE:  teamCmd,
	}

	rootCmd.AddCommand(singleCmd)
	rootCmd.AddCommand(blogCmd)
	rootCmd.AddCommand(teamCmd)

	rootCmd.Execute()

}

func blogCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println("Please specify a bot")
		return nil
	}
	if len(args) < 2 {
		fmt.Println("Please specify a blot topic")
		return nil
	}
	bot := args[0]
	taskStr := fmt.Sprintf(`Task: Be the blogger for Forky, a mindmapping app. Output one markdown file, and one yml file. The markdown should contain a blog post you write. The yml file will contain some metadata... i.e. { title, url: "/<slug>" }. 
	The topic is: %s
	Please write a good one that will help my site grow organically!!`, args[1])
	// Test data
	rx := chatgpt.ChatRequest{
		Model:     "gpt-3.5-turbo-0301",
		MaxTokens: 500,
		Messages: []chatgpt.ChatMessage{
			{Role: "system", Content: bots.Bots[bot].Prompt},
			{Role: "user",
				Content: taskStr,
			},
		},
	}

	// get bot message
	completion, err := chatgpt.GetGPTTurbo(token, rx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	txt := applyMiddleware(middlewares, "blogger", completion.Choices[0].Message.Content)
	fmt.Println(txt)
	return nil
}

func singleCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println("Please specify a bot")
		return nil
	}
	bot := args[0]
	// Test data
	rx := chatgpt.ChatRequest{
		Model:     "gpt-3.5-turbo-0301",
		MaxTokens: 500,
		Messages: []chatgpt.ChatMessage{
			{Role: "system", Content: bots.Bots[bot].Prompt},
		},
	}

	for {
		// prompt user
		input := applyMiddleware(middlewares, "user", getUserInput())

		rx.Messages = append(rx.Messages, chatgpt.ChatMessage{Role: "user", Content: input})
		removeOld(rx.Messages, 2500)
		// get bot message
		completion, err := chatgpt.GetGPTTurbo(token, rx)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// output bot message
		txt := applyMiddleware(middlewares, bot, completion.Choices[0].Message.Content)
		fmt.Println(wrapInMagenta(bot)+":", txt)
		rx.Messages = append(rx.Messages, completion.Choices[0].Message)
		err = tts(txt, "en-US-Wavenet-H")
		if err != nil {
			fmt.Println(err)
		}

	}
}

func getUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(wrapInBlue("user") + ": ")
	text, _ := reader.ReadString('\n')

	// multiline support for user input
	if text == "<<<\n" {
		for {
			_text, _ := reader.ReadString('\n')
			if _text == ">>>\n" {
				break
			}
			text += "\n" + _text
		}
	}
	return text
}

func teamCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		fmt.Println("Please specify two bots")
		return nil
	}

	botNames := args

	bot1GreetsBots := fmt.Sprintf("Hello. I am %s. Let us commence.", botNames[0])

	// everyone sees themselves as "user", and everyone else as "assistant". were talking in a circle
	rxs := make([]*chatgpt.ChatRequest, len(botNames))
	for k, _ := range rxs {
		rx := &chatgpt.ChatRequest{}
		rx.Messages = []chatgpt.ChatMessage{
			{Role: "system", Content: bots.Bots[botNames[k]].Prompt},
			{Role: "assistant", Content: bot1GreetsBots},
		}
		rx.Model = "gpt-3.5-turbo-0301"
		rx.MaxTokens = 500
		rx.Messages[0].Content += "\nThe participants in the chat are: " + strings.Join(botNames, ", ")
		rxs[k] = rx
	}
	rxs[0].Messages[1].Role = "user"

	// Infinite conversation
	for {
		for i, rx := range rxs {

			currentBot := botNames[i]
			// send request
			// pretty.Println(rx.Messages)
			reply, err := chatgpt.GetGPTTurbo(token, *rx)
			if err != nil {
				fmt.Println(err)
				return err
			}
			txt := applyMiddleware(middlewares, currentBot, reply.Choices[0].Message.Content)

			fmt.Println(wrapInGreen(currentBot)+":", txt)
			if i == 0 {
				err = tts(txt, "en-US-Wavenet-H")
			} else {
				err = tts(txt, "en-US-Wavenet-A")
			}
			if err != nil {
				fmt.Println(err)
			}

			// prepare next request
			// set up next data
			// each bot takes a turn receiving the last chat as user input and responding as the assistant version of itself.
			for j, _rx := range rxs {
				if j == (i+1)%len(rxs) {
					_rx.Messages = append(_rx.Messages, chatgpt.ChatMessage{
						Content: txt,
						Role:    "user",
					})
				} else {
					_rx.Messages = append(_rx.Messages, reply.Choices[0].Message)
				}
				removeOld(_rx.Messages, 2500)
			}
		}
	}
}

func removeOld(messages []chatgpt.ChatMessage, maxTokens int) {
	// keep first message and last 29 messages
	if len(messages) > 30 {
		fmt.Println(wrapInRed("more than 30 messages"))
		messages = append([]chatgpt.ChatMessage{messages[0]}, messages[len(messages)-29:]...)
	}

	// if too many tokens, remove the second message
	for numberofTokens(messages) > maxTokens {
		fmt.Println(wrapInRed("more than maxTokens tokens, estimated " + strconv.Itoa(numberofTokens(messages))))
		messages = append([]chatgpt.ChatMessage{messages[0]}, messages[2:]...)
	}
}

func numberofTokens(messages []chatgpt.ChatMessage) int {
	n := 0
	for _, m := range messages {
		n += len(m.Content)
	}
	return n / 4
}

func wrapInRed(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func wrapInBlue(text string) string {
	return "\033[34m" + text + "\033[0m"
}

func wrapInGreen(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func wrapInYellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func wrapInMagenta(text string) string {
	return "\033[35m" + text + "\033[0m"
}

func wrapInCyan(text string) string {
	return "\033[36m" + text + "\033[0m"
}

func wrapInWhite(text string) string {
	return "\033[37m" + text + "\033[0m"
}

func tts(text string, voice string) error {
	if noTTS, _ := rootCmd.PersistentFlags().GetBool("no-tts"); noTTS {
		return nil
	}
	lines := strings.Split(text, "\n")
	newLines := []string{}
	inCodeBlock := false
	for _, v := range lines {
		if strings.HasPrefix(v, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				newLines = append(newLines, "code block.")
			}
			continue
		}
		if !inCodeBlock {
			newLines = append(newLines, v)
		}
	}
	text = strings.Join(newLines, "\n")
	text = strings.ReplaceAll(text, "`", "\"")

	// regex replace words separated by a . with the word dot if they are longer than 3 characters
	text = regexp.MustCompile(`(\w\w\w\w*)\.(\w\w\w\w*)`).ReplaceAllString(text, "$1 dot $2")

	client, err := texttospeech.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			Name:         voice,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}

	// The resp's AudioContent is binary.
	filename := "/tmp/output.mp3"
	err = ioutil.WriteFile(filename, resp.AudioContent, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// play the mp3 file
	cmd := exec.Command("mpg123", filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func applyMiddleware(middlewares []MessageMiddleware, sender string, text string) string {
	for _, middleware := range middlewares {
		text = middleware(sender, text)
	}
	return text
}

type MessageMiddleware func(sender string, text string) string

// LogToFileMiddleware
func LogToFileMiddleware(filename string) MessageMiddleware {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return func(sender string, text string) string {
		_, err := f.WriteString(sender + ": " + text)
		if err != nil {
			log.Fatal(err)
		}
		return text
	}
}

// Syntax Highlight Code Blocks Middleware
func SyntaxHighlightCodeBlocksMiddleware(sender, text string) string {
	language := "go"
	// Highlight the source code

	inCode := false
	newLines := []string{}
	codeLines := []string{}

	// replace code blocks with syntax highlighted code blocks
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "```") {
			inCode = !inCode
			if inCode {
				language = strings.TrimPrefix(line, "```")
			} else {
				w := &bytes.Buffer{}
				err := quick.Highlight(w, strings.Join(codeLines, "\n"), language, "terminal256", "monokai")
				if err != nil {
					panic(err)
				}
				newLines = append(newLines, "```"+language)
				newLines = append(newLines, w.String())
				newLines = append(newLines, "```")
				codeLines = []string{}
			}
			continue
		}

		if inCode {
			codeLines = append(codeLines, line)
		} else {
			newLines = append(newLines, line)
		}
	}
	return strings.Join(newLines, "\n")
}
