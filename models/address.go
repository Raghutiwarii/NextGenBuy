package models

import (
	"ecom/backend/utils"
	"time"

	"gorm.io/gorm"
)

type Address struct {
	ID        uint64          `json:"-" gorm:"primaryKey"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `json:"-" gorm:"index"`

	Line1         string `json:"line_1" gorm:"AUDITABLE"`
	Line2         string `json:"line_2" gorm:"AUDITABLE"`
	City          string `json:"city" gorm:"AUDITABLE"`
	ZipCode       string `json:"zipcode" gorm:"AUDITABLE"`
	Country       string `json:"country" gorm:"AUDITABLE;default:US"`
	State         string `json:"state" gorm:"AUDITABLE"`
	IsResidential bool   `json:"is_residential" gorm:"AUDITABLE"`
	IsPrimary     *bool  `json:"is_primary,omitempty" gorm:"default:false"`

	OwnerID   uint   `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
}

type addressRepo struct {
	db *gorm.DB
}

// GetAll implements IAddress.
func (ar *addressRepo) GetAll(where *Address) ([]Address, error) {
	var (
		a []Address
	)

	err := ar.db.Model(&Address{}).Where(where).Find(&a).Error
	if err != nil {
		utils.Error("error in getting address ", err)
		return nil, err
	}
	return a, nil
}

// Create implements IAddress.
func (ar *addressRepo) Create(a *Address) error {
	return ar.CreateWithTX(ar.db, a)
}

// CreateWithTX implements IAddress.
func (ar *addressRepo) CreateWithTX(tx *gorm.DB, a *Address) error {
	err := tx.Model(&Address{}).Create(a).Error
	if err != nil {
		utils.Error("unable to create address ", err)
		return err
	}
	return nil
}

// Get implements IAddress.
func (ar *addressRepo) Get(where *Address) (*Address, error) {
	return ar.GetWithTx(ar.db, where)
}

// GetWithTx implements IAddress.
func (ar *addressRepo) GetWithTx(tx *gorm.DB, where *Address) (*Address, error) {
	var (
		address = Address{}
	)
	err := tx.Model(&Address{}).Where(where).Last(&address).Error
	if err != nil {
		utils.Error("unable to get address ", err)
		return nil, err
	}
	return &address, nil
}

// Update implements IAddress.
func (ar *addressRepo) Update(where *Address, a *Address) error {
	return ar.UpdateWithTx(ar.db, where, a)
}

// UpdateWithTx implements IAddress.
func (ar *addressRepo) UpdateWithTx(tx *gorm.DB, where *Address, a *Address) error {
	err := tx.Model(&Address{}).Where(where).Updates(a).Error
	if err != nil {
		utils.Error("unable to update address ", err)
		return err
	}
	return nil

}

func (old *Address) CheckSameAddress(new Address) bool {
	return old.Line1 == new.Line1 &&
		old.City == new.City &&
		old.State == new.State &&
		old.ZipCode == new.ZipCode
}
