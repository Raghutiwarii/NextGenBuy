package models

// Add list of model add for migrations
var migrationModels = []interface{}{
	&Account{},
	&Credential{},
	&Email{},
}

func GetMigrationModels() []interface{} {
	return migrationModels
}
