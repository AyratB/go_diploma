package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AyratB/go_diploma/internal/app"
	"github.com/AyratB/go_diploma/internal/customerrors"
	"github.com/AyratB/go_diploma/internal/entities"
	"github.com/AyratB/go_diploma/internal/storage"
	"github.com/AyratB/go_diploma/internal/utils"
	"github.com/go-resty/resty/v2"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Handler struct {
	Configs            *utils.Config
	gm                 *app.Gofermart
	externalHTTPClient http.Client
	HTTPClient         *resty.Client
	ProcessedOrders    chan entities.OrderQueueEntry
	NoProcessedOrders  chan entities.OrderQueueEntry
}

func NewHandler(ctx context.Context, configs *utils.Config, decoder *utils.Decoder, wg *sync.WaitGroup) (*Handler, error) {

	// TODO - comment only when local
	if len(configs.DatabaseURI) == 0 {
		return nil, errors.New("need Database URI")
	}

	// test local connection
	//configs.DatabaseURI = "postgres://postgres:test@localhost:5432/postgres?sslmode=disable"

	repo, err := storage.NewDBStorage(ctx, configs.DatabaseURI, wg)
	if err != nil {
		return nil, err
	}

	return &Handler{
		gm:                app.NewGofermart(repo, decoder),
		Configs:           configs,
		HTTPClient:        resty.New(),
		ProcessedOrders:   repo.ProcessedOrders,
		NoProcessedOrders: repo.NoProcessedOrders,
	}, nil
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
		http.Error(w, "not valid order number", http.StatusUnprocessableEntity)
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

	h.NoProcessedOrders <- entities.OrderQueueEntry{
		OrderNumber: string(orderNumber),
		OrderStatus: string(utils.New),
	}

	w.WriteHeader(http.StatusAccepted)
}

type decreaseRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (h Handler) DecreaseBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dr := decreaseRequest{}

	if err := json.Unmarshal(body, &dr); err != nil {
		http.Error(w, "Incorrect body JSON format", http.StatusBadRequest)
		return
	}
	if len(dr.Order) == 0 {
		http.Error(w, "order number can not be empty", http.StatusBadRequest)
		return
	}
	convertedOrderNumber, err := strconv.Atoi(dr.Order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !utils.ValidOrderNumber(convertedOrderNumber) {
		http.Error(w, "no valid number", http.StatusUnprocessableEntity)
		return
	}
	if dr.Sum <= 0 {
		http.Error(w, "sum must be greater than 0", http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	userLogin := getUserLogin(r)

	err = h.gm.DecreaseBalance(userLogin, dr.Order, dr.Sum)

	if err != nil {
		if errors.As(err, &customerrors.ErrNoEnoughMoney{}) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusOK)

}

type GetUserOrdersResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func (h Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("content-type", "application/json")

	userOrders, err := h.gm.GetUserOrders(getUserLogin(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(userOrders) == 0 {
		w.WriteHeader(http.StatusNoContent)
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
			responseOrder.Accrual = float32(order.Accrual.Float64)
		}

		responseOrders = append(responseOrders, responseOrder)
	}

	resp, err := json.Marshal(responseOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type GetUserBalanceResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func (h Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	userBalance, err := h.gm.GetUserBalance(getUserLogin(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if userBalance == nil {
		http.Error(w, "some error with users data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")

	response := GetUserBalanceResponse{
		Current:   userBalance.Current,
		Withdrawn: userBalance.SummaryWithdrawn,
	}

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type GetUserBalanceDecreases struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h Handler) GetUserBalanceDecreases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed by this route!", http.StatusMethodNotAllowed)
		return
	}

	userWithdrawals, err := h.gm.GetUserWithdrawals(getUserLogin(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(userWithdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userBalanceDecreases := make([]GetUserBalanceDecreases, 0, len(userWithdrawals))

	for _, userWithdrawal := range userWithdrawals {
		responseUserBalanceDecrease := GetUserBalanceDecreases{
			Order:       userWithdrawal.Order,
			Sum:         userWithdrawal.Sum,
			ProcessedAt: userWithdrawal.ProcessedAt.Format(time.RFC3339),
		}
		userBalanceDecreases = append(userBalanceDecreases, responseUserBalanceDecrease)
	}

	w.Header().Set("content-type", "application/json")

	resp, err := json.Marshal(userBalanceDecreases)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func getUserLogin(r *http.Request) string {
	return fmt.Sprint(r.Context().Value(utils.KeyPrincipalID))
}
