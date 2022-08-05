package models

type Item struct {
	Name             string   `json:"name"`
	Brand            string   `json:"brand"`
	Package          string   `json:"package"`
	PricePerItem     float32  `json:"price_per_item"`
	PricePerQuantity *float32 `json:"price_per_quantity,omitempty"`
	QuantityUnit     *string  `json:"quantity_unit,omitempty"`
	Url              string   `json:"url"`
	ImageUrl         string   `json:"image_url"`
}

type Market struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type Message struct {
	Item   Item   `json:"item"`
	Market Market `json:"market"`
}
