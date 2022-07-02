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

func (h Handler) RegisterUser(writer http.ResponseWriter, request *http.Request) {

}

func (h Handler) LoginUser(writer http.ResponseWriter, request *http.Request) {

}

func (h Handler) LoadUserOrders(writer http.ResponseWriter, request *http.Request) {

}

func (h Handler) DecreaseBalance(writer http.ResponseWriter, request *http.Request) {

}

func (h Handler) GetUserOrders(writer http.ResponseWriter, request *http.Request) {

}

func (h Handler) GetUserBalance(writer http.ResponseWriter, request *http.Request) {

}

func (h Handler) GetUserBalanceDecreases(writer http.ResponseWriter, request *http.Request) {

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
