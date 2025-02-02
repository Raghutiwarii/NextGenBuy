package models

import (
	utils "ecom/backend/utils"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type customerRepo struct {
	db *gorm.DB
}

type Customer struct {
	gorm.Model
	AccountUUID           string    `json:"account_uuid,omitempty" gorm:"uniqueIndex:unique_user"`
	CustomerID            string    `json:"customer_id" gorm:"unique"`
	ReferralCode          string    `json:"referral_code,omitempty"`
	ProfileImageURL       string    `json:"profile_image_url"`
	ReferredByID          *uint64   `json:"referred_by_id,omitempty"`
	ReferredBy            *Customer `json:"referred_by,omitempty" gorm:"foreignKey:ReferredByID;references:ID"`
	WalletID              string    `json:"wallet_id,omitempty"`
	BirthMonth            uint      `json:"birth_month,omitempty"`
	BirthDay              uint      `json:"birth_day,omitempty"`
	BirthYear             uint      `json:"birth_year,omitempty"`
	SSN                   string    `json:"ssn"`
	HasWalletTransactions *bool     `json:"has_wallet_transactions,omitempty"`

	Addresses []Address `json:"-" gorm:"polymorphic:Owner;"`

	// common associations
	RoleID RoleID `json:"role_id" gorm:"not null"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	nanoID, err := utils.GenerateNanoID(15, "C_")
	if err != nil {
		utils.Error("error in creating nano id ", err)
		tx.Rollback()
		return err
	}

	CustomerUUID := fmt.Sprintf("m_%s", nanoID)
	c.CustomerID = CustomerUUID
	return nil
}

func (c *customerRepo) GetAccountUUIDsByWalletUUIDs(uuids []string) ([]string, error) {
	var (
		accountUUIDs = []string{}
	)
	err := c.db.Model(&Customer{}).
		Where("wallet_id IN (?)", uuids).
		Select("account_uuid").Scan(&accountUUIDs).Error
	if err != nil {
		utils.Error("error in getting wallet uuids ", err)
		return accountUUIDs, nil
	}
	return accountUUIDs, err
}

func (cus *customerRepo) Create(c *Customer) error {
	return cus.CreateWithTx(cus.db, c)
}

func (cus *customerRepo) CreateWithTx(tx *gorm.DB, c *Customer) error {
	err := tx.Clauses(clause.OnConflict{UpdateAll: true}).
		Model(&Customer{}).Create(&c).Error
	if err != nil {
		utils.Error("error in creating user ", err)
		return err
	}
	return nil
}

func (cus *customerRepo) Get(where *Customer) (*Customer, error) {
	return cus.GetWithTx(cus.db, where)
}

func (cus *customerRepo) GetWithTx(tx *gorm.DB, where *Customer) (*Customer, error) {
	var (
		c = Customer{}
	)
	err := tx.Model(&Customer{}).
		Where(where).
		Scopes(OmitIDToDeletedAtFields).
		Last(&c).Error
	if err != nil {
		utils.Error("unable to query customer ", err)
		return nil, err
	}
	return &c, nil
}
