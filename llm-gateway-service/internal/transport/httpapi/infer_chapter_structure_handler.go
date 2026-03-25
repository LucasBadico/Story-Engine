package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/story-engine/llm-gateway-service/internal/application/inference"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

type InferChapterStructureHandler struct {
	useCase *inference.InferChapterStructureUseCase
	logger  *logger.Logger
}

func NewInferChapterStructureHandler(useCase *inference.InferChapterStructureUseCase, logger *logger.Logger) *InferChapterStructureHandler {
	return &InferChapterStructureHandler{
		useCase: useCase,
		logger:  logger,
	}
}

type inferChapterStructureRequest struct {
	Scenes []inference.SceneInput `json:"scenes"`
	Prose  string                 `json:"prose"`
}

func (h *InferChapterStructureHandler) Infer(w http.ResponseWriter, r *http.Request) {
	var req inferChapterStructureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	output, err := h.useCase.Execute(r.Context(), inference.InferChapterStructureInput{
		Scenes: req.Scenes,
		Prose:  req.Prose,
	})
	if err != nil {
		h.logger.Error(fmt.Sprintf("chapter structure inference failed: %v", err))
		writeError(w, http.StatusInternalServerError, "chapter structure inference failed")
		return
	}

	writeJSON(w, http.StatusOK, output)
}
