package service

type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ServedBy string `json:"servedBy"`
}

type AccountData struct {
	ID   string `json:""`
	Name string `json:"name"`
}

type ReportNotification struct {
	AccountID string `json:"accountId"`
	ReadAt    string `json:"readAt"`
}
