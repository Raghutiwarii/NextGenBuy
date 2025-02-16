package models

import (
	"ecom/backend/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	ID        uint           `json:"-" gorm:"primaryKey"`
	CreatedAt *time.Time     `json:"-"`
	UpdatedAt *time.Time     `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UUID      string         `gorm:"unique" json:"uuid,omitempty"`

	PaymentID      string `json:"-"`
	UserID             int    `json:"-"`
	MerchantID         int    `json:"-"`
	CheckoutID         int    `json:"checkout_id,omitempty"`
	NetAmountInCents   int    `json:"net_amount_in_cents,omitempty"`
	TotalAmountInCents int    `json:"total_amount_in_cents,omitempty"`
	GrossAmountInCents int    `json:"gross_amount_in_cents,omitempty"`
	PurchaseProductID  int    `json:"purchase_product_id,omitempty"`
	BankAccountID      string `json:"bank_account_id"`
	OfferID            *int   `json:"-"`

	Checkout Checkout `gorm:"foreignKey:CheckoutID" json:"checkout,omitempty"`
	Merchant Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	User     Account  `gorm:"foreignKey:UserID" json:"-"`
	Offer    *Offer   `gorm:"foreignKey:OfferID" json:"offer"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.UUID == "" {
		orderUUID, err := utils.GenerateNanoID(12, "ord_")
		if err != nil {
			utils.Error("error in creating nano id ", err)
			return err
		}

		o.UUID = fmt.Sprintf("ord_%s", orderUUID)
	}
	return nil
}

type OrderRepo struct {
	db *gorm.DB
}

// GetWithPreloads implements IOrderRepo.
func (d *OrderRepo) GetWithPreloads(where *Order) (*Order, error) {
	var (
		o = Order{}
	)

	err := d.db.
		Model(&Order{}).
		Preload(clause.Associations).
		Where(where).Last(&o).Error
	if err != nil {
		utils.Error("error in getting order ", err)
		return nil, err
	}
	return &o, nil
}

// Create implements IOrderRepo.
func (d *OrderRepo) Create(o *Order) error {
	return d.CreateWithTx(d.db, o)
}

// CreateWithTx implements IOrderRepo.
func (d *OrderRepo) CreateWithTx(tx *gorm.DB, o *Order) error {
	err := tx.Model(&Order{}).Create(o).Error
	if err != nil {
		utils.Error("error in creating order ", err)
		return err
	}
	return nil
}

// Get implements IOrderRepo.
func (d *OrderRepo) Get(where *Order) (*Order, error) {
	return d.GetWithTx(d.db, where)
}

// GetWithTx implements IOrderRepo.
func (d *OrderRepo) GetWithTx(tx *gorm.DB, where *Order) (*Order, error) {
	var (
		o = Order{}
	)

	err := tx.Model(&Order{}).
		Where(where).
		Last(&o).Error
	if err != nil {
		utils.Error("error in getting order ", err)
		return nil, err
	}
	return &o, nil
}

// GetCount implements IOrderRepo.
func (d *OrderRepo) GetCount(where *Order) (int, error) {
	var (
		count int
		err   error
	)
	err = d.db.
		Model(&Order{}).
		Where(where).Select("COUNT(id)").Scan(&count).Error
	if err != nil {
		utils.Error("error in getting count ", err)
		return count, err
	}
	return count, nil

}

// GetTotalOfferAmount implements IOrderRepo.
func (d *OrderRepo) GetTotalOfferAmount(where *Order) (*int64, error) {
	var (
		totalOfferAmountInCents *int64
	)

	err := d.db.Model(&Order{}).
		Joins("INNER JOIN d2c_checkouts ON d2c_checkouts.id = d2c_orders.checkout_id").
		Where(where).
		Select("COALESCE(SUM(d2c_checkouts.offer_amount_in_cents), 0)").
		Scan(&totalOfferAmountInCents).Error
	if err != nil {
		utils.Error("unable to total offer amount for the user ", err)
		return totalOfferAmountInCents, err
	}
	return totalOfferAmountInCents, nil
}

// Update implements IOrderRepo.
func (d *OrderRepo) Update(where *Order, c *Order) error {
	return d.UpdateWithTx(d.db, where, c)
}

// UpdateWithTx implements IOrderRepo.
func (d *OrderRepo) UpdateWithTx(tx *gorm.DB, where *Order, c *Order) error {
	err := tx.Model(&Order{}).Where(where).Updates(&c).Error
	if err != nil {
		utils.Error("error in updating checkout ", err)
		return err
	}

	return nil
}

func (d *OrderRepo) GetOrderWithIds(ids []int) (*[]string, error) {
	var (
		o = []string{}
	)

	err := d.db.Model(&Order{}).
		Where(`id IN ?`, ids).Select("uuid").
		Find(&o).Error
	if err != nil {
		utils.Error("error in getting order ", err)
		return nil, err
	}
	return &o, nil
}
