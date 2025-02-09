package models

type CheckoutItem struct {
	ID         uint `gorm:"primaryKey"`
	ProductID  uint `json:"product_id" gorm:"index"`
	Quantity   uint `json:"quantity"`
	Price      int  `json:"price"`
	TotalPrice int  `json:"total_price"`
	CheckoutID uint `json:"checkout_id"`
}
