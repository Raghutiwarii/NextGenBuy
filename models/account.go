package models

import (
	"errors"
	"time"

	utils "ecom/backend/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	ID        uint           `json:"id,omitempty" gorm:"primaryKey"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	UUID                string  `gorm:"unique" json:"uuid,omitempty"`
	FirstName           string  `json:"first_name"`
	LastName            string  `json:"last_name"`
	CountryCode         string  `gorm:"not null;default:+1" json:"country_code,omitempty"`
	PhoneNumber         *string `gorm:"uniqueIndex:idx_unique_phone_email" json:"phone_number,omitempty"`
	PhoneNumberVerified *bool   `json:"phone_number_verified" gorm:"default:false"`

	PrimaryEmailID *uint  `gorm:"uniqueIndex:idx_unique_phone_email" json:"-"`
	PrimaryEmail   *Email `json:"primary_email,omitempty"`

	Emails      []*Email     `gorm:"many2many:account_emails" json:"emails,omitempty"`
	Credentials []Credential `json:"-"`
	Role        uint         `json:"role"`
}

func GetAccountByPhoneNumber(db *gorm.DB, phoneNumber string) (*Account, error) {
	var account Account
	err := db.Where("phone_number = ?", phoneNumber).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

type accountRepo struct {
	db *gorm.DB
}

func (a *Account) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(clause.OnConflict{
		Columns:   []clause.Column{{Name: "phone_number"}, {Name: "primary_email_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"phone_number", "primary_email_id"}),
	})
	tx.Statement.AddClause(clause.Returning{Columns: []clause.Column{
		{
			Name: "uuid",
		},
		{
			Name: "id",
		},
	}})

	userUUID, err := utils.GenerateNanoID(12, "acc_")
	if err != nil {
		utils.Error("unable to generate nano id ", err)
		return err
	}
	a.UUID = userUUID

	totpAccountName := userUUID
	if a.PhoneNumber != nil && *a.PhoneNumber != "" {
		totpAccountName = *a.PhoneNumber
	}

	ac, err := CreateOTPSecret(totpAccountName)
	if err != nil {
		return err
	}
	a.Credentials = append(a.Credentials, *ac)

	if len(a.Emails) > 0 && a.PrimaryEmailID == nil {
		a.PrimaryEmail = a.Emails[0]
	}

	return nil
}

func (a *Account) AfterSave(tx *gorm.DB) error {
	if len(a.Emails) > 0 && a.PrimaryEmailID == nil {
		return tx.Model(&Account{ID: a.ID}).Updates(&Account{PrimaryEmailID: &a.Emails[0].ID}).Error
	}
	return nil
}

// Create implements IUser.
func (ar *accountRepo) Create(u *Account) error {
	return ar.CreateWithTx(ar.db, u)
}

// CreateWithTx implements IUser.
func (ar *accountRepo) CreateWithTx(tx *gorm.DB, u *Account) error {
	err := tx.Model(&Account{}).Create(&u).Error
	if err != nil {
		utils.Error("unable to create user ", err)
		return err
	}
	return nil
}

func (repo *accountRepo) AddEmails(a *Account, emails []Email) error {
	return repo.db.Model(&a).Association("Emails").Append(emails)
}

// Delete implements IUser.
func (ar *accountRepo) Delete(userID uint) error {
	return ar.DeleteWithTx(ar.db, &Account{ID: userID})
}

// GetByID implements IUser.
func (ar *accountRepo) GetByID(userID uint) (*Account, error) {
	return ar.GetWithTx(ar.db, &Account{ID: userID})
}

// GetWithTx implements IUser.
func (ar *accountRepo) GetWithTx(tx *gorm.DB, where *Account) (*Account, error) {
	var (
		u = Account{}
	)
	err := tx.Model(&Account{}).
		Preload("Emails").
		Where(where).
		Scopes(OmitIDToDeletedAtFields).
		Last(&u).Error
	if err != nil {
		utils.Error("unable to query user ", err)
		return nil, err
	}
	return &u, nil
}

// Update implements IUser.
func (ar *accountRepo) Update(where *Account, a *Account) error {
	return ar.UpdateWithTx(ar.db, where, a)
}

// UpdateWithTx implements IUser.
func (ar *accountRepo) UpdateWithTx(tx *gorm.DB, where *Account, a *Account) error {
	err := tx.Model(&Account{}).
		Where(where).Updates(&a).Error
	if err != nil {
		utils.Error("unable to update user ", err)
		return err
	}
	return nil
}

// DeleteWithTx implements IUser.
func (ar *accountRepo) DeleteWithTx(tx *gorm.DB, u *Account) error {
	err := tx.Model(&Account{}).
		Where(u).
		Delete(&Account{}).
		Error
	if err != nil {
		utils.Error("unable to delete user ", err)
		return err
	}
	return nil
}

// GetByEmail implements IUser.
func (ar *accountRepo) GetByEmail(email string) ([]*Account, error) {
	var (
		accountEmail *Email
	)

	err := ar.db.Preload("Account").
		Model(&Email{}).Where(Email{
		Email: email,
	}).Last(&accountEmail).Error
	if err != nil {
		utils.Error("unable to fetch account email ", err)
		return nil, err
	}
	if accountEmail != nil {
		return accountEmail.Accounts, nil
	}
	return nil, errors.New("error in getting account associated with email")
}

func (ar *accountRepo) Get(where *Account) (*Account, error) {
	return ar.GetWithTx(ar.db, where)
}

// FindOne implements IAccount.
func (ar *accountRepo) FindOne(email, phoneNumber, accountUUID string) (*Account, error) {
	var (
		account = Account{}
	)

	builder := ar.db.Model(&Account{}).Preload("Emails", func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	}).Preload("PrimaryEmail")

	if phoneNumber != "" {
		builder.Or(&Account{
			PhoneNumber: &phoneNumber,
		})
	}

	if accountUUID != "" {
		builder.Or(&Account{
			UUID: accountUUID,
		})
	}

	if email != "" {
		builder.
			Joins("LEFT JOIN account_emails ON account_emails.account_id = accounts.id").
			Joins("LEFT JOIN emails on emails.id = account_emails.email_id").
			Or("emails.email = ?", email)
	}

	err := builder.Find(&account).Error
	if err != nil {
		utils.Error("unable to find account ", err)
		return nil, err
	}

	return &account, nil
}

func (ar *accountRepo) GetAllAccountsWithSameMailDomain(tx *gorm.DB, emailDomain string) (*[]Account, error) {

	var (
		accounts = []Account{}
	)
	err := tx.Model([]Account{}).Preload("Emails").
		Joins("account_emails on  account_emails.account_id = accounts.id").
		Joins("emails on emails.id = account_emails.email_id").
		Where("emails.domain = ? AND emails.is_verified = ?", emailDomain, true).
		Order("id").Find(&accounts).Error

	if err != nil {
		utils.Error("unable to fetch email id ", err)
		return nil, err
	}
	return &accounts, nil
}

func (ar *accountRepo) GetWithCredentials(where *Account, credcredentialType CredentialsTypeSlug) (*Account, error) {
	var (
		u = Account{}
	)
	err := ar.db.Model(&Account{}).
		Preload("Emails").
		Preload("Credentials", "Type = ?", credcredentialType).
		Where(where).
		Scopes(OmitIDToDeletedAtFields).
		Last(&u).Error
	if err != nil {
		utils.Error("unable to query user ", err)
		return nil, err
	}
	return &u, nil
}

func (ar *accountRepo) CheckAccountAssociatedForTheEmail(accountUUID string, email string) (*Email, error) {
	var accountEmail Email
	if err := ar.db.Model(&Email{}).
		Joins("JOIN account_emails on account_emails.email_id = emails.id").
		Joins("JOIN accounts ON account_emails.account_id = accounts.id").
		Where("accounts.uuid = ? AND emails.email = ?", accountUUID, email).
		First(&accountEmail).Error; err != nil {
		utils.Error("failed to get account associations: ", err)
		return nil, err
	}
	return &accountEmail, nil
}

func (ar *accountRepo) MarkAccountAsNotVerified(where *Account) error {
	err := ar.db.Model(&Account{}).Scopes(OmitIDToDeletedAtFields).Where(&Account{ID: where.ID}).
		Save(where).Error
	if err != nil {
		utils.Error("unable to update user email as not verified: ", err)
		return err
	}
	return nil
}

func (repo *accountRepo) BulkInsert(records *[]Account, conds ...clause.Expression) error {
	err := repo.db.Model(&Account{}).Clauses(conds...).Create(records).Error
	if err != nil {
		utils.Error("error in bulk creating", err)
		return err
	}
	return nil
}
