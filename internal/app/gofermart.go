package app

import "github.com/AyratB/go_diploma/internal/repositories"

type Gofermart struct {
	repo repositories.Repository
}

func NewGofermart(repo repositories.Repository) *Gofermart {
	return &Gofermart{repo: repo}
}
