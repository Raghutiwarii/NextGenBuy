package models

import (
	"ecom/backend/utils"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CheckoutStatus string

const (
	CheckoutStatusPending   CheckoutStatus = "PENDING"
	CheckoutStatusCompleted CheckoutStatus = "COMPLETED"
	CheckoutStatusFailed    CheckoutStatus = "FAILED"
)

type Checkout struct {
	gorm.Model
	CheckoutID    string         `json:"checkout_id" gorm:"index"`
	UserID        string         `json:"user_id" gorm:"index"`
	TotalAmount   int            `json:"total_amount"`
	Status        CheckoutStatus `json:"status" gorm:"AUDITABLE"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CheckoutItems []CheckoutItem `json:"checkout_items" foreignKey:"CheckoutID",references:"CheckoutID"`
}

type checkoutRepo struct {
	db *gorm.DB
}

func (chk *Checkout) BeforeCreate(tx *gorm.DB) error {
	// Generate a unique checkout UUID
	checkoutId, err := utils.GenerateNanoID(12, "chk_")
	if err != nil {
		utils.Error("unable to generate nano id ", err)
		return err
	}
	chk.CheckoutID = checkoutId

	tx.Statement.AddClause(clause.Returning{Columns: []clause.Column{
		{Name: "checkout_id"},
		{Name: "id"},
		{Name: "user_id"},
	}})

	return nil
}

// Create a new checkout
func (chk *checkoutRepo) Create(c *Checkout) error {
	return chk.CreateWithTx(chk.db, c)
}

// Create a checkout within a transaction
func (chk *checkoutRepo) CreateWithTx(tx *gorm.DB, c *Checkout) error {
	err := tx.Model(&Product{}).
		Clauses(clause.Returning{Columns: []clause.Column{
			{Name: "checkout_id"},
			{Name: "id"},
		}}).Create(c).Error
	if err != nil {
		utils.Error("unable to create checkout ", err)
		return err
	}
	return nil
}

// Get checkout by ID
func (chk *checkoutRepo) GetByID(checkoutId string) (*Checkout, error) {
	var checkout Checkout
	err := chk.db.Model(&Product{}).Where("checkout_id = ?", checkoutId).Last(&checkout).Error
	if err != nil {
		utils.Error("unable to get checkout ", err)
		return nil, err
	}
	return &checkout, nil
}

func (chk *checkoutRepo) Get(where *Checkout) (*Checkout, error) {
	return chk.GetWithTx(chk.db, where)
}

func (chk *checkoutRepo) GetWithTx(tx *gorm.DB, where *Checkout) (*Checkout, error) {
	var (
		c = Checkout{}
	)
	err := tx.Model(&Checkout{}).
		Where(where).
		Scopes(OmitIDToDeletedAtFields).
		Last(&c).Error
	if err != nil {
		utils.Error("unable to query checkout ", err)
		return nil, err
	}
	return &c, nil
}
