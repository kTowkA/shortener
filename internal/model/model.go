package model

type RequestShortURL struct {
	URL string `json:"url,omitempty"`
}

type ResponseShortURL struct {
	Result string `json:"result,omitempty"`
}

type StorageJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	IsDeleted   bool   `json:"is_deleted"`
}
type StorageJSONWithUserID struct {
	StorageJSON
	UserID string `json:"user_id"`
}
type DeleteURLMessage struct {
	UserID   string `json:"user_id"`
	ShortURL string `json:"short_url"`
}
type BatchRequest []BatchRequestElement
type BatchRequestElement struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"-"`
}

type BatchResponse []BatchResponseElement
type BatchResponseElement struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	OriginalURL   string `json:"-"`
	Collision     bool   `json:"-"`
	Error         error  `json:"-"`
}
