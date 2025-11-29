package transcription

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/core/logger"
)

// Client handles audio transcription API interactions
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	log        *logger.ContextualLogger
}

// StreamOptions configures the transcription request
type StreamOptions struct {
	Model      string // Model to use (e.g., "nova-3")
	Language   string // Language code (e.g., "en")
	Diarize    bool   // Enable speaker detection
	Punctuate  bool   // Enable punctuation
	Utterances bool   // Enable utterance grouping
}

// Word represents a single word in the transcript
type Word struct {
	Word              string  `json:"word"`
	PunctuatedWord    string  `json:"punctuated_word"`
	Start             float64 `json:"start"`
	End               float64 `json:"end"`
	Confidence        float64 `json:"confidence"`
	Speaker           int     `json:"speaker"`
	SpeakerConfidence float64 `json:"speaker_confidence,omitempty"`
}

// Alternative represents transcript alternatives
type Alternative struct {
	Transcript string  `json:"transcript"`
	Confidence float64 `json:"confidence"`
	Words      []Word  `json:"words"`
}

// Channel represents a channel in the response
type Channel struct {
	Alternatives []Alternative `json:"alternatives"`
}

// Metadata contains response metadata
type Metadata struct {
	RequestID string  `json:"request_id"`
	Duration  float64 `json:"duration"`
}

// Results wraps the channels array
type Results struct {
	Channels []Channel `json:"channels"`
}

// StreamResponse represents a transcription streaming response
type StreamResponse struct {
	Metadata Metadata `json:"metadata"`
	Results  Results  `json:"results"`
	IsFinal  bool     `json:"is_final"`
	Type     string   `json:"type"` // "Results", "UtteranceEnd", "Metadata"
}

// StreamCallback is called for each response chunk
type StreamCallback func(response *StreamResponse) error
