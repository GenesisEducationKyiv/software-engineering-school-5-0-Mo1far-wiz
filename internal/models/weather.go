package models

type Weather struct {
	Temperature int    `json:"temperature"`
	Humidity    int    `json:"humidity"`
	Description string `json:"description"`
}
