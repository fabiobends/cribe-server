# Transcript Streaming API

Real-time podcast episode transcription with speaker identification.

## Features

- **Real-time Streaming**: SSE endpoint for live transcript delivery
- **Speaker Diarization**: Automatic speaker detection and identification
- **AI-Powered Speaker Naming**: Infers speaker names from episode context during streaming
- **Persistent Caching**: Saves transcripts to avoid re-processing (cost optimization)
- **Vendor Agnostic**: Easily swap transcription/inference providers

## Architecture

### Endpoint

#### SSE Endpoint (Server-Sent Events)
```
GET /transcripts/stream/sse?episode_id=1
```

**Response Events:**
- `chunk`: Transcript text segments with timestamps and speaker
  ```json
  {
    "position": 0,
    "speaker_index": 0,
    "start": 0.5,
    "end": 2.3,
    "text": "Welcome to the podcast"
  }
  ```
- `speaker`: Speaker identification updates (sent twice per speaker: initial placeholder + inferred name)
  ```json
  {
    "index": 0,
    "name": "John Doe"
  }
  ```
- `complete`: Transcription and speaker inference finished
- `error`: Error occurred
  ```json
  {
    "error": "error message"
  }
  ```

### Data Flow

```
1. Client requests transcript
2. Check DB cache → Stream from DB if exists
3. If new:
   a. Stream from transcription API (Deepgram)
   b. Send chunks to client in real-time
   c. Detect speakers → Send initial names ("Speaker 0")
   d. After chunks complete: Infer real names with AI (OpenAI)
   e. Send updated speaker names to client
   f. Save everything to DB
4. Future requests: Stream from DB (instant, free)
```

### Database Schema

**transcripts**
- Tracks processing status per episode
- Prevents duplicate processing

**transcript_chunks**
- Stores every word with timestamp
- Linked to speaker index

**transcript_speakers**
- Maps speaker index → real name
- Never re-infer speaker names

## Configuration

### Environment Variables

```bash
# Transcription API
TRANSCRIPTION_API_KEY=your_api_key
TRANSCRIPTION_API_BASE_URL=https://api.deepgram.com/v1

# LLM API
LLM_API_KEY=your_api_key
LLM_API_BASE_URL=https://api.openai.com/v1
```

## Cost Optimization

**Transcription API** (Deepgram)
- Called: Once per episode
- Cached: Forever in DB
- Cost: ~$0.0125/minute

**LLM API** (OpenAI gpt-4o-mini)
- Called: Once per unique speaker per episode
- Cached: Forever in DB
- Cost: ~$0.0001/request (2-5 speakers typical)

**Total**: ~$1.25 per 100 minutes of podcast (first time), $0 after caching
