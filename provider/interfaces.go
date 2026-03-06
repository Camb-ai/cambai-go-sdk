package provider

import (
	"context"
	"io"

	cambai "github.com/camb-ai/cambai-go-sdk"
)

// TtsProvider defines the interface for Text-to-Speech operations.
// This allows for swapping the default Camb.ai implementation with custom hosting providers (e.g., Baseten, Vertex AI).
type TtsProvider interface {
	CreateTts(ctx context.Context, request *cambai.CreateTtsRequestPayload) (*cambai.CreateTtsOut, error)
	Tts(ctx context.Context, request *cambai.CreateStreamTtsRequestPayload) (io.Reader, error)
}
