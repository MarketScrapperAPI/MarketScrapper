package models

type Message struct {
	Item   Item   `json:"item"`
	Market Market `json:"market"`
}
