package quizzes

import (
	"net/http"
	"strconv"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/middlewares"
	"cribeapp.com/cribe-server/internal/utils"
)

type QuizHandler struct {
	service QuizService
	logger  *logger.ContextualLogger
}

func NewQuizHandler(service *QuizService) *QuizHandler {
	return &QuizHandler{service: *service, logger: logger.NewHandlerLogger("QuizHandler")}
}

func (h *QuizHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, _ := r.Context().Value(middlewares.UserIDContextKey).(int)

	// Extract path after /quizzes
	path := strings.TrimPrefix(r.URL.Path, "/quizzes")
	path = strings.Trim(path, "/")

	h.logger.Info("Quiz request path", map[string]any{
		"method": r.Method,
		"path":   path,
	})

	// Handle root /quizzes path
	if path == "" {
		h.logger.Debug("Handling sessions")
		h.handleSessions(w, r, userID)
		return
	}

	parts := strings.Split(path, "/")

	h.logger.Info("Handling quiz session path", map[string]any{
		"parts": parts,
	})

	sessionID, err := strconv.Atoi(parts[0])
	if err != nil {
		// Invalid session ID format
		utils.NotFound(w, r)
		return
	}

	// Route based on sub-path
	if len(parts) == 1 {
		// /quizzes/:session_id
		h.handleSessionWithDetailsByID(w, r, sessionID)
	} else {
		switch parts[1] {
		case "answers":
			// /quizzes/:session_id/answers
			h.handleSessionAnswers(w, r, sessionID, userID)
		case "status":
			// /quizzes/:session_id/status
			h.handleSessionStatus(w, r, sessionID, userID)
		default:
			utils.NotFound(w, r)
		}
	}
}

// handleSessions handles both GET and POST requests to /quizzes.
func (h *QuizHandler) handleSessions(w http.ResponseWriter, r *http.Request, userID int) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetSessionsWithDetails(w, r, userID)
	case http.MethodPost:
		h.handleGetOrCreateSessionWithDetails(w, r, userID)
	default:
		utils.NotAllowed(w)
	}
}

// GET /quizzes - Get all sessions with details for user
func (h *QuizHandler) handleGetSessionsWithDetails(w http.ResponseWriter, r *http.Request, userID int) {
	sessions, errResp := h.service.GetSessionsWithDetailsByUserID(userID)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}

	utils.EncodeResponse(w, http.StatusOK, sessions)
}

// POST /quizzes - Get existing session or create a new for user
func (h *QuizHandler) handleGetOrCreateSessionWithDetails(w http.ResponseWriter, r *http.Request, userID int) {
	req, errResp := utils.DecodeBody[GetOrCreateSessionRequest](r)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, errResp)
		return
	}

	if err := req.Validate(); err != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, err)
		return
	}

	session, errResp := h.service.GetOrCreateSessionWithDetails(userID, req.EpisodeID)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}

	utils.EncodeResponse(w, http.StatusOK, session)
}

// GET /quizzes/:session_id - Get session by ID
func (h *QuizHandler) handleSessionWithDetailsByID(w http.ResponseWriter, r *http.Request, sessionID int) {
	if r.Method != http.MethodGet {
		utils.NotAllowed(w)
		return
	}

	session, errResp := h.service.GetSessionWithDetailsByID(sessionID)
	if errResp != nil {
		// Return 404 for not found errors
		if errResp.Message == errors.DatabaseNotFound {
			utils.EncodeResponse(w, http.StatusNotFound, errResp)
			return
		}
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}

	utils.EncodeResponse(w, http.StatusOK, session)
}

// POST /quizzes/:session_id/answers - Submit an answer
func (h *QuizHandler) handleSessionAnswers(w http.ResponseWriter, r *http.Request, sessionID, userID int) {
	switch r.Method {
	case http.MethodPost:
		h.handleSubmitAnswer(w, r, sessionID, userID)
	default:
		utils.NotAllowed(w)
	}
}

func (h *QuizHandler) handleSubmitAnswer(w http.ResponseWriter, r *http.Request, sessionID, userID int) {
	req, errResp := utils.DecodeBody[SubmitAnswerRequest](r)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, errResp)
		return
	}

	if err := req.Validate(); err != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, err)
		return
	}

	answer, errResp := h.service.SubmitAnswer(sessionID, userID, req.QuestionID, req)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}

	utils.EncodeResponse(w, http.StatusCreated, answer)
}

// PATCH /quizzes/:session_id/status - Update session status
func (h *QuizHandler) handleSessionStatus(w http.ResponseWriter, r *http.Request, sessionID, userID int) {
	if r.Method != http.MethodPatch {
		utils.NotAllowed(w)
		return
	}

	req, errResp := utils.DecodeBody[UpdateSessionStatusRequest](r)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, errResp)
		return
	}

	if err := req.Validate(); err != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, err)
		return
	}

	session, errResp := h.service.UpdateSessionStatus(sessionID, userID, req.Status)
	if errResp != nil {
		// Return appropriate status codes based on error type
		if errResp.Message == errors.DatabaseNotFound {
			utils.EncodeResponse(w, http.StatusNotFound, errResp)
			return
		}
		if errResp.Message == errors.Unauthorized {
			utils.EncodeResponse(w, http.StatusForbidden, errResp)
			return
		}
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}

	utils.EncodeResponse(w, http.StatusOK, session)
}
