package app

import (
	"github.com/AyratB/go_diploma/internal/repositories"
	"github.com/AyratB/go_diploma/internal/utils"
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
