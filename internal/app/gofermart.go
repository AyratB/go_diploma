package app

import (
	"github.com/AyratB/go_diploma/internal/entities"
	"github.com/AyratB/go_diploma/internal/repositories"
	"github.com/AyratB/go_diploma/internal/utils"
	"sort"
)

type Gofermart struct {
	repo    repositories.Repository
	Decoder *utils.Decoder
}

func NewGofermart(repo repositories.Repository, decoder *utils.Decoder) *Gofermart {
	return &Gofermart{
		repo:    repo,
		Decoder: decoder,
	}
}

func (g Gofermart) RegisterUser(login, password string) error {
	return g.repo.RegisterUser(login, g.Decoder.Encode(password))
}

func (g Gofermart) LoginUser(login, password string) error {
	return g.repo.LoginUser(login, g.Decoder.Encode(password))
}

func (g Gofermart) CheckOrderExists(orderNumber string) error {
	return g.repo.CheckOrderExists(orderNumber)
}

func (g Gofermart) SaveOrder(orderNumber, userLogin string) error {
	return g.repo.SaveOrder(orderNumber, userLogin)
}

func (g Gofermart) GetUserOrders(userLogin string) ([]entities.OrderEntity, error) {
	orders, err := g.repo.GetUserOrders(userLogin)
	if err != nil {
		return nil, err
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.After(orders[j].UploadedAt)
	})

	return orders, err
}

func (g Gofermart) GetUserBalance(userLogin string) (*entities.UserBalance, error) {
	return g.repo.GetUserBalance(userLogin)
}
