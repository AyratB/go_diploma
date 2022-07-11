package server

import (
	"context"
	"github.com/AyratB/go_diploma/internal/handlers"
	"github.com/AyratB/go_diploma/internal/middlewares"
	"github.com/AyratB/go_diploma/internal/service/listener"
	"github.com/AyratB/go_diploma/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"sync"
)

func Run(configs *utils.Config, ctx context.Context) error {

	r := chi.NewRouter()

	decoder := utils.NewDecoder()
	cookieHandler := middlewares.NewCookieHandler(decoder)

	r.Use(middleware.RequestID)
	r.Use(middlewares.GzipHandle)
	r.Use(cookieHandler.CookieHandler)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	wg := &sync.WaitGroup{}

	handler, err := handlers.NewHandler(ctx, configs, decoder, wg)
	if err != nil {
		return err
	}

	listener := listener.NewListener(ctx, handler, wg)
	listener.ListenAndProcess()

	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", handler.RegisterUser)
		r.Post("/api/user/login", handler.LoginUser)
		//r.Get("/api/orders/{number}", handler.GetOrdersPoints)

		// эти запросы закрыты для неавторизованных пользователей
		r.Post("/api/user/orders", handler.LoadUserOrders)
		r.Get("/api/user/orders", handler.GetUserOrders)
		r.Get("/api/user/balance", handler.GetUserBalance)
		r.Post("/api/user/balance/withdraw", handler.DecreaseBalance)
		r.Get("/api/user/withdrawals", handler.GetUserBalanceDecreases)
	})
	return http.ListenAndServe(configs.RunAddress, r)
}
