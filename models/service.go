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

func InitUserRoleRepo(db *gorm.DB) IUserRoleRepo {
	return &userRoleRepo{
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
