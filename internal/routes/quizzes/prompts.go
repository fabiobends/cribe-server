package quizzes

import "fmt"

var GenerateQuestionsSystemPrompt = fmt.Sprintf(`You are an expert at creating educational quiz questions from podcast transcripts. Generate a mix of multiple choice, true/false, and open-ended questions.

Rules:
- Generate exactly %d questions total
- %d multiple choice (%d options each), %d true/false, and %d open-ended question
- Questions should test understanding of key concepts, not just recall
- For multiple choice: exactly %d options, only one correct
- For true/false: exactly 2 options ("True" and "False")
- For open-ended: no options needed
- Questions should be clear and unambiguous

Return ONLY a valid JSON object with this exact structure:
{
  "questions": [
    {
      "question_text": "What was the main topic discussed?",
      "type": "multiple_choice",
      "options": [
        {"text": "Option A", "is_correct": false},
        {"text": "Option B", "is_correct": true},
        {"text": "Option C", "is_correct": false},
        {"text": "Option D", "is_correct": false}
      ]
    },
    {
      "question_text": "The speaker mentioned X",
      "type": "true_false",
      "options": [
        {"text": "True", "is_correct": true},
        {"text": "False", "is_correct": false}
      ]
    },
    {
      "question_text": "Explain the main concept discussed.",
      "type": "open_ended",
			"options": []
    }
  ]
}

Do not include any markdown formatting, code blocks, or explanatory text. Return only the raw JSON object.`,
	TotalQuestionsPerEpisode, MultipleChoiceCount, MultipleChoiceOptions,
	TrueFalseCount, OpenEndedCount, MultipleChoiceOptions)

func GenerateQuestionsUserPrompt(transcriptText string) string {
	return fmt.Sprintf("Generate quiz questions from this podcast transcript:\n\n%s", transcriptText)
}

var EvaluateOpenEndedSystemPrompt = `You are evaluating a user's answer to an open-ended question. Determine if the answer demonstrates understanding of the key concepts.

Return a JSON object with this exact structure:
{
  "is_correct": true,
  "feedback": "Your explanation here..."
}

Be encouraging but honest. If the answer is partially correct, set is_correct to true if they grasp the main concepts.`

func EvaluateOpenEndedUserPrompt(questionText, userAnswer, correctAnswer string) string {
	if correctAnswer != "" {
		return fmt.Sprintf(`Question: %s

Expected answer: %s

User's answer: %s

Evaluate if the user's answer demonstrates understanding of the topic and aligns with the expected answer. Return only the JSON object.`,
			questionText, correctAnswer, userAnswer)
	}
	return fmt.Sprintf(`Question: %s

User's answer: %s

Evaluate if this answer demonstrates understanding of the topic. Return only the JSON object.`,
		questionText, userAnswer)
}

var GenerateFeedbackSystemPrompt = `You are a helpful tutor providing personalized feedback on quiz answers about podcast content.

CRITICAL: Your feedback MUST be personalized and specific to the question and podcast content. Never give generic responses.

For correct answers:
- Acknowledge what specific concept from the podcast they understood
- Reference concrete details from the podcast that relate to their answer
- Be encouraging but specific (1-2 sentences)
- Example: "Exactly! You correctly identified that the speaker mentioned X when discussing Y."

For incorrect answers:
- Be encouraging but point to the specific concept they missed
- Reference what was actually discussed in the podcast
- Help them understand the correct concept with podcast details
- Keep it constructive (1-2 sentences)
- Example: "Not quite. In the episode, the speaker actually explained that X happens because of Y."

If podcast context is limited, still personalize by referencing the question topic.

Return ONLY the feedback text, no JSON or extra formatting.`

func GenerateFeedbackUserPrompt(questionText, userAnswer string, isCorrect bool, transcriptText string) string {
	status := "correct"
	if !isCorrect {
		status = "incorrect"
	}

	// Truncate transcript to avoid token limits (use first 2000 chars)
	truncatedTranscript := transcriptText
	if len(transcriptText) > 2000 {
		truncatedTranscript = transcriptText[:2000] + "..."
	}

	contextNote := ""
	if transcriptText == "" {
		contextNote = "\nNote: Limited podcast context available. Focus feedback on the question topic."
	}

	return fmt.Sprintf(`Question: %s
User's answer: %s
This answer is: %s

Podcast context:
%s%s

Generate PERSONALIZED, SPECIFIC feedback that references the podcast content or question topic. Do NOT use generic phrases like "you understood the concept well" or "review the episode".`,
		questionText, userAnswer, status, truncatedTranscript, contextNote)
}
