// пакет model служит для представления используемых моделей приложения
package model

// RequestShortURL запрос с ссылкой для сокращения
type RequestShortURL struct {
	URL string `json:"url,omitempty"`
}

// ResponseShortURL запрос получения оригинальной ссылки
type ResponseShortURL struct {
	Result string `json:"result,omitempty"`
}

// StorageJSON структура для хранения в файле
type StorageJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
	IsDeleted   bool   `json:"is_deleted"`
}

// StorageJSONWithUserID структура для хранения в файле с добавлением функицональности разделения пользователей
type StorageJSONWithUserID struct {
	StorageJSON
	UserID string `json:"user_id"`
}

// DeleteURLMessage запрос на удаление сокращенной ссылки для конкретного пользователя
type DeleteURLMessage struct {
	UserID   string `json:"user_id"`
	ShortURL string `json:"short_url"`
}

// BatchRequest запрос для массового сокращения ссылок
type BatchRequest []BatchRequestElement

// BatchRequestElement отдельный элемент BatchRequest
type BatchRequestElement struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"-"`
}

// BatchResponse ответ на массовый запрос сокращения ссылок
type BatchResponse []BatchResponseElement

// BatchResponseElement отдельный элемент BatchResponse
type BatchResponseElement struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	OriginalURL   string `json:"-"`
	Collision     bool   `json:"-"`
	Error         error  `json:"-"`
}
