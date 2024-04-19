package model

type RequestShortURL struct {
	URL string `json:"url,omitempty"`
}

type ResponseShortURL struct {
	Result string `json:"result,omitempty"`
}
