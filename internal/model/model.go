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
