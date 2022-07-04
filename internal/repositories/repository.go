package repositories

import (
	"github.com/AyratB/go_diploma/internal/entities"
)

type Repository interface {
	RegisterUser(login, password string) error
	LoginUser(login, password string) error
	CheckOrderExists(orderNumber string) error
	SaveOrder(number, userLogin string) error
	GetUserOrders(userLogin string) ([]entities.OrderEntity, error)
	GetUserBalance(userLogin string) (*entities.UserBalance, error)
}
