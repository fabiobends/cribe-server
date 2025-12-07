package transcription

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"cribeapp.com/cribe-server/internal/core/logger"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		baseURL       string
		shouldSucceed bool
	}{
		{"success with environment variables", "test-key", "https://api.deepgram.com", true},
		{"missing API key", "", "https://api.deepgram.com", false},
		{"missing base URL", "test-key", "", false},
		{"both missing", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiKey != "" {
				_ = os.Setenv("TRANSCRIPTION_API_KEY", tt.apiKey)
				defer func() { _ = os.Unsetenv("TRANSCRIPTION_API_KEY") }()
			} else {
				_ = os.Unsetenv("TRANSCRIPTION_API_KEY")
			}
			if tt.baseURL != "" {
				_ = os.Setenv("TRANSCRIPTION_API_BASE_URL", tt.baseURL)
				defer func() { _ = os.Unsetenv("TRANSCRIPTION_API_BASE_URL") }()
			} else {
				_ = os.Unsetenv("TRANSCRIPTION_API_BASE_URL")
			}

			client := NewClient()
			if tt.shouldSucceed && (client == nil || client.apiKey != tt.apiKey) {
				t.Fatal("Failed to create client")
			}
			if !tt.shouldSucceed && client != nil {
				t.Error("Expected nil client")
			}
		})
	}
}

func TestStreamAudioURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := StreamResponse{
			Type: "Results",
			Results: Results{
				Channels: []Channel{{Alternatives: []Alternative{{
					Words: []Word{{Word: "test", PunctuatedWord: "test", Start: 0.0, End: 0.5, Speaker: 0}},
				}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		baseURL:    server.URL,
		httpClient: &http.Client{},
		log:        logger.NewServiceLogger("TestClient"),
	}

	callbackCount := 0
	err := client.StreamAudioURL(
		context.Background(),
		"https://example.com/audio.mp3",
		StreamOptions{Model: "nova-3", Language: "en", Diarize: true, Punctuate: true},
		func(response *StreamResponse) error {
			callbackCount++
			return nil
		},
	)

	if err != nil || callbackCount != 2 {
		t.Errorf("Expected no error and 2 callbacks, got err=%v, count=%d", err, callbackCount)
	}
}

func TestStreamAudioURL_RequestCreationError(t *testing.T) {
	client := &Client{
		apiKey:     "test-key",
		baseURL:    ":\n\n\t\tinvalid-url", // Invalid URL to trigger NewRequestWithContext error
		httpClient: &http.Client{},
		log:        logger.NewServiceLogger("TestClient"),
	}

	ctx := context.Background()
	err := client.StreamAudioURL(
		ctx,
		"https://example.com/audio.mp3",
		StreamOptions{Model: "nova-3"},
		func(response *StreamResponse) error { return nil },
	)

	if err == nil {
		t.Fatal("Expected error when baseURL is invalid, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("Expected 'failed to create request' error, got: %v", err)
	}
}

func TestBuildWebSocketURL(t *testing.T) {
	client := &Client{
		baseURL: "https://api.deepgram.com/v1/listen",
		log:     logger.NewServiceLogger("TestClient"),
	}

	wsURL, err := client.buildWebSocketURL(StreamOptions{
		Model: "nova-2", Language: "en", Diarize: true, Punctuate: true,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !strings.HasPrefix(wsURL, "wss://") || !strings.Contains(wsURL, "model=nova-2") {
		t.Errorf("Invalid WebSocket URL: %s", wsURL)
	}
}

func TestStreamAudioToWebSocket(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		audioServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("fake audio data"))
		}))
		defer audioServer.Close()

		var receivedData []byte
		wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, _ := upgrader.Upgrade(w, r, nil)
			defer func() { _ = conn.Close() }()
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil || messageType == websocket.TextMessage {
					return
				}
				receivedData = append(receivedData, message...)
			}
		}))
		defer wsServer.Close()

		wsURL := "ws://" + strings.TrimPrefix(wsServer.URL, "http://")
		conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
		defer func() { _ = conn.Close() }()

		client := &Client{httpClient: &http.Client{}, log: logger.NewServiceLogger("TestClient")}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		errCh, doneCh := make(chan error, 1), make(chan struct{}, 1)
		go client.streamAudioToWebSocket(ctx, conn, audioServer.URL, errCh, doneCh)

		select {
		case <-doneCh:
			if len(receivedData) == 0 {
				t.Error("Expected to receive audio data")
			}
		case err := <-errCh:
			t.Fatalf("Expected no error, got %v", err)
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		audioServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer audioServer.Close()

		wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := websocket.Upgrader{}
			conn, _ := upgrader.Upgrade(w, r, nil)
			defer func() { _ = conn.Close() }()
		}))
		defer wsServer.Close()

		wsURL := "ws://" + strings.TrimPrefix(wsServer.URL, "http://")
		conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
		defer func() { _ = conn.Close() }()

		client := &Client{httpClient: &http.Client{}, log: logger.NewServiceLogger("TestClient")}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		errCh, doneCh := make(chan error, 1), make(chan struct{}, 1)
		go client.streamAudioToWebSocket(ctx, conn, audioServer.URL, errCh, doneCh)

		select {
		case err := <-errCh:
			if err == nil || !strings.Contains(err.Error(), "audio download failed") {
				t.Errorf("Expected 'audio download failed' error, got: %v", err)
			}
		case <-doneCh:
			t.Fatal("Expected error, got success")
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	})
}

