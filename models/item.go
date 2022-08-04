package models

type Item struct {
	Name             string   `json:"name"`
	Brand            string   `json:"brand"`
	Package          string   `json:"package"`
	PricePerItem     float32  `json:"price_per_item"`
	PricePerQuantity *float32 `json:"price_per_quantity,omitempty"`
	QuantityUnit     *string  `json:"quantity_unit,omitempty"`
}
