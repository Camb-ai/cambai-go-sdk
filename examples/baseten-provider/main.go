package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	cambai "github.com/camb-ai/cambai-go-sdk"
	"github.com/camb-ai/cambai-go-sdk/provider"
)

// BasetenProvider implements provider.TtsProvider for the Baseten Mars8-Flash model.
// API reference: https://www.baseten.co/library/mars8-flash/
type BasetenProvider struct {
	APIKey            string
	URL               string
	ReferenceAudio    string // Public URL or base64-encoded audio file
	ReferenceLanguage string // ISO locale of the reference audio, e.g. "en-us"
}

// CreateTts is a stub — Baseten does not support async TTS.
func (b *BasetenProvider) CreateTts(ctx context.Context, request *cambai.CreateTtsRequestPayload) (*cambai.CreateTtsOut, error) {
	return nil, fmt.Errorf("Baseten custom hosting provider does not support async CreateTts; use Tts (streaming) instead")
}

// Tts calls the Baseten Mars8-Flash endpoint and returns the audio as an io.Reader.
func (b *BasetenProvider) Tts(ctx context.Context, request *cambai.CreateStreamTtsRequestPayload) (io.Reader, error) {
	// Normalise language: SDK enum is a string type, ensure lowercase ISO format.
	langStr := strings.ToLower(strings.ReplaceAll(string(request.Language), "_", "-"))

	// Build the Mars8-Flash payload.
	// Docs: https://www.baseten.co/library/mars8-flash/
	payload := map[string]interface{}{
		"text":               request.Text,
		"language":           langStr,
		"output_duration":    nil, // null = model infers optimal duration
		"reference_audio":    b.ReferenceAudio,
		"reference_language": b.ReferenceLanguage,
		"output_format":      "flac", // flac is the default; wav also supported
		"apply_ner_nlp":      false,  // disable NER (faster; pass pronunciation_dictionary instead)
	}

	// Optional: temperature (default 1.0 — keep unless you need determinism)
	// Optional: cfg_weight (default 4.2 — controls adherence to reference style)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Api-Key "+b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("baseten error (%d): %s", resp.StatusCode, string(body))
	}

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &buf, nil
}

func main() {
	cambApiKey := os.Getenv("CAMB_API_KEY")
	basetenApiKey := os.Getenv("BASETEN_API_KEY")
	basetenUrl := os.Getenv("BASETEN_URL")
	referenceAudio := os.Getenv("BASETEN_REFERENCE_AUDIO")
	referenceLanguage := os.Getenv("BASETEN_REFERENCE_LANGUAGE")

	// Loud fail for missing required env vars
	missing := []string{}
	if cambApiKey == "" {
		missing = append(missing, "CAMB_API_KEY")
	}
	if basetenApiKey == "" {
		missing = append(missing, "BASETEN_API_KEY")
	}
	if basetenUrl == "" {
		missing = append(missing, "BASETEN_URL (your Baseten model prediction endpoint)")
	}
	if referenceAudio == "" {
		missing = append(missing, "BASETEN_REFERENCE_AUDIO (public URL or base64-encoded audio file)")
	}

	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "Error: Missing required environment variables:")
		for _, v := range missing {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}
		os.Exit(1)
	}

	// Default reference language to en-us if not set
	if referenceLanguage == "" {
		referenceLanguage = "en-us"
	}

	// Initialise the Baseten custom hosting provider
	var ttsProvider provider.TtsProvider = &BasetenProvider{
		APIKey:            basetenApiKey,
		URL:               basetenUrl,
		ReferenceAudio:    referenceAudio,
		ReferenceLanguage: referenceLanguage,
	}

	fmt.Println("Generating speech via Baseten Mars8-Flash custom hosting provider...")

	req := &cambai.CreateStreamTtsRequestPayload{
		Text:     "Hello. This is speech generated via a Baseten Mars8-Flash custom hosting provider.",
		Language: cambai.CreateStreamTtsRequestPayloadLanguageEnUs,
	}

	stream, err := ttsProvider.Tts(context.Background(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputFile := "baseten_output.flac"
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, stream); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing audio: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Success! Audio saved to %s\n", outputFile)
}
