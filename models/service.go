package models

import "gorm.io/gorm"

func InitAccountRepo(db *gorm.DB) IAccount {
	return &accountRepo{
		db: db,
	}
}

func InitCredentialRepo(db *gorm.DB) IAuthCredential {
	return &credentialRepo{
		DB: db,
	}
}

func InitEmailRepo(db *gorm.DB) IEmailRepo {
	return &emailRepo{
		db: db,
	}
}