func TestReadTranscriptionResults(t *testing.T) {
	tests := []struct {
		name          string
		messages      []any
		expectedCalls int
		description   string
	}{
		{
			name: "filters interim results",
			messages: []any{
				StreamResponse{Type: "Results", IsFinal: false, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "interim"}}}}}},
				StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "final"}}}}}},
			},
			expectedCalls: 1,
			description:   "should only process final results",
		},
		{
			name: "skips invalid JSON",
			messages: []any{
				"{invalid json}",
				StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "test"}}}}}},
			},
			expectedCalls: 1,
			description:   "invalid JSON should be skipped",
		},
		{
			name: "skips non-Results type",
			messages: []any{
				StreamResponse{Type: "Metadata", Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "ignored"}}}}}},
				StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "test"}}}}}},
			},
			expectedCalls: 1,
			description:   "non-Results type should be skipped",
		},
		{
			name: "skips empty alternatives",
			messages: []any{
				StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{}}},
				StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "test"}}}}}},
			},
			expectedCalls: 1,
			description:   "message with no alternatives should be skipped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}
				conn, _ := upgrader.Upgrade(w, r, nil)
				defer func() { _ = conn.Close() }()

				for _, msg := range tt.messages {
					switch v := msg.(type) {
					case string:
						_ = conn.WriteMessage(websocket.TextMessage, []byte(v))
					case StreamResponse:
						jsonData, _ := json.Marshal(v)
						_ = conn.WriteMessage(websocket.TextMessage, jsonData)
					}
				}
				_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			}))
			defer wsServer.Close()

			wsURL := "ws://" + strings.TrimPrefix(wsServer.URL, "http://")
			conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)

			client := &Client{log: logger.NewServiceLogger("TestClient")}
			errCh, doneCh := make(chan error, 1), make(chan struct{}, 1)

			callbackCount := 0
			callback := func(resp *StreamResponse) error {
				callbackCount++
				return nil
			}

			go client.readTranscriptionResults(context.Background(), conn, callback, errCh, doneCh)

			select {
			case <-doneCh:
				if callbackCount != tt.expectedCalls {
					t.Errorf("%s: expected %d callbacks, got %d", tt.description, tt.expectedCalls, callbackCount)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Test timed out")
			}
		})
	}
}

func TestStreamAudioURLWebSocket(t *testing.T) {
	setupServers := func(audioHandler, wsHandler http.HandlerFunc) (audioURL, wsURL string, cleanup func()) {
		audioServer := httptest.NewServer(audioHandler)
		wsServer := httptest.NewServer(wsHandler)
		return audioServer.URL, strings.Replace(wsServer.URL, "http://", "ws://", 1), func() {
			audioServer.Close()
			wsServer.Close()
		}
	}

	t.Run("success", func(t *testing.T) {
		audioURL, wsURL, cleanup := setupServers(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "audio data")
			},
			func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}
				conn, _ := upgrader.Upgrade(w, r, nil)
				defer func() { _ = conn.Close() }()
				for {
					messageType, message, err := conn.ReadMessage()
					if err != nil {
						return
					}
					if messageType == websocket.TextMessage && strings.Contains(string(message), "CloseStream") {
						response := StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "test"}}}}}}
						jsonData, _ := json.Marshal(response)
						_ = conn.WriteMessage(websocket.TextMessage, jsonData)
						_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						return
					}
				}
			},
		)
		defer cleanup()

		client := &Client{apiKey: "test-key", baseURL: wsURL, httpClient: http.DefaultClient, log: logger.NewServiceLogger("TestClient")}

		called := false
		err := client.StreamAudioURLWebSocket(context.Background(), audioURL, StreamOptions{Model: "nova-2", Language: "en"}, func(resp *StreamResponse) error {
			called = true
			return nil
		})

		if err != nil || !called {
			t.Errorf("Expected success and callback, got err=%v, called=%v", err, called)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		audioURL, wsURL, cleanup := setupServers(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				for range 100 {
					_, _ = io.WriteString(w, "chunk ")
				}
			},
			func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}
				conn, _ := upgrader.Upgrade(w, r, nil)
				defer func() { _ = conn.Close() }()
				for {
					if _, _, err := conn.ReadMessage(); err != nil {
						return
					}
				}
			},
		)
		defer cleanup()

		client := &Client{apiKey: "test-key", baseURL: wsURL, httpClient: http.DefaultClient, log: logger.NewServiceLogger("TestClient")}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			cancel()
		}()

		err := client.StreamAudioURLWebSocket(ctx, audioURL, StreamOptions{Model: "nova-2"}, func(resp *StreamResponse) error { return nil })
		if err == nil {
			t.Error("Expected context cancellation error")
		}
	})

	t.Run("callback error", func(t *testing.T) {
		audioURL, wsURL, cleanup := setupServers(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "audio")
			},
			func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}
				conn, _ := upgrader.Upgrade(w, r, nil)
				defer func() { _ = conn.Close() }()
				for {
					messageType, message, err := conn.ReadMessage()
					if err != nil {
						return
					}
					if messageType == websocket.TextMessage && strings.Contains(string(message), "CloseStream") {
						response := StreamResponse{Type: "Results", IsFinal: true, Channel: Channel{Alternatives: []Alternative{{Words: []Word{{Word: "test"}}}}}}
						jsonData, _ := json.Marshal(response)
						_ = conn.WriteMessage(websocket.TextMessage, jsonData)
						return
					}
				}
			},
		)
		defer cleanup()

		client := &Client{apiKey: "test-key", baseURL: wsURL, httpClient: http.DefaultClient, log: logger.NewServiceLogger("TestClient")}

		err := client.StreamAudioURLWebSocket(context.Background(), audioURL, StreamOptions{Model: "nova-2"}, func(resp *StreamResponse) error {
			return fmt.Errorf("callback error")
		})

		if err == nil {
			t.Error("Expected callback error to be propagated")
		}
	})
}
