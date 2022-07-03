package customerrors

import "errors"

var ErrDuplicateUserLogin = errors.New("user with same login already exists")
var ErrNoUserByLoginAndPassword = errors.New("no user by this login/password")
var ErrInvalidCookie = errors.New("crumbled cookie. Clear and relogin, please")
