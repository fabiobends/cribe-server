package transcripts

import "time"

// TranscriptStatus represents the valid states of a transcript
type TranscriptStatus string

const (
	TranscriptStatusProcessing TranscriptStatus = "processing"
	TranscriptStatusComplete   TranscriptStatus = "complete"
	TranscriptStatusFailed     TranscriptStatus = "failed"
)

type Transcript struct {
	ID           int        `json:"id"`
	EpisodeID    int        `json:"episode_id"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type TranscriptChunk struct {
	ID           int     `json:"id"`
	TranscriptID int     `json:"transcript_id"`
	Position     int     `json:"position"`
	SpeakerIndex int     `json:"speaker_index"`
	StartTime    float64 `json:"start_time"`
	EndTime      float64 `json:"end_time"`
	Text         string  `json:"text"`
}

type TranscriptSpeaker struct {
	ID           int       `json:"id"`
	TranscriptID int       `json:"transcript_id"`
	SpeakerIndex int       `json:"speaker_index"`
	SpeakerName  string    `json:"speaker_name"`
	InferredAt   time.Time `json:"inferred_at"`
}

type Episode struct {
	ID          int    `json:"id"`
	AudioURL    string `json:"audio_url"`
	Description string `json:"description"`
}

type Chunk struct {
	Position     int     `json:"position"`
	SpeakerIndex int     `json:"speaker_index"`
	Start        float64 `json:"start"`
	End          float64 `json:"end"`
	Text         string  `json:"text"`
}

type Speaker struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}
