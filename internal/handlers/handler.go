package handlers

import (
	"errors"
	"github.com/AyratB/go_diploma/internal/app"
	"github.com/AyratB/go_diploma/internal/storage"
	"github.com/AyratB/go_diploma/internal/utils"
	"net/http"
)

type Handler struct {
	configs *utils.Config
	gm      *app.Gofermart
}

func (h Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func (h Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func (h Handler) LoadUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func (h Handler) DecreaseBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func (h Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func (h Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func (h Handler) GetUserBalanceDecreases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

func NewHandler(configs *utils.Config) (*Handler, func() error, error) {

	if len(configs.DatabaseURI) == 0 {
		return nil, nil, errors.New("need Database URI")
	}

	repo, err := storage.NewDBStorage(configs.DatabaseURI)
	if err != nil {
		return nil, nil, err
	}

	return &Handler{
		gm:      app.NewGofermart(repo),
		configs: configs,
	}, repo.CloseResources, nil
}
