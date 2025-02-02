package models

// Add list of model add for migrations
var migrationModels = []interface{}{
	&Account{},
	&Email{},
	&UserRole{},
	&Customer{},
	&Address{},
	&Merchant{},
}

func GetMigrationModels() []interface{} {
	return migrationModels
}
