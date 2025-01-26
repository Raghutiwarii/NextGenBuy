package models

import (
	utils "ecom/backend/utils"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CredentialsTypeSlug string

const (
	CredentialsTypeOTPSecret CredentialsTypeSlug = "otp_secret"
	CredentialsTypePassword  CredentialsTypeSlug = "password"
)

type Credential struct {
	ID        uint64         `json:"id" gorm:"primaryKey"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Password  string              `gorm:"not null"`
	Type      CredentialsTypeSlug `gorm:"not null,default:password"`
	AccountID uint

	Account Account
}

type credentialRepo struct {
	DB *gorm.DB
}

func (c *Credential) BeforeCreate(tx *gorm.DB) (err error) {
	// check if the password is alerady hashed, if not then bcrypt it.
	if c.Type != CredentialsTypeOTPSecret && !strings.HasPrefix(c.Password, fmt.Sprintf("$2a$%02d$", bcrypt.DefaultCost)) {
		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		c.Password = string(hashedPasswordBytes)
	}

	return
}

func CreateOTPSecret(accountName string) (ac *Credential, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      utils.JWTIssuer,
		AccountName: accountName,
	})

	if err != nil {
		return nil, err
	}

	return &Credential{
		Password: key.Secret(),
		Type:     CredentialsTypeOTPSecret,
	}, nil
}

// Create implements IAuthCredential.
func (cr *credentialRepo) Create(c *Credential) error {
	return cr.CreateWithTx(cr.DB, c)
}

// CreateWithTx implements IAuthCredential.
func (cr *credentialRepo) CreateWithTx(tx *gorm.DB, c *Credential) error {
	err := tx.
		Model(&Credential{}).
		Create(c).Error
	if err != nil {
		utils.Error("unable to create auth credential ", err)
		return err
	}
	return nil
}

// Delete implements IAuthCredential.
func (cr *credentialRepo) Delete(where *Credential) error {
	return cr.DeleteWithTx(cr.DB, where)
}

// DeleteWithTx implements IAuthCredential.
func (cr *credentialRepo) DeleteWithTx(tx *gorm.DB, where *Credential) error {
	err := tx.
		Model(&Credential{}).
		Delete(where).Error
	if err != nil {
		utils.Error("unable to delete auth credential record ", err)
		return err
	}
	return nil
}

// Get implements IAuthCredential.
func (cr *credentialRepo) Get(where *Credential) (*Credential, error) {
	return cr.GetWithTx(cr.DB, where)
}

// GetWithTx implements IAuthCredential.
func (cr *credentialRepo) GetWithTx(tx *gorm.DB, where *Credential) (*Credential, error) {
	var (
		result = Credential{}
	)
	err := tx.Scopes(OmitIDToDeletedAtFields).
		Model(&Credential{}).
		Where(where).Last(&result).Error
	if err != nil {
		utils.Error("unable to get the last auth credential ", err)
		return nil, err
	}
	return &result, nil
}

// Update implements IAuthCredential.
func (cr *credentialRepo) Update(where *Credential, u *Credential) error {
	return cr.UpdateWithTx(cr.DB, where, u)
}

// UpdateWithTx implements IAuthCredential.
func (cr *credentialRepo) UpdateWithTx(tx *gorm.DB, where *Credential, u *Credential) error {
	err := tx.Model(&Credential{}).
		Where(where).
		Updates(&u).Error
	if err != nil {
		utils.Error("unable to update auth credential record ", err)
		return err
	}
	return nil
}

// CheckIfPasswordIsValid implements IAuthCredential.
func (cr *credentialRepo) CheckIfPasswordIsValid(userID uint, password string) (bool, error) {
	// Fetch password credential for the user
	var (
		hashedPassword string
	)

	err := cr.DB.Model(&Credential{}).Where(&Credential{
		AccountID: userID,
		Type:      CredentialsTypePassword,
	}).
		Select("password").
		Scan(&hashedPassword).Error
	if err != nil {
		utils.Error("unable  to scan password ", err)
		return false, nil
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, nil
}
