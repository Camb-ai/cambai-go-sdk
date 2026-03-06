package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	cambai "github.com/camb-ai/cambai-go-sdk"
	"github.com/camb-ai/cambai-go-sdk/provider"
)

// BasetenOptionsKey is the context key for Baseten-specific options
type BasetenOptionsKey struct{}

// BasetenOptions contains parameters specific to the Baseten Mars-Pro model
type BasetenOptions struct {
	ReferenceAudio    string `json:"reference_audio"`
	ReferenceLanguage string `json:"reference_language"`
}

// BasetenProvider implements the provider.TtsProvider interface for Baseten
type BasetenProvider struct {
	APIKey string
	URL    string
}

// CreateTts is a stub to satisfy the interface.
func (b *BasetenProvider) CreateTts(ctx context.Context, request *cambai.CreateTtsRequestPayload) (*cambai.CreateTtsOut, error) {
	return nil, fmt.Errorf("CreateTts (async) not supported by BasetenProvider, use Tts (stream)")
}

// Tts implements the streaming text-to-speech using Baseten
func (b *BasetenProvider) Tts(ctx context.Context, request *cambai.CreateStreamTtsRequestPayload) (io.Reader, error) {
	// 1. Retrieve Baseten-specific options from Context
	opts, ok := ctx.Value(BasetenOptionsKey{}).(BasetenOptions)
	if !ok || opts.ReferenceAudio == "" || opts.ReferenceLanguage == "" {
		return nil, errors.New("BasetenProvider requires BasetenOptions (ReferenceAudio, ReferenceLanguage) in context")
	}

	// 2. Map Language Enum to String (Simplified mapping)
	// In a real app, you'd map all enums. Here we handle EN-US and fallback.
	langStr := "en-us"
	if request.Language == cambai.CreateStreamTtsRequestPayloadLanguageEnUs {
		langStr = "en-us"
	} else {
		// Fallback or attempt to cast if possible (but the enum is a string type in this SDK version?)
		// Let's check the type. in the file view it says: type CreateStreamTtsRequestPayloadLanguage string
		langStr = strings.ToLower(string(request.Language))
		langStr = strings.ReplaceAll(langStr, "_", "-")
	}

	// 3. Construct Baseten Payload
	payload := map[string]interface{}{
		"text":               request.Text,
		"stream":             true,
		"output_format":      "mp3",
		"language":           langStr,
		"reference_audio":    opts.ReferenceAudio,
		"audio_ref":          opts.ReferenceAudio, // Baseten sometimes checks both
		"reference_language": opts.ReferenceLanguage,
		"apply_ner_nlp":      false,
	}

	// Map optional fields if present
	if request.SpeechModel != nil {
		payload["speech_model"] = string(*request.SpeechModel)
	}
	if request.UserInstructions != nil {
		payload["user_instructions"] = *request.UserInstructions
	}

	// Map Inference Options
	if request.InferenceOptions != nil {
		if request.InferenceOptions.Temperature != nil {
			payload["temperature"] = *request.InferenceOptions.Temperature
		}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 4. Make Request
	req, err := http.NewRequestWithContext(ctx, "POST", b.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Api-Key "+b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Note: We copy to a buffer to match the io.Reader interface cleanly and avoid leaks.
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("baseten error (%d): %s", resp.StatusCode, string(body))
	}

	// 5. Success - Read into buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func main() {
	apiKey := os.Getenv("BASETEN_API_KEY")
	url := os.Getenv("BASETEN_URL")
	if url == "" {
		url = "https://model-5qeryx53.api.baseten.co/environments/production/predict"
	}

	if apiKey == "" {
		fmt.Println("Please set BASETEN_API_KEY")
	}

	// Initialize our custom provider
	var ttsProvider provider.TtsProvider = &BasetenProvider{
		APIKey: apiKey,
		URL:    url,
	}

	fmt.Println("Using Custom Baseten Provider via Go SDK...")

	// Create a standard Camb.ai Request
	req := &cambai.CreateStreamTtsRequestPayload{
		Text:        "Hello from Go Custom Provider with Context options!",
		Language:    cambai.CreateStreamTtsRequestPayloadLanguageEnUs,
		VoiceID:     1, // Mandatory in payload
		SpeechModel: cambai.CreateStreamTtsRequestPayloadSpeechModelMarsPro.Ptr(),
	}

	// Prepare Context with Baseten Options
	dummyAudio := "UklGRi..." // shortened for brevity
	ctx := context.WithValue(context.Background(), BasetenOptionsKey{}, BasetenOptions{
		ReferenceAudio:    dummyAudio,
		ReferenceLanguage: "en-us",
	})

	// Execute
	stream, err := ttsProvider.Tts(ctx, req)
	if err != nil {
		if apiKey == "" {
			fmt.Println("Skipping execution (No API Key).")
			return
		} else {
			panic(err)
		}
	}

	// Save to file
	outFile, err := os.Create("baseten-output.mp3")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, stream)
	if err != nil {
		panic(err)
	}

	fmt.Println("Success! Saved to baseten-output.mp3")
}
