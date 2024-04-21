package request

type Channel struct {
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	IsNews      bool   `json:"is_news"`
	RecipientID string `json:"recipient_id"`
}

type EditChannel struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}
