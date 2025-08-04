package models

type Email struct {
	ToEmail string `json:"to_email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
