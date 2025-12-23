package quizzes

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/utils"
)

// generateQuestionsWithLLM calls the LLM to generate questions
func (s *QuizService) generateQuestionsWithLLM(transcriptText string) ([]LLMQuestion, error) {
	// Create LLM request
	reqBody := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: GenerateQuestionsSystemPrompt},
			{Role: "user", Content: GenerateQuestionsUserPrompt(transcriptText)},
		},
		MaxTokens: 2000,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Call LLM
	response, err := s.llmClient.Chat(ctx, reqBody)
	if err != nil {
		s.logger.Error("LLM request failed", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse LLM response
	content := response.Choices[0].Message.Content

	// Clean up the response (remove markdown code blocks if present)
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	llmResponse, err := utils.DecodeResponse[LLMQuestionsResponse](content)
	if err != nil {
		s.logger.Error("Failed to parse LLM response", map[string]any{
			"error":   err.Error(),
			"content": content,
		})
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return llmResponse.Questions, nil
}

// evaluateOpenEndedAnswer uses LLM to evaluate open-ended answers
func (s *QuizService) evaluateOpenEndedAnswer(question Question, userAnswer string) (bool, string, error) {
	reqBody := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: EvaluateOpenEndedSystemPrompt},
			{Role: "user", Content: EvaluateOpenEndedUserPrompt(question.QuestionText, userAnswer, "")},
		},
		MaxTokens: 300,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call LLM
	response, err := s.llmClient.Chat(ctx, reqBody)
	if err != nil {
		return false, "Unable to evaluate answer at this time", err
	}

	if len(response.Choices) == 0 {
		return false, "Unable to evaluate answer at this time", fmt.Errorf("no response from LLM")
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	evaluation, err := utils.DecodeResponse[LLMEvaluationResponse](content)

	if err != nil {
		s.logger.Error("Failed to parse LLM evaluation", map[string]any{
			"error":   err.Error(),
			"content": content,
		})
		return false, "Unable to evaluate answer at this time", err
	}

	return evaluation.IsCorrect, evaluation.Feedback, nil
}

// generateFeedbackWithLLM generates personalized feedback using LLM
func (s *QuizService) generateFeedbackWithLLM(question Question, userAnswer string, isCorrect bool) string {
	// If LLM client is not available (nil interface or nil pointer), return basic feedback
	if s.llmClient == nil || reflect.ValueOf(s.llmClient).IsNil() {
		if isCorrect {
			return "Correct answer!"
		}
		return "Incorrect answer!"
	}

	transcriptText, _ := s.getTranscriptText(question.EpisodeID)

	reqBody := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: GenerateFeedbackSystemPrompt},
			{Role: "user", Content: GenerateFeedbackUserPrompt(question.QuestionText, userAnswer, isCorrect, transcriptText)},
		},
		MaxTokens: 200,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Call LLM
	response, err := s.llmClient.Chat(ctx, reqBody)

	if err != nil || len(response.Choices) == 0 {
		if isCorrect {
			return "Correct answer!"
		}
		return "Incorrect answer!"
	}

	return strings.TrimSpace(response.Choices[0].Message.Content)
}
