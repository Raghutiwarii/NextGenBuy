package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IAccount interface {
	GetByID(userID uint) (*Account, error)
	GetByEmail(email string) ([]*Account, error)
	Get(where *Account) (*Account, error)
	GetWithTx(tx *gorm.DB, where *Account) (*Account, error)
	Create(u *Account) error
	CreateWithTx(tx *gorm.DB, u *Account) error
	Update(where *Account, a *Account) error
	UpdateWithTx(tx *gorm.DB, where *Account, a *Account) error
	Delete(userID uint) error
	DeleteWithTx(tx *gorm.DB, u *Account) error
	FindOne(email, phoneNumber, accountUUID string) (*Account, error)
	GetAllAccountsWithSameMailDomain(tx *gorm.DB, emailDomain string) (*[]Account, error)
	CheckAccountAssociatedForTheEmail(accountUUID, email string) (*Email, error)
	MarkAccountAsNotVerified(where *Account) error
	BulkInsert(records *[]Account, conds ...clause.Expression) error
	AddEmails(a *Account, emails []Email) error
}

type IEmailRepo interface {
	Create(i *Email) error
	CreateWithTx(tx *gorm.DB, i *Email) error
	Get(where *Email) (*Email, error)
	GetWithTx(tx *gorm.DB, where *Email) (*Email, error)
	Update(where *Email, i *Email) error
	UpdateWithTx(tx *gorm.DB, where *Email, i *Email) error
	Delete(where *Email) error
	DeleteWithTx(tx *gorm.DB, where *Email) error
	GetWithAccount(where *Email) (*Email, error)
	UpdateEmailToVerified(tx *gorm.DB, accountID uint, email *Email) error
	CreateAccountEmail(e *Email) (*Email, error)
	CreateAccountEmailWithTx(tx *gorm.DB, u *Email) (*Email, error)
	GetAll(where *Email) (*[]Email, error)
	GetLastEmailOfAccount(accountID uint) (*Email, error)
}

type IAddress interface {
	Create(a *Address) error
	CreateWithTX(tx *gorm.DB, a *Address) error
	Get(where *Address) (*Address, error)
	GetAll(where *Address) ([]Address, error)
	GetWithTx(tx *gorm.DB, where *Address) (*Address, error)
	Update(where *Address, a *Address) error
	UpdateWithTx(tx *gorm.DB, where *Address, a *Address) error
}

type IMerchant interface {
	Get(where *Merchant) (*Merchant, error)
	GetByID(id uint) (*Merchant, error)
	GetWithTX(tx *gorm.DB, where *Merchant) (*Merchant, error)
	Create(m *Merchant) error
	GetByUUID(merchantUUID string) (*Merchant, error)
	CreateWithTx(tx *gorm.DB, m *Merchant) error
	Update(where *Merchant, m *Merchant) error
	UpdateWithTx(tx *gorm.DB, where *Merchant, m *Merchant) error
	Delete(where *Merchant) error
	DeleteWithTx(where *Merchant) error
}

type ICustomerRepo interface {
	Create(c *Customer) error
	CreateWithTx(tx *gorm.DB, c *Customer) error
	Get(where *Customer) (*Customer, error)
}

type IOTPRepo interface {
	Create(otp *OTP) error
	Get(where *OTP) (*OTP, error)
	Update(where *OTP, o *OTP) error
	UpdateWithTx(tx *gorm.DB, where *OTP, o *OTP) error
}

type IProductRepo interface {
	Create(p *Product) error
	CreateWithTx(tx *gorm.DB, p *Product) error
	Delete(where *Product) error
	DeleteWithTx(tx *gorm.DB, where *Product) error
	GetByCategory(category string) ([]Product, error)
	GetByID(id uint) (*Product, error)
	GetByUUID(uuid string) (*Product, error)
	Update(where *Product, p *Product) error
	UpdateWithTx(tx *gorm.DB, where *Product, p *Product) error
	Get(where *Product) (*Product, error)
	GetWithTx(tx *gorm.DB, where *Product) (*Product, error)
	CreateInBatches(products []*Product, batchSize int, merchantID string) error
}

type ICheckoutRepo interface {
	Create(c *Checkout) error
	CreateWithTx(tx *gorm.DB, c *Checkout) error
	GetByID(checkoutId string) (*Checkout, error)
	Get(where *Checkout) (*Checkout, error)
	GetWithTx(tx *gorm.DB, where *Checkout) (*Checkout, error)
}

type IOrderRepo interface {
	Create(o *Order) error
	CreateWithTx(tx *gorm.DB, o *Order) error
	Get(where *Order) (*Order, error)
	GetWithTx(tx *gorm.DB, where *Order) (*Order, error)
	GetWithPreloads(where *Order) (*Order, error)
	GetCount(where *Order) (int, error)
	GetTotalOfferAmount(where *Order) (*int64, error)
	Update(where *Order, c *Order) error
	UpdateWithTx(tx *gorm.DB, where *Order, c *Order) error
	GetOrderWithIds(ids []int) (*[]string, error)
}
