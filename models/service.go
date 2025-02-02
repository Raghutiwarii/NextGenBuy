package models

import "gorm.io/gorm"

func InitAccountRepo(db *gorm.DB) IAccount {
	return &accountRepo{
		db: db,
	}
}

func InitEmailRepo(db *gorm.DB) IEmailRepo {
	return &emailRepo{
		db: db,
	}
}

func InitCustomerRepo(db *gorm.DB) ICustomerRepo {
	return &customerRepo{
		db: db,
	}
}

func InitAddressRepo(db *gorm.DB) IAddress {
	return &addressRepo{
		db: db,
	}
}

func InitMerchantRepo(db *gorm.DB) IMerchant {
	return &merchantRepo{
		db: db,
	}
}