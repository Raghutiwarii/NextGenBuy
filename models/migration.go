package models

// Add list of model add for migrations
var migrationModels = []interface{}{
	&Account{},
	&Email{},
	&UserRole{},
}

func GetMigrationModels() []interface{} {
	return migrationModels
}
