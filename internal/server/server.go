package server

import (
	"github.com/AyratB/go_diploma/internal/handlers"
	"github.com/AyratB/go_diploma/internal/middlewares"
	"github.com/AyratB/go_diploma/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func Run(configs *utils.Config) (func() error, error) {

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

	handler, resourcesCloser, err := handlers.NewHandler(configs, decoder)
	if err != nil {
		return resourcesCloser, err
	}

	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", handler.RegisterUser)
		r.Post("/api/user/login", handler.LoginUser)

		// эти запросы должны быть закрыты для неавторизованных пользователей
		r.Post("/api/user/orders", handler.LoadUserOrders)
		r.Post("/api/user/balance/withdraw", handler.DecreaseBalance)

		r.Get("/api/user/orders", handler.GetUserOrders)
		r.Get("/api/user/balance", handler.GetUserBalance)
		r.Get("/api/user/balance/withdrawals", handler.GetUserBalanceDecreases)
		r.Get("/api/orders/{number}", handler.GetOrdersPoints)
	})
	return resourcesCloser, http.ListenAndServe(configs.RunAddress, r)
}
