package cards

import (
	"errors"
	"net/http"

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
	router.GET("/cards/:id", h.getCardSet)
	router.POST("/cards/:id", h.startSession)
	router.POST("/cards/:id/:session_id", h.nextCard)
	router.GET("/cards/:id/:session_id", h.getSessionProgress)
	router.DELETE("/cards/:id/:session_id", h.skipCard)
}

func (h Handler) createCardSet(ctx *gin.Context) {
	var request []schema.CardData
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(request) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cards payload must not be empty"})
		return
	}
	id, err := h.store.CreateCardSet(request)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, schema.CreateCardSetResponse{ID: id})
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
	sessionID, err := h.store.CreateSession(ctx.Param("id"))
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
