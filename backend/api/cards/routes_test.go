package cards

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pxc1984/flashcards-trainer/backend/domain/schema"
	"github.com/pxc1984/flashcards-trainer/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardsAPIFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeObj := store.NewMemoryStore()
	require.NoError(t, storeObj.Init("admin"))

	router := gin.New()
	apiGroup := router.Group("/api/v1")
	RegisterRoutes(apiGroup, storeObj)

	pngFile := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0x99, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92, 0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82}
	uploadResp := performMultipartRequest(router, http.MethodPost, "/api/v1/images", "file", "pixel.png", pngFile)
	require.Equal(t, http.StatusCreated, uploadResp.Code)

	var uploaded schema.UploadImageResponse
	require.NoError(t, json.Unmarshal(uploadResp.Body.Bytes(), &uploaded))

	createBody := mustJSON(t, schema.CreateCardSetRequest{
		CardSetMetadata: schema.CardSetMetadata{Title: "<b>Title</b>", Description: "desc", Author: "<script>alert(1)</script>"},
		Cards:           []schema.CardData{{Question: "q1", Answer: "a1", Remarks: "r1", QuestionImages: []schema.CardImage{uploaded.Image}}, {Question: "q2", Answer: "a2", Remarks: "r2"}},
	})
	createResp := performRequest(router, http.MethodPost, "/api/v1/cards", createBody)
	require.Equal(t, http.StatusCreated, createResp.Code)

	var created schema.CreateCardSetResponse
	require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
	assert.Len(t, created.ID, 8)

	getSetResp := performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID, nil)
	require.Equal(t, http.StatusOK, getSetResp.Code)

	var cardsResp schema.CardSetResponse
	require.NoError(t, json.Unmarshal(getSetResp.Body.Bytes(), &cardsResp))
	require.Len(t, cardsResp.Cards, 2)
	assert.Equal(t, "&lt;b&gt;Title&lt;/b&gt;", cardsResp.Title)
	assert.Equal(t, "&lt;script&gt;alert(1)&lt;/script&gt;", cardsResp.Author)
	assert.Equal(t, "q1", cardsResp.Cards[0].Question)
	assert.Equal(t, "q2", cardsResp.Cards[1].Question)
	assert.Len(t, cardsResp.Cards[0].QuestionImages, 1)
	assert.Equal(t, uploaded.Image.ID, cardsResp.Cards[0].QuestionImages[0].ID)

	startResp := performRequest(router, http.MethodPost, "/api/v1/cards/"+created.ID, nil)
	require.Equal(t, http.StatusOK, startResp.Code)

	var session schema.StartSessionResponse
	require.NoError(t, json.Unmarshal(startResp.Body.Bytes(), &session))
	assert.Len(t, session.SessionID, 8)

	progressResp := performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)

	var progress schema.SessionProgressResponse
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 2, progress.Total)
	assert.Equal(t, 0, progress.Passed)
	assert.Equal(t, "&lt;b&gt;Title&lt;/b&gt;", progress.Title)
	assert.Equal(t, "desc", progress.Description)
	assert.Equal(t, "&lt;script&gt;alert(1)&lt;/script&gt;", progress.Author)
	require.NotNil(t, progress.Card)
	firstCard := progress.Card.Question
	secondCard := "q2"
	if firstCard == "q2" {
		secondCard = "q1"
	}

	advanceResp := performRequest(router, http.MethodPost, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusSeeOther, advanceResp.Code)
	assert.Equal(t, "/api/v1/cards/"+created.ID+"/"+session.SessionID, advanceResp.Header().Get("Location"))

	progressResp = performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 1, progress.Passed)
	require.NotNil(t, progress.Card)
	assert.Equal(t, secondCard, progress.Card.Question)

	skipResp := performRequest(router, http.MethodDelete, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusSeeOther, skipResp.Code)
	assert.Equal(t, "/api/v1/cards/"+created.ID+"/"+session.SessionID, skipResp.Header().Get("Location"))

	progressResp = performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 1, progress.Passed)
	require.NotNil(t, progress.Card)
	assert.Equal(t, secondCard, progress.Card.Question)

	advanceResp = performRequest(router, http.MethodPost, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusSeeOther, advanceResp.Code)

	progressResp = performRequest(router, http.MethodGet, "/api/v1/cards/"+created.ID+"/"+session.SessionID, nil)
	require.Equal(t, http.StatusOK, progressResp.Code)
	require.NoError(t, json.Unmarshal(progressResp.Body.Bytes(), &progress))
	assert.Equal(t, 2, progress.Passed)
	assert.Nil(t, progress.Card)
}

func TestCardsAPIErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storeObj := store.NewMemoryStore()
	require.NoError(t, storeObj.Init("admin"))

	router := gin.New()
	apiGroup := router.Group("/api/v1")
	RegisterRoutes(apiGroup, storeObj)

	resp := performRequest(router, http.MethodPost, "/api/v1/cards", mustJSON(t, schema.CreateCardSetRequest{}))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	resp = performRequest(router, http.MethodPost, "/api/v1/cards", mustJSON(t, schema.CreateCardSetRequest{
		CardSetMetadata: schema.CardSetMetadata{Title: strings.Repeat("a", 121)},
		Cards:           []schema.CardData{{Question: "q1", Answer: "a1"}},
	}))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	resp = performRequest(router, http.MethodPost, "/api/v1/cards", mustJSON(t, schema.CreateCardSetRequest{
		Cards: []schema.CardData{{Question: "q1", Answer: "a1", QuestionImages: make([]schema.CardImage, 6)}},
	}))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	resp = performMultipartRequest(router, http.MethodPost, "/api/v1/images", "file", "notes.txt", []byte("not an image"))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	resp = performRequest(router, http.MethodGet, "/api/v1/cards/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	resp = performRequest(router, http.MethodPost, "/api/v1/cards/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	resp = performRequest(router, http.MethodGet, "/api/v1/cards/missing/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)
	resp = performRequest(router, http.MethodDelete, "/api/v1/cards/missing/missing", nil)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func performRequest(router http.Handler, method string, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.5:1234"
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}

func performMultipartRequest(router http.Handler, method string, path string, fieldName string, fileName string, content []byte) *httptest.ResponseRecorder {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		panic(err)
	}
	if _, err := part.Write(content); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	req := httptest.NewRequest(method, path, body)
	req.RemoteAddr = "203.0.113.5:1234"
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}
