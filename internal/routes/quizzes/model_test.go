package quizzes

import "testing"

func TestGenerateQuestionsRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		episodeID int
		wantErr   bool
	}{
		{
			name:      "valid episode ID",
			episodeID: 1,
			wantErr:   false,
		},
		{
			name:      "zero episode ID",
			episodeID: 0,
			wantErr:   true,
		},
		{
			name:      "negative episode ID",
			episodeID: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := GenerateQuestionsRequest{EpisodeID: tt.episodeID}
			err := req.Validate()

			if tt.wantErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err.Details)
			}
		})
	}
}

func TestUpdateSessionStatusRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		status  SessionStatus
		wantErr bool
	}{
		{
			name:    "valid completed status",
			status:  Completed,
			wantErr: false,
		},
		{
			name:    "valid abandoned status",
			status:  Abandoned,
			wantErr: false,
		},
		{
			name:    "invalid in_progress status",
			status:  InProgress,
			wantErr: true,
		},
		{
			name:    "invalid empty status",
			status:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := UpdateSessionStatusRequest{Status: tt.status}
			err := req.Validate()

			if tt.wantErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err.Details)
			}
		})
	}
}
