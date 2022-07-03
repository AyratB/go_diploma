package repositories

type Repository interface {
	RegisterUser(login, password string) error
	LoginUser(login, password string) error
	CheckOrderExists(orderNumber string) error
	SaveOrder(number, userLogin string) error
}
