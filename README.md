# Camb.ai Go SDK

<div id="top" align="center">

   ![Banner](https://github.com/Camb-ai/cambai-python-sdk/raw/main/assets/banner5_720.jpg)
   <h3>
   <a href="https://camb.ai/"> Camb AI Website </a></h3>

</div>

The official Go SDK for interacting with [Camb.ai](https://camb.ai)'s powerful voice and audio generation APIs. Create expressive speech, unique voices, and rich soundscapes with idiomatic Go.

## ✨ Features

- **Dubbing**: End-to-end video dubbing and translation.
- **Expressive Text-to-Speech**: High-fidelity speech synthesis with standard and custom voices.
- **Generative Voices**: Create entirely new voices from text descriptions.
- **Voice Cloning**: Clone voices from audio samples.
- **Soundscapes**: Generate ambient audio and sound effects.

## 📦 Installation

```bash
go get github.com/camb-ai/cambai-go-sdk
```

## 🔑 Authentication

Initialize the client with your API key. You can pass it directly or load it from an environment variable.

```go
import (
    "os"
    "github.com/camb-ai/cambai-go-sdk/client"
    "github.com/camb-ai/cambai-go-sdk/option"
)

func main() {
    c := client.NewClient(
        option.WithAPIKey(os.Getenv("CAMB_API_KEY")),
    )
}
```

## 🚀 Getting Started: Examples

### Supported Models & Sample Rates

| Model Name | Sample Rate | Description |
| :--- | :--- | :--- |
| **mars-pro** | **48kHz** | High-fidelity, professional-grade speech synthesis. Ideal for long-form content and dubbing. |
| **mars-instruct** | **22.05kHz** | Optimized for instruction-following and nuance control. |
| **mars-flash** | **22.05kHz** | Low-latency model optimized for real-time applications and conversational AI. |

### 1. Text-to-Speech (TTS)

Generate and stream speech in real-time. This example uses the high-fidelity `mars-pro` model.

```go
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
```

### 2. End-to-End Dubbing

Dub a video from one language to another and poll for the result.

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/camb-ai/cambai-go-sdk"
    "github.com/camb-ai/cambai-go-sdk/client"
    "github.com/camb-ai/cambai-go-sdk/option"
)

func main() {
    c := client.NewClient(option.WithAPIKey(os.Getenv("CAMB_API_KEY")))

    resp, _ := c.Dub.EndToEndDubbing(
        context.Background(),
        &cambai.EndToEndDubbingRequestPayload{
            VideoURL:        "https://example.com/video.mp4",
            SourceLanguage:  cambai.LanguagesEnUs,
            TargetLanguages: []cambai.Languages{cambai.LanguagesHiIn},
        },
    )

    taskID := *resp.TaskID
    fmt.Printf("Dubbing started. Task ID: %s\n", taskID)

    for {
        time.Sleep(5 * time.Second)
        statusResp, _ := c.Dub.GetEndToEndDubbingStatus(context.Background(), taskID, nil)
        
        if statusResp.Status == cambai.TaskStatusSuccess {
            runID := *statusResp.RunID
            result, _ := c.Dub.GetDubbedRunInfo(context.Background(), &runID, nil)
            
            if result.DubbingResult != nil {
                fmt.Printf("✓ Dubbing successful! Video URL: %s\n", *result.DubbingResult.VideoURL)
                fmt.Printf("  Transcript: %v\n", result.DubbingResult.Transcript)
            }
            break
        } else if statusResp.Status == cambai.TaskStatusError {
            fmt.Println("Dubbing failed.")
            break
        }
    }
}
```

### 3. Text-to-Voice (Generative Voice)

Create a unique new voice from a description.

```go
resp, _ := c.TextToVoice.CreateTextToVoice(
    context.Background(),
    &cambai.CreateTextToVoiceRequestPayload{
        Text:             "Crafting a truly unique and captivating voice.",
        VoiceDescription: "A smooth, rich baritone voice layered with warmth.",
    },
)

taskID := *resp.TaskID
fmt.Printf("Voice generation task created: %s\n", taskID)
// Poll status using c.TextToVoice.GetTextToVoiceStatus(taskID)
```

### 4. Text-to-Audio (Sound Generation)

Generate sound effects from a descriptive prompt.

```go
resp, _ := c.TextToAudio.CreateTextToAudio(
    context.Background(),
    &cambai.CreateTextToAudioRequestPayload{
        Prompt:    "A gentle breeze rustling through autumn leaves.",
        Duration:  cambai.Float64(3.0),
        AudioType: cambai.String("sound"),
    },
)

taskID := *resp.TaskID
fmt.Printf("Sound task created: %s\n", taskID)
// Poll status and get result run_id, then download using c.TextToAudio.GetTextToAudioResult(run_id)
```

## ⚙️ Advanced Usage & Other Features

The Camb AI SDK offers a wide range of capabilities beyond these examples, including:

- Voice Cloning
- Translated TTS
- Audio Dubbing
- Transcription
- And more!

Please refer to the [Official Camb AI API Documentation](https://docs.camb.ai) for a comprehensive list of features and advanced usage patterns.

## 📖 Examples

Check out the `examples/` directory for complete, runnable examples:

- `examples/basic-tts` - Basic text-to-speech example
- `examples/text-to-audio` - Sound generation example
- `examples/dubbing` - Video dubbing workflow
- `examples/baseten-provider` - Using custom providers

## 🔗 Links

- [Documentation](https://docs.camb.ai)
- [GitHub Repository](https://github.com/Camb-ai/cambai-go-sdk)
- [Python SDK](https://github.com/Camb-ai/cambai-python-sdk)

## License

This project is licensed under the MIT License - see the LICENSE file for details
