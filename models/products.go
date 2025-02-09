package models

import (
	utils "ecom/backend/utils"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Product struct {
	gorm.Model
	UUID           string         `gorm:"unique" json:"uuid,omitempty"`
	MerchantID     string         `json:"merchant_id" gorm:"index"`
	Title          string         `json:"title" gorm:"not null"`
	Description    string         `json:"description,omitempty"`
	Price          uint           `json:"price" gorm:"not null"`
	Stock          uint           `json:"stock" gorm:"default:0"`
	Category       string         `json:"category"`
	ImageURL       string         `json:"image_url,omitempty"`
	IsActive       *bool          `json:"is_active" gorm:"default:false"`
	Specifications datatypes.JSON `json:"specifications" gorm:"type:jsonb"`
	// Offers      []Offer  `json:"offers,omitempty" gorm:"foreignKey:UUID;references:ProductID"`
	Merchant Merchant `json:"merchant,omitempty" gorm:"foreignKey:MerchantID;references:UUID"`
}

type productRepo struct {
	db *gorm.DB
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	// Generate a unique product UUID
	productId, err := utils.GenerateNanoID(12, "p_")
	if err != nil {
		utils.Error("unable to generate nano id ", err)
		return err
	}
	p.UUID = productId

	tx.Statement.AddClause(clause.Returning{Columns: []clause.Column{
		{Name: "uuid"},
		{Name: "id"},
		{Name: "merchant_id"},
	}})

	return nil
}

// Create a new product
func (pr *productRepo) Create(p *Product) error {
	return pr.CreateWithTx(pr.db, p)
}

// Create a product within a transaction
func (pr *productRepo) CreateWithTx(tx *gorm.DB, p *Product) error {
	err := tx.Model(&Product{}).
		Clauses(clause.Returning{Columns: []clause.Column{
			{Name: "uuid"},
			{Name: "id"},
		}}).Create(p).Error
	if err != nil {
		utils.Error("unable to create product ", err)
		return err
	}
	return nil
}

// Get product by ID
func (pr *productRepo) GetByID(id uint) (*Product, error) {
	var product Product
	err := pr.db.Model(&Product{}).Where("id = ?", id).Last(&product).Error
	if err != nil {
		utils.Error("unable to get product ", err)
		return nil, err
	}
	return &product, nil
}

// Get product by UUID
func (pr *productRepo) GetByUUID(uuid string) (*Product, error) {
	var product Product
	err := pr.db.Model(&Product{}).Where("uuid = ?", uuid).Last(&product).Error
	if err != nil {
		utils.Error("unable to get product ", err)
		return nil, err
	}
	return &product, nil
}

// Get products by category
func (pr *productRepo) GetByCategory(category string) ([]Product, error) {
	var products []Product
	err := pr.db.Model(&Product{}).Where("category = ?", category).Find(&products).Error
	if err != nil {
		utils.Error("unable to get products by category ", err)
		return nil, err
	}
	return products, nil
}

// Update product details
func (pr *productRepo) Update(where *Product, p *Product) error {
	return pr.UpdateWithTx(pr.db, where, p)
}

// Update product within a transaction
func (pr *productRepo) UpdateWithTx(tx *gorm.DB, where *Product, p *Product) error {
	p.UUID = where.UUID
	err := tx.Model(&p).
		Clauses(clause.Returning{Columns: []clause.Column{
			{Name: "uuid"},
			{Name: "id"},
		}}).
		Where(where).
		Updates(p).Error
	if err != nil {
		utils.Error("error in updating product ", err)
		return err
	}
	return nil
}

// Delete product by ID
func (pr *productRepo) Delete(where *Product) error {
	return pr.DeleteWithTx(pr.db, where)
}

// Delete product within a transaction
func (pr *productRepo) DeleteWithTx(tx *gorm.DB, where *Product) error {
	err := tx.Where(where).Delete(&Product{}).Error
	if err != nil {
		utils.Error("error in deleting product ", err)
		return err
	}
	return nil
}

func (pr *productRepo) Get(where *Product) (*Product, error) {
	return pr.GetWithTx(pr.db, where)
}

func (pr *productRepo) GetWithTx(tx *gorm.DB, where *Product) (*Product, error) {
	var (
		p = Product{}
	)
	err := tx.Model(&Product{}).
		Where(where).
		Scopes(OmitIDToDeletedAtFields).
		Last(&p).Error
	if err != nil {
		utils.Error("unable to query product ", err)
		return nil, err
	}
	return &p, nil
}

func (pr *productRepo) CreateInBatches(products []*Product, batchSize int, merchantID string) error {
	return pr.CreateInBatchesWithTx(pr.db, products, batchSize, merchantID)
}

func (pr *productRepo) CreateInBatchesWithTx(tx *gorm.DB, products []*Product, batchSize int, merchantID string) error {
	if merchantID == "" {
		return fmt.Errorf("merchantID cannot be empty")
	}

	// Set MerchantID for all products
	for _, p := range products {
		p.MerchantID = merchantID
	}

	err := tx.Model(&Product{}).
		Clauses(clause.Returning{Columns: []clause.Column{
			{Name: "uuid"},
			{Name: "id"},
		}}).
		CreateInBatches(products, batchSize).Error

	if err != nil {
		utils.Error("unable to create products in batch ", err)
		return err
	}

	return nil
}
