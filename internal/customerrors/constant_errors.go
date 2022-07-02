package customerrors

import "errors"

var ErrDuplicateUserLogin = errors.New("user with same login already exists")
