package schema

type CardData struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
	Remarks  string `json:"remarks"`
}

type CardSetMetadata struct {
	Title       string `json:"title" binding:"max=120"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

type CreateCardSetRequest struct {
	CardSetMetadata
	Cards []CardData `json:"cards" binding:"required"`
}

type Card struct {
	ID string `json:"id"`
	CardData
}

type CardSetResponse struct {
	ID string `json:"id"`
	CardSetMetadata
	Cards []Card `json:"cards"`
}

type CreateCardSetResponse struct {
	ID string `json:"id"`
}

type StartSessionResponse struct {
	SessionID string `json:"session_id"`
}

type SessionProgressResponse struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	CardSetMetadata
	Card *Card `json:"card"`
}
