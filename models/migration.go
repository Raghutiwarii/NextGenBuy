package models

// Add list of model add for migrations
var migrationModels = []interface{}{
	&Account{},
	&Email{},
	&UserRole{},
	&Customer{},
	&Address{},
	&Merchant{},
	&OTP{},
}

func GetMigrationModels() []interface{} {
	return migrationModels
}
