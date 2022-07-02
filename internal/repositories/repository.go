package repositories

type Repository interface {
	RegisterUser(login, password string) error
}
