package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/camb-ai/cambai-go-sdk"
	"github.com/camb-ai/cambai-go-sdk/client"
	"github.com/camb-ai/cambai-go-sdk/option"
)

func main() {
	apiKey := os.Getenv("CAMB_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set CAMB_API_KEY")
		return
	}

	c := client.NewClient(option.WithAPIKey(apiKey))

	fmt.Println("Starting Dubbing Task...")

	videoURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	resp, err := c.Dub.EndToEndDubbing(
		context.Background(),
		&cambai.EndToEndDubbingRequestPayload{
			VideoURL:        videoURL,
			SourceLanguage:  cambai.LanguagesEnUs,
			TargetLanguages: []cambai.Languages{cambai.LanguagesFrFr},
		},
	)
	if err != nil {
		panic(err)
	}

	// TaskID is *string in response
	if resp.TaskID == nil {
		panic("TaskID is nil")
	}
	taskID := *resp.TaskID
	fmt.Printf("Dubbing Task Started: %s\n", taskID)

	for {
		time.Sleep(5 * time.Second)
		status, err := c.Dub.GetEndToEndDubbingStatus(
			context.Background(),
			taskID,
			&cambai.GetEndToEndDubbingStatusDubTaskIDGetRequest{},
		)
		if err != nil {
			fmt.Printf("Error polling: %v\n", err)
			continue
		}

		fmt.Printf("Status: %v\n", status.Status)

		if status.Status == cambai.TaskStatusSuccess {
			if status.RunID != nil {
				runID := *status.RunID
				res, err := c.Dub.GetDubbedRunInfo(
					context.Background(),
					&runID,
					&cambai.GetDubbedRunInfoDubResultRunIDGetRequest{},
				)
				if err != nil {
					panic(err)
				}

				if res.DubbingResult != nil && res.DubbingResult.VideoURL != nil {
					fmt.Printf("Success! Video URL: %s\n", *res.DubbingResult.VideoURL)
				} else {
					fmt.Println("Success, but no video URL found in result.")
				}
			}
			break
		} else if status.Status == cambai.TaskStatusError {
			fmt.Println("Task failed.")
			if status.ExceptionReason != nil {
				fmt.Printf("Reason: %v\n", status.ExceptionReason.String)
			}
			break
		}
	}
}
