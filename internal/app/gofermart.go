package app

import (
	"github.com/AyratB/go_diploma/internal/repositories"
	"github.com/AyratB/go_diploma/internal/utils"
)

type Gofermart struct {
	repo    repositories.Repository
	decoder *utils.Decoder
}

func (g Gofermart) RegisterUser(login, password string) error {
	return g.repo.RegisterUser(login, g.decoder.Encode(password))
}

func NewGofermart(repo repositories.Repository, decoder *utils.Decoder) *Gofermart {
	return &Gofermart{
		repo:    repo,
		decoder: decoder,
	}
}
