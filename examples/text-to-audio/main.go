package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

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

	c := client.NewClient(option.WithAPIKey(apiKey))

	fmt.Println("Creating Text-to-Audio (Sound) task...")

	resp, err := c.TextToAudio.CreateTextToAudio(
		context.Background(),
		&cambai.CreateTextToAudioRequestPayload{
			Prompt:    "A futuristic sci-fi laser sound effect",
			Duration:  cambai.Float64(3.0),
			AudioType: cambai.TextToAudioTypeSound.Ptr(),
		},
	)
	if err != nil {
		panic(err)
	}

	taskID := *resp.TaskID
	fmt.Printf("Task created with ID: %s\n", taskID)

	// Poll for status
	for {
		time.Sleep(2 * time.Second)
		status, err := c.TextToAudio.GetTextToAudioStatus(
			context.Background(),
			taskID,
			&cambai.GetTextToAudioStatusTextToSoundTaskIDGetRequest{},
		)
		if err != nil {
			fmt.Printf("Error polling: %v\n", err)
			continue
		}

		fmt.Printf("Current Status: %v\n", status.Status)

		if status.Status == cambai.TaskStatusSuccess {
			fmt.Println("Task completed! Downloading result...")
			runID := *status.RunID

			// Get result stream
			audioStream, err := c.TextToAudio.GetTextToAudioResult(
				context.Background(),
				&runID,
				&cambai.GetTextToAudioResultTextToSoundResultRunIDGetRequest{},
			)
			if err != nil {
				panic(err)
			}

			outputFile := "text_to_audio_output.wav"
			out, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer out.Close()

			written, _ := io.Copy(out, audioStream)
			fmt.Printf("✓ Success! Audio (%d bytes) saved to %s\n", written, outputFile)
			break
		} else if status.Status == cambai.TaskStatusError {
			fmt.Printf("Task failed: %v\n", status.Message)
			break
		}
	}
}
