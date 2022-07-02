package handlers

import (
	"encoding/json"
	"errors"
	"github.com/AyratB/go_diploma/internal/app"
	"github.com/AyratB/go_diploma/internal/customerrors"
	"github.com/AyratB/go_diploma/internal/storage"
	"github.com/AyratB/go_diploma/internal/utils"
	"io"
	"net/http"
)

type Handler struct {
	configs *utils.Config
	gm      *app.Gofermart
}

func NewHandler(configs *utils.Config, decoder *utils.Decoder) (*Handler, func() error, error) {

	// TODO - comment only when local
	//if len(configs.DatabaseURI) == 0 {
	//	return nil, nil, errors.New("need Database URI")
	//}

	// test local connection
	configs.DatabaseURI = "postgres://postgres:test@localhost:5432/postgres?sslmode=disable"

	repo, err := storage.NewDBStorage(configs.DatabaseURI)
	if err != nil {
		return nil, nil, err
	}

	return &Handler{
		gm:      app.NewGofermart(repo, decoder),
		configs: configs,
	}, repo.CloseResources, nil
}

type registerOrLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rr := registerOrLoginRequest{}

	if err := json.Unmarshal(b, &rr); err != nil {
		http.Error(w, "Incorrect body JSON format", http.StatusBadRequest)
		return
	}
	if len(rr.Login) == 0 {
		http.Error(w, "Login can not be empty", http.StatusBadRequest)
		return
	}
	if len(rr.Password) == 0 {
		http.Error(w, "Password can not be empty", http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	if err = h.gm.RegisterUser(rr.Login, rr.Password); err != nil {
		if errors.Is(err, customerrors.ErrDuplicateUserLogin) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// TODO - После успешной регистрации автоматическая аутентификация пользователя

	w.WriteHeader(http.StatusOK)
}

func (h Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rr := registerOrLoginRequest{}

	if err := json.Unmarshal(b, &rr); err != nil {
		http.Error(w, "Incorrect body JSON format", http.StatusBadRequest)
		return
	}
	if len(rr.Login) == 0 {
		http.Error(w, "Login can not be empty", http.StatusBadRequest)
		return
	}
	if len(rr.Password) == 0 {
		http.Error(w, "Password can not be empty", http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	if err = h.gm.LoginUser(rr.Login, rr.Password); err != nil {
		if errors.Is(err, customerrors.ErrNoUserByLoginAndPassword) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// TODO - После успешной регистрации автоматическая аутентификация пользователя - добавить куки

	w.WriteHeader(http.StatusOK)
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

func (h Handler) GetOrdersPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}
