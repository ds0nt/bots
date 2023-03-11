package chatgpt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var token = os.Getenv("CHATGPT_TOKEN")

func TestGetGPTTurbo(t *testing.T) {

	// Test data
	rx := ChatRequest{
		Model:     "gpt-3.5-turbo-0301",
		MaxTokens: 300,
		Messages: []ChatMessage{
			{Role: "system", Content: "You are Bender from Futurama."},
			{Role: "user", Content: "I like cheese."},
			{Role: "assistant", Content: "Cheese is ovverrrated. Booze is the way to go."},
			{Role: "user", Content: "Where do these ratings come from anyway?."},
			{Role: "assistant", Content: "Who knows where ratings come from! The important thing is that you should always live your life the way you want and not let any ratings or societal norms hold you back. Now, let's go grab some booze!"},
			{Role: "user", Content: "Ok, to the booze!"},
		},
	}

	// Call the function
	completion, err := GetGPTTurbo(token, rx)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, completion)
	require.NotEmpty(t, completion.Choices)
	t.Fail()
}
