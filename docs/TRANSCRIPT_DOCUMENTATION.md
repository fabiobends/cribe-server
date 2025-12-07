# Transcript Streaming API

Real-time podcast transcription with speaker diarization via Server-Sent Events.

## Endpoint

```
GET /transcripts/stream/sse?episode_id={id}
```

### SSE Events

| Event      | Data                                          | Description                                    |
| ---------- | --------------------------------------------- | ---------------------------------------------- |
| `chunk`    | `{position, speaker_index, start, end, text}` | Transcript word with timestamp                 |
| `speaker`  | `{index, name}`                               | Speaker identification (initial + AI-inferred) |
| `complete` | -                                             | Processing finished                            |
| `error`    | `{error}`                                     | Error occurred                                 |

## Architecture

### Flow Diagram

```
┌──────┐  GET /transcripts/stream/sse?episode_id=1
│Client│───────────────────────────────────────────────┐
└──┬───┘                                               │
   │                                           ┌───────▼────────┐
   │◄──────────── SSE: chunk events ───────────┤  HTTP Handler  │
   │◄──────────── SSE: speaker events ─────────┤   (service.go) │
   │◄──────────── SSE: complete event ─────────└───────┬────────┘
   │                                                    │
   │                                         Check DB for cached?
   │                                                    │
   │                                        ┌───────────▼──────────┐
   │                                        │   YES: Stream from   │
   │                                        │   DB (instant)       │
   │                                        └──────────────────────┘
   │                                        ┌──────────────────────┐
   │                                        │   NO: New transcript │
   │                                        └───────┬──────────────┘
   │                                                │
   │                        ┌───────────────────────┼───────────────────────┐
   │                        │                       │                       │
   │                 ┌──────▼─────┐         ┌──────▼─────┐         ┌───────▼──────┐
   │                 │  Goroutine 1│         │ Goroutine 2│         │  Main Thread │
   │                 │Download+Send│         │Read Results│         │  Stream SSE  │
   │                 │ Audio → WSS │         │  WSS → CB  │         │  to Client   │
   │                 └──────┬──────┘         └──────┬─────┘         └──────────────┘
   │                        │                       │
   │                        ▼                       ▼
   │                 ┌─────────────────────────────────┐
   │                 │   Deepgram WebSocket (WSS)      │
   │                 │   - Receives: Binary audio      │
   │                 │   - Sends: JSON transcription   │
   │                 └─────────────────────────────────┘
   │
   │              After streaming completes:
   │              ┌────────────────────────────────┐
   │              │  Background Goroutine:         │
   │              │  1. Save chunks to DB          │
   │              │  2. Infer speaker names (LLM)  │
   │              │  3. Update transcript status   │
   │              └────────────────────────────────┘
```

### WebSocket Goroutines

**Goroutine 1** (Audio Uploader):

- Downloads audio from URL
- Streams binary chunks to Deepgram WebSocket
- Signals completion via `doneCh`

**Goroutine 2** (Transcript Reader):

- Reads JSON messages from WebSocket
- Invokes callback for each word
- Main thread accumulates chunks and streams to client

**Synchronization**:

- Shared `WebSocket conn` (full-duplex)
- `streamCtx` for cancellation
- `errCh` / `doneCh` for error/completion signaling

### Database Schema

```sql
transcripts (id, episode_id, status, error_message, created_at, completed_at)
  ├── transcript_chunks (id, transcript_id, position, speaker_index, start_time, end_time, text)
  └── transcript_speakers (id, transcript_id, speaker_index, speaker_name, inferred_at)
```

**Constraints**:

- `transcripts`: One per episode (unique `episode_id`)
- `transcript_chunks`: Unique `(transcript_id, position)`
- `transcript_speakers`: Unique `(transcript_id, speaker_index)`

## Configuration

```bash
TRANSCRIPTION_API_KEY=<deepgram_key>
TRANSCRIPTION_API_BASE_URL=https://api.deepgram.com/v1

LLM_API_KEY=<openai_key>
LLM_API_BASE_URL=https://api.openai.com/v1
```

## Cost Analysis

| Service            | Frequency            | Cost          | Cache     |
| ------------------ | -------------------- | ------------- | --------- |
| Deepgram           | Once/episode         | ~$0.0125/min  | Permanent |
| OpenAI gpt-4o-mini | Once/speaker/episode | ~$0.0001/call | Permanent |

**Example**: 100-minute podcast with 3 speakers = $1.25 + $0.0003 = **$1.25 first request, $0 cached**
