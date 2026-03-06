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
	c := client.NewClient(option.WithAPIKey(os.Getenv("CAMB_API_KEY")))

	fmt.Println("Streaming TTS...")

	resp, err := c.TextToSpeech.Tts(
		context.Background(),
		&cambai.CreateStreamTtsRequestPayload{
			Text:        "Hello from Camb AI Go SDK!",
			VoiceID:     20303,
			Language:    cambai.CreateStreamTtsRequestPayloadLanguageEnUs,
			SpeechModel: cambai.CreateStreamTtsRequestPayloadSpeechModelMarsPro.Ptr(),
			OutputConfiguration: &cambai.StreamTtsOutputConfiguration{
				Format: cambai.OutputFormatWav.Ptr(),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	out, _ := os.Create("tts_output.wav")
	defer out.Close()

	io.Copy(out, resp)
	fmt.Println("✓ Success! Audio saved to tts_output.wav")
}
