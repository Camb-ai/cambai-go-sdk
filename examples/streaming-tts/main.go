package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/camb-ai/cambai-go-sdk"
	"github.com/camb-ai/cambai-go-sdk/client"
	"github.com/camb-ai/cambai-go-sdk/option"
)

func main() {
	apiKey := os.Getenv("CAMB_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set CAMB_API_KEY environment variable")
		return
	}

	c := client.NewClient(
		option.WithAPIKey(apiKey),
	)

	fmt.Println("Sending Streaming TTS request...")
	resp, err := c.TextToSpeech.Tts(
		context.Background(),
		&cambai.CreateStreamTtsRequestPayload{
			Text:        "Hello from Camb AI! This is a Go SDK streaming test.",
			VoiceID:     20303, // Standard voice
			Language:    cambai.CreateStreamTtsRequestPayloadLanguageEnUs,
			SpeechModel: cambai.CreateStreamTtsRequestPayloadSpeechModelMars8.Ptr(),
			OutputConfiguration: &cambai.StreamTtsOutputConfiguration{
				Format: cambai.OutputFormatMp3.Ptr(),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	outputFile := "streaming_output.mp3"
	out, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("failed to create output file: %v", err))
	}
	defer out.Close()

	written, err := io.Copy(out, resp)
	if err != nil {
		panic(fmt.Errorf("failed to save streaming audio: %v", err))
	}

	fmt.Printf("Success! %d bytes of streaming audio saved to %s\n", written, outputFile)
}
