package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AyratB/go_diploma/internal/app"
	"github.com/AyratB/go_diploma/internal/customerrors"
	"github.com/AyratB/go_diploma/internal/storage"
	"github.com/AyratB/go_diploma/internal/utils"
	"io"
	"net/http"
	"strconv"
	"time"
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

	newCookie := h.getCookie(rr.Login)

	http.SetCookie(w, newCookie)
	r.AddCookie(newCookie)

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

	newCookie := h.getCookie(rr.Login)

	http.SetCookie(w, newCookie)
	r.AddCookie(newCookie)

	w.WriteHeader(http.StatusOK)
}

func (h Handler) getCookie(userLogin string) *http.Cookie {
	return &http.Cookie{
		Name:  utils.CookieUserName,
		Value: h.gm.Decoder.Encode(userLogin),
		Path:  "/",
	}
}

func (h Handler) LoadUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	orderNumber, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	convertedOrderNumber, err := strconv.Atoi(string(orderNumber))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !utils.ValidOrderNumber(convertedOrderNumber) {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")

	userLogin := getUserLogin(r)

	err = h.gm.CheckOrderExists(string(orderNumber))
	if err != nil {
		if errors.As(err, &customerrors.ErrOrderNumberAlreadyBusy{}) {
			if castError, ok := err.(customerrors.ErrOrderNumberAlreadyBusy); ok {
				if castError.OrderUserLogin == userLogin { //200 — номер заказа уже был загружен этим пользователем;
					w.WriteHeader(http.StatusOK)
					return
				} else { // 409 — номер заказа уже был загружен другим пользователем;
					http.Error(w, castError.Error(), http.StatusConflict)
					return
				}
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.gm.SaveOrder(string(orderNumber), userLogin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 202 — новый номер заказа принят в обработку;
	w.WriteHeader(http.StatusAccepted)
}

func (h Handler) DecreaseBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
}

type GetUserOrdersResponse struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int64  `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

func (h Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	userOrders, err := h.gm.GetUserOrders(getUserLogin(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(userOrders) == 0 {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	responseOrders := make([]GetUserOrdersResponse, 0, len(userOrders))

	for _, order := range userOrders {

		responseOrder := GetUserOrdersResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}

		if order.Accrual.Valid {
			responseOrder.Accrual = order.Accrual.Int64
		}

		responseOrders = append(responseOrders, responseOrder)

	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp, err := json.Marshal(responseOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
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

func getUserLogin(r *http.Request) string {
	return fmt.Sprint(r.Context().Value(utils.KeyPrincipalID))
}
