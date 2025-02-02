package models

import (
	"ecom/backend/utils"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MerchantOnboardingState string

const (
	MerchantOnboardingStateUpdateMerchantUserDetails MerchantOnboardingState = "UPDATE_MERCHANT_USER_DETAILS"
	MerchantOnboardingStateVerifyAccount             MerchantOnboardingState = "VERIFY_ACCOUNT"
	MerchantOnboardingStateUpdateMerchantDetails     MerchantOnboardingState = "UPDATE_MERCHANT_DETAILS"
	MerchantOnboardingStateConnectBankAccount        MerchantOnboardingState = "CONNECT_BANK_ACCOUNT"
	MerchantOnboardingStateConfirmDetails            MerchantOnboardingState = "CONFIRM_DETAILS"
	MerchantOnboardingStateSelectSubscriptionPlan    MerchantOnboardingState = "SELECT_SUBSCRIPTION_PLAN"
	MerchantOnboardingStateAgreeToMerchantAgreement  MerchantOnboardingState = "AGREE_TO_MERCHANT_AGREEMENT"
	MerchantOnboardingStateApprovalPending           MerchantOnboardingState = "APPROVAL_PENDING"
	MerchantOnboardingStateApproved                  MerchantOnboardingState = "APPROVED"

	MerchantOnboardingStateShowOnboardingComplete MerchantOnboardingState = "ONBOARDING_COMPLETE"
	MerchantOnboardingStateShowConfirmScreen      MerchantOnboardingState = "SHOW_CONFIRM_SCREEN"
)

type Merchant struct {
	ID                           uint           `json:"-" gorm:"primaryKey"`
	CreatedAt                    *time.Time     `json:"-"`
	UpdatedAt                    *time.Time     `json:"-"`
	DeletedAt                    gorm.DeletedAt `json:"-" gorm:"index"`
	UUID                         string         `gorm:"unique" json:"uuid,omitempty"`
	AccountUUID                  string         `json:"account_uuid"`
	CorporateName                string         `json:"corporate_name,omitempty" gorm:"AUDITABLE"`
	DetailsConfirmedAt           *time.Time     `json:"details_confirmed_at,omitempty"`
	LogoURL                      string         `json:"logo_url,omitempty" gorm:"AUDITABLE"`
	DoingBusinessAs              string         `json:"doing_business_as,omitempty" gorm:"AUDITABLE"`
	Website                      string         `json:"website,omitempty" gorm:"AUDITABLE"`
	EmployerIdentificationNumber string         `json:"employer_identification_number" gorm:"AUDITABLE"`
	ApprovedAt                   *time.Time     `json:"approved_at,omitempty"`
	SupportEmail                 string         `json:"support_email,omitempty" gorm:"AUDITABLE"`
	CorporatePhone               string         `json:"corporate_phone,omitempty" gorm:"AUDITABLE"`
	SupportPhoneNumber           string         `json:"support_phone_number,omitempty" gorm:"AUDITABLE"`
	WalletID                     string         `json:"wallet_id,omitempty"`
	Address                      Address        `gorm:"embedded"`

	ApplicationCurrentStatus  MerchantOnboardingState `json:"application_current_status" gorm:"AUDITABLE"`
	IsBlocked                 *bool                   `json:"is_blocked,omitempty" gorm:"default:false"`
	BankProviderReferenceUUID string                  `json:"bank_provider_reference_uuid,omitempty"`
	ReturnAndRefundPolicyLink string                  `json:"return_and_refund_policy_link,omitempty"`
	NextBillingDate           *time.Time              `json:"next_billing_date"`

	Logo string `json:"logo" gorm:"polymorphic:Owner"`
}

type merchantRepo struct {
	db *gorm.DB
}

// Create implements IMerchant.
func (mr *merchantRepo) Create(m *Merchant) error {
	return mr.CreateWithTx(mr.db, m)
}

func (m *Merchant) BeforeCreate(tx *gorm.DB) error {
	merchantUUID, err := utils.GenerateNanoID(15, "m_")
	if err != nil {
		utils.Error("error in creating nano id ", err)
		tx.Rollback()
		return err
	}

	m.UUID = merchantUUID
	return nil
}

// GetByUUID implements IMerchant.
func (mr *merchantRepo) GetByUUID(merchantUUID string) (*Merchant, error) {
	return mr.GetWithTX(mr.db, &Merchant{
		UUID: merchantUUID,
	})
}

func (mr *merchantRepo) Get(where *Merchant) (*Merchant, error) {
	return mr.GetWithTX(mr.db, where)
}

// GetWithTX implements IMerchant.
func (mr *merchantRepo) GetWithTX(tx *gorm.DB, where *Merchant) (*Merchant, error) {
	var (
		merchant = Merchant{}
	)

	err := tx.Model(&Merchant{}).Preload("Logo").Where(where).Last(&merchant).Error
	if err != nil {
		utils.Error("unable to get merchant ", err)
		return nil, err
	}
	return &merchant, nil
}

// CreateWithTx implements IMerchant.
func (mr *merchantRepo) CreateWithTx(tx *gorm.DB, m *Merchant) error {
	err := tx.Model(&Merchant{}).
		Clauses(clause.Returning{
			Columns: []clause.Column{
				{
					Name: "uuid",
				},
				{
					Name: "id",
				},
			},
		}).Create(m).Error
	if err != nil {
		utils.Error("unable to create merchant ", err)
		return err
	}
	return nil
}

// Delete implements IMerchant.
func (*merchantRepo) Delete(where *Merchant) error {
	panic("unimplemented")
}

// DeleteWithTx implements IMerchant.
func (*merchantRepo) DeleteWithTx(where *Merchant) error {
	panic("unimplemented")
}

// GetByID implements IMerchant.
func (mr *merchantRepo) GetByID(id uint) (*Merchant, error) {
	var (
		merchant = Merchant{}
	)

	err := mr.db.Model(&Merchant{}).Where("id = ?", id).Last(&merchant).Error
	if err != nil {
		utils.Error("unable to get merchant ", err)
		return nil, err
	}
	return &merchant, nil
}

// Update implements IMerchant.
func (mr *merchantRepo) Update(where *Merchant, m *Merchant) error {
	return mr.UpdateWithTx(mr.db, where, m)
}

// UpdateWithTx implements IMerchant.
func (mr *merchantRepo) UpdateWithTx(tx *gorm.DB, where *Merchant, m *Merchant) error {
	m.UUID = where.UUID
	err := tx.
		Model(&m).
		Clauses(clause.Returning{
			Columns: []clause.Column{
				{
					Name: "uuid",
				},
				{
					Name: "id",
				},
			},
		}).
		Where(where).
		Updates(m).Error
	if err != nil {
		utils.Error("error in updating ", err)
		return err
	}
	return nil
}
