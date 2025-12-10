package transcripts

import "fmt"

// InferSpeakerNameSystemPrompt is the system prompt for speaker name inference
var InferSpeakerNameSystemPrompt = "You are an expert at identifying speakers in podcast transcripts. Look for explicit name mentions in the text (e.g., 'this is John', 'I'm Sarah', 'talking with Mike'). Return ONLY the person's name."

// InferSpeakerNameUserPrompt generates the user prompt for speaker name inference
func InferSpeakerNameUserPrompt(episodeDescription string, speakerIndex int, chunksText string) string {
	return fmt.Sprintf(`Episode description:
%s

Transcript excerpt with context around speaker %d:
%s

Instructions:
- Look for the speaker's name mentioned BEFORE they speak (introductions)
- Look for the speaker's name mentioned WHILE they speak (self-introduction)
- Look for the speaker's name mentioned AFTER they speak (references)
- Common patterns: "I'm [name]", "this is [name]", "with [name]", "[name] said"

Who is speaker %d? Return only their full name (e.g., "John Smith"). If uncertain, return "Speaker %d".`,
		episodeDescription, speakerIndex, chunksText, speakerIndex, speakerIndex)
}
