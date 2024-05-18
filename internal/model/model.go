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
}
type StorageJSONWithUserID struct {
	StorageJSON
	UserID string `json:"user_id"`
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
