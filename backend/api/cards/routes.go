package cards

import (
	"encoding/base64"
	"errors"
	"html"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
	"github.com/pxc1984/flashcards-trainer/backend/store"
	"github.com/pxc1984/flashcards-trainer/backend/store/interfaces"
)

type Handler struct {
	store interfaces.StoreBase
}

func RegisterRoutes(router gin.IRoutes, store interfaces.StoreBase) {
	h := Handler{store: store}
	router.POST("/cards", h.createCardSet)
	router.POST("/images", h.uploadImage)
	router.GET("/cards/:id", h.getCardSet)
	router.POST("/cards/:id", h.startSession)
	router.POST("/cards/:id/:session_id", h.nextCard)
	router.GET("/cards/:id/:session_id", h.getSessionProgress)
	router.DELETE("/cards/:id/:session_id", h.skipCard)
}

func (h Handler) createCardSet(ctx *gin.Context) {
	var request schema.CreateCardSetRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	request.Title = sanitizePlainText(request.Title)
	request.Description = sanitizePlainText(request.Description)
	request.Author = sanitizePlainText(request.Author)
	if len(request.Cards) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cards payload must not be empty"})
		return
	}
	for _, card := range request.Cards {
		if len(card.QuestionImages) > 5 || len(card.AnswerImages) > 5 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "each side of a card supports at most 5 images"})
			return
		}
	}
	if len(request.Title) > 120 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "title must be at most 120 characters"})
		return
	}
	id, err := h.store.CreateCardSet(request, ctx.ClientIP())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, schema.CreateCardSetResponse{ID: id})
}

func (h Handler) uploadImage(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uploaded file"})
		return
	}
	if len(content) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "uploaded file is empty"})
		return
	}

	mimeType := http.DetectContentType(content)
	if !strings.HasPrefix(mimeType, "image/") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "uploaded file must be an image"})
		return
	}

	image, err := h.store.CreateUploadedImage(mimeType, base64.StdEncoding.EncodeToString(content), ctx.ClientIP())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, schema.UploadImageResponse{Image: *image})
}

func sanitizePlainText(value string) string {
	return strings.TrimSpace(html.EscapeString(value))
}

func (h Handler) getCardSet(ctx *gin.Context) {
	cards, err := h.store.GetCardSet(ctx.Param("id"))
	if err != nil {
		ctx.JSON(statusForError(err), gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, cards)
}

func (h Handler) startSession(ctx *gin.Context) {
	sessionID, err := h.store.CreateSession(ctx.Param("id"), ctx.ClientIP())
	if err != nil {
		ctx.JSON(statusForError(err), gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, schema.StartSessionResponse{SessionID: sessionID})
}

func (h Handler) nextCard(ctx *gin.Context) {
	_, err := h.store.AdvanceSession(ctx.Param("id"), ctx.Param("session_id"))
	if err != nil {
		ctx.JSON(statusForError(err), gin.H{"error": err.Error()})
		return
	}
	ctx.Redirect(http.StatusSeeOther, "/api/v1/cards/"+ctx.Param("id")+"/"+ctx.Param("session_id"))
}

func (h Handler) getSessionProgress(ctx *gin.Context) {
	progress, err := h.store.GetSessionProgress(ctx.Param("id"), ctx.Param("session_id"))
	if err != nil {
		ctx.JSON(statusForError(err), gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, progress)
}

func (h Handler) skipCard(ctx *gin.Context) {
	_, err := h.store.SkipSessionCard(ctx.Param("id"), ctx.Param("session_id"))
	if err != nil {
		ctx.JSON(statusForError(err), gin.H{"error": err.Error()})
		return
	}
	ctx.Redirect(http.StatusSeeOther, "/api/v1/cards/"+ctx.Param("id")+"/"+ctx.Param("session_id"))
}

func statusForError(err error) int {
	if errors.Is(err, store.ErrCardSetNotFound) || errors.Is(err, store.ErrSessionNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
