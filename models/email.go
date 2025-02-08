package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	logger "ecom/backend/utils"
)

type Email struct {
	ID        uint           `json:"-" gorm:"primaryKey"`
	CreatedAt *time.Time     `json:"-"`
	UpdatedAt *time.Time     `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Email      string `json:"email" gorm:"unique;not null;"`
	IsVerified *bool  `json:"-" gorm:"default:false"`
	Domain     string `json:"-" gorm:"not null"`

	Accounts []*Account `gorm:"many2many:account_emails" json:"-"`
}

func (e *Email) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"email"}),
	})
	return nil
}

type emailRepo struct {
	db *gorm.DB
}

// GetLastActiveEmailOfAccount implements IEmailRepo.
func (repo *emailRepo) GetLastEmailOfAccount(accountID uint) (*Email, error) {
	var (
		email = Email{}
	)
	err := repo.db.Model(&Email{}).
		Joins("JOIN account_emails on account_emails.email_id =emails.id ").
		Joins("JOIN accounts on accounts.id = account_emails.account_id").
		Where("accounts.id = ?", accountID).Last(&email).Error
	if err != nil {
		logger.Error("error in getting email ", err)
		return nil, err
	}
	return &email, nil
}

// GetAll implements IEmailRepo.
func (repo *emailRepo) GetAll(where *Email) (*[]Email, error) {
	var (
		results = &[]Email{}
	)
	err := repo.db.Model(&Email{}).Preload("Accounts").Where(where).Find(results).Error
	if err != nil {
		logger.Error("unable to create email record: ", err)
		return nil, err
	}
	return results, nil
}

// GetWithAccount implements IAccountEmailRepo.
func (a *emailRepo) GetWithAccount(where *Email) (*Email, error) {
	var (
		accountEmail = Email{}
	)
	err := a.db.
		Model(&Email{}).Preload("Accounts").
		Where(where).
		Last(&accountEmail).Error
	if err != nil {
		logger.Error("error in getting account email ", err)
		return nil, err
	}
	return &accountEmail, nil
}

// Create implements IAccountEmailRepo.
func (a *emailRepo) Create(i *Email) error {
	return a.CreateWithTx(a.db, i)
}

// CreateWithTx implements IAccountEmailRepo.
func (a *emailRepo) CreateWithTx(tx *gorm.DB, i *Email) error {
	err := tx.Model(&Email{}).Create(i).Error
	if err != nil {
		logger.Error("unable to create account email association ", err)
		return err
	}
	return nil
}

// Delete implements IAccountEmailRepo.
func (a *emailRepo) Delete(where *Email) error {
	return a.DeleteWithTx(a.db, where)
}

// DeleteWithTx implements IAccountEmailRepo.
func (a *emailRepo) DeleteWithTx(tx *gorm.DB, where *Email) error {
	err := tx.
		Model(&Email{}).Where(where).
		Delete(&Email{}).Error
	if err != nil {
		logger.Error("error in deleting account email ", err)
		return err
	}
	return nil
}

// Get implements IAccountEmailRepo.
func (a *emailRepo) Get(where *Email) (*Email, error) {
	return a.GetWithTx(a.db, where)
}

// GetWithTx implements IAccountEmailRepo.
func (a *emailRepo) GetWithTx(tx *gorm.DB, where *Email) (*Email, error) {
	var (
		accountEmail = Email{}
	)

	err := tx.
		Model(&Email{}).
		Scopes(OmitIDToDeletedAtFields).
		Where(where).Last(&accountEmail).Error
	if err != nil {
		logger.Error("error in getting account email ", err)
		return nil, err

	}
	return &accountEmail, nil
}

// Update implements IAccountEmailRepo.
func (a *emailRepo) Update(where *Email, i *Email) error {
	return a.UpdateWithTx(a.db, where, i)
}

// UpdateWithTx implements IAccountEmailRepo.
func (a *emailRepo) UpdateWithTx(tx *gorm.DB, where *Email, i *Email) error {
	err := tx.Model(&Email{}).
		Clauses(clause.Returning{
			Columns: []clause.Column{
				{
					Name: "id",
				},
			},
		}).Where(where).Updates(i).Error
	if err != nil {
		logger.Error("error in updating account email ", err)
		return err
	}
	return nil
}

func (a *emailRepo) UpdateEmailToVerified(tx *gorm.DB, accountID uint, email *Email) error {
	err := tx.Model(&Email{}).
		Clauses(clause.Returning{
			Columns: []clause.Column{
				{
					Name: "id",
				},
			},
		}).Where(&Email{
		Email: email.Email,
	}).Updates(email).Error
	if err != nil {
		logger.Error("error in  updating account email ", err)
		return err
	}
	return nil
}

// CreateAccountEmail implements IUser.
func (ar *emailRepo) CreateAccountEmail(e *Email) (*Email, error) {
	return ar.CreateAccountEmailWithTx(ar.db, e)
}

// CreateAccountEmailWithTx implements IUser.
func (ar *emailRepo) CreateAccountEmailWithTx(tx *gorm.DB, e *Email) (*Email, error) {

	err := tx.Clauses(clause.OnConflict{DoNothing: true}).
		Clauses(clause.Returning{Columns: []clause.Column{
			{
				Name: "id",
			},
		}}).Model(&Email{}).Create(e).Error
	if err != nil {
		logger.Error("unable to create email record: ", err)
		return nil, err
	}
	return e, nil
}
