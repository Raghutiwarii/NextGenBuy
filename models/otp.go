package models

import (
	"ecom/backend/utils"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OTPRepo struct {
	db *gorm.DB
}

type OTP struct {
	gorm.Model
	AccountUUID string    `gorm:"index"`
	Code        string    `gorm:"type:varchar(6);not null"`
	ExpiresAt   time.Time `gorm:"not null"`
}

func (otp *OTPRepo) Create(op *OTP) error {
	return otp.CreateWithTx(otp.db, op)
}

func (otp *OTPRepo) CreateWithTx(tx *gorm.DB, op *OTP) error {
	err := tx.Clauses(clause.OnConflict{UpdateAll: true}).
		Model(&OTP{}).Create(&op).Error
	if err != nil {
		utils.Error("error in creating otp ", err)
		return err
	}
	return nil
}

func (otp *OTPRepo) Get(where *OTP) (*OTP, error) {
	return otp.GetWithTx(otp.db, where)
}

func (otp *OTPRepo) GetWithTx(tx *gorm.DB, where *OTP) (*OTP, error) {
	var (
		op = OTP{}
	)
	err := tx.Model(&OTP{}).
		Where(where).
		Scopes(OmitIDToDeletedAtFields).
		Last(&op).Error
	if err != nil {
		utils.Error("unable to query otp ", err)
		return nil, err
	}
	return &op, nil
}

func (otp *OTPRepo) Update(where *OTP, op *OTP) error {
	return otp.UpdateWithTx(otp.db, where, op)
}

func (otp *OTPRepo) UpdateWithTx(tx *gorm.DB, where *OTP, op *OTP) error {
	err := tx.
		Model(&OTP{}).Where(where).Updates(&op).Error
	if err != nil {
		utils.Error("unable to update otp ", err)
		return err
	}
	return nil
}
