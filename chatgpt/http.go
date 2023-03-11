package chatgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	openAIAPI     = "https://api.openai.com/v1/completions"
	openAIChatAPI = "https://api.openai.com/v1/chat/completions"
)

// ChatGPTResponse is the response of completions api
type ChatGPTResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []Choice         `json:"choices"`
	Usage   map[string]int64 `json:"usage"`
}
type Choice struct {
	Text         string      `json:"text"`
	Index        int64       `json:"index"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

// getDavinci003 performs an http call to the openAI completions API with the given prompt.
func GetDavinci003(chatGPTToken, prompt string) (string, error) {
	fmt.Println("GetDavinici003 Prompt:", prompt)
	data := map[string]interface{}{
		"model":       "text-davinci-003",
		"prompt":      prompt,
		"max_tokens":  1500,
		"temperature": 0,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", openAIAPI, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+chatGPTToken)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode > 299 {
		return "", fmt.Errorf("Status code %d", res.StatusCode)
	}
	var response ChatGPTResponse
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		return "", err
	}
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("No choices in response")
	}
	return response.Choices[0].Text, nil
}

type ChatRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
	// LogitBias map[int]float64 `json:"logit_bias,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatCompletion struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// getDavinci003 performs an http call to the openAI completions API with the given prompt.
func GetGPTTurbo(chatGPTToken string, rx ChatRequest) (*ChatCompletion, error) {
	// fmt.Println("GetGPTTurbo Prompt:", rx)
	body, err := json.Marshal(rx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", openAIChatAPI, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+chatGPTToken)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("Status code %d. Response %s", res.StatusCode, string(resBody))
	}
	var completion ChatCompletion
	err = json.Unmarshal(resBody, &completion)
	if err != nil {
		return nil, err
	}

	// fmt.Println("GetGPTTurbo Response:", completion)

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("No choices in response")
	}
	return &completion, nil
}
