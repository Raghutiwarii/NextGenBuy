package constants

import "os"

var JWT_SECRET = []byte(os.Getenv("JWT_SECRET_KEY"))
var PASSWORD_SECRET = []byte(os.Getenv("PASSWORD_SECRET"))
