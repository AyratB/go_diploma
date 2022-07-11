package listener

import (
	"context"
	"encoding/json"
	"github.com/AyratB/go_diploma/internal/entities"
	"github.com/AyratB/go_diploma/internal/handlers"
	"github.com/AyratB/go_diploma/internal/utils"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Listener struct {
	ctx               context.Context
	noProcessedOrders chan entities.OrderQueueEntry
	processedOrders   chan entities.OrderQueueEntry
	wg                *sync.WaitGroup
	retryNumber       int

	externalClient *resty.Client
	externalAPIURL string
}

func NewListener(ctx context.Context, handler *handlers.Handler, wg *sync.WaitGroup) *Listener {
	return &Listener{
		ctx:               ctx,
		noProcessedOrders: handler.NoProcessedOrders,
		processedOrders:   handler.ProcessedOrders,
		wg:                wg,
		retryNumber:       5,
		externalClient:    handler.HTTPClient,
		externalAPIURL:    handler.Configs.AccrualSystemAddress + "/api/orders/{orderNumber}",
	}
}

func (l *Listener) ListenAndProcess() {

	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		l.processAsync()

		<-l.ctx.Done()

		close(l.processedOrders)
		close(l.noProcessedOrders)
	}()
}

func (l *Listener) processAsync() error {

	for noProcessedOrder := range l.noProcessedOrders {

		if noProcessedOrder.RetryAfter != 0 && time.Since(noProcessedOrder.LastChecked) < noProcessedOrder.RetryAfter {
			l.noProcessedOrders <- noProcessedOrder
			continue
		}

		for time.Since(noProcessedOrder.LastChecked) < 10*time.Second {
			select {
			case <-l.ctx.Done():
				return nil
			default:
			}
		}

		statusMap := map[string]string{
			"INVALID":    "INVALID",
			"PROCESSED":  "PROCESSED",
			"PROCESSING": "PROCESSING",
			"REGISTERED": "NEW",
		}

		// TEST
		// l.externalAPIURL = "http://localhost:8080/api/orders/{orderNumber}"

		response, err := l.externalClient.
			R().
			SetContext(l.ctx).
			SetPathParams(map[string]string{"orderNumber": noProcessedOrder.OrderNumber}).
			Get(l.externalAPIURL)

		sc := response.StatusCode()

		if err != nil ||
			(response != nil && sc != http.StatusTooManyRequests && sc != http.StatusOK) {

			noProcessedOrder.RetryCount += 1
			noProcessedOrder.LastChecked = time.Now()

			l.noProcessedOrders <- noProcessedOrder
			continue
		}

		if sc == http.StatusTooManyRequests {
			seconds, _ := strconv.Atoi(response.Header().Get("Retry-After"))

			noProcessedOrder.LastChecked = time.Now()
			noProcessedOrder.RetryAfter = time.Duration(int(time.Second) * seconds)

			l.noProcessedOrders <- noProcessedOrder
			continue
		}

		type externalAPIResponse struct {
			Order   string   `json:"order"`
			Status  string   `json:"status"`
			Accrual *float64 `json:"accrual"`
		}

		exrr := externalAPIResponse{}

		if err = json.Unmarshal(response.Body(), &exrr); err != nil {

			noProcessedOrder.RetryCount += 1

			noProcessedOrder.LastChecked = time.Now()
			noProcessedOrder.RetryAfter = 0

			l.processedOrders <- noProcessedOrder
			continue
		}

		// статус прежний
		if statusMap[exrr.Status] == noProcessedOrder.OrderStatus {

			noProcessedOrder.LastChecked = time.Now()
			noProcessedOrder.RetryAfter = 0

			l.noProcessedOrders <- noProcessedOrder
		} else {

			// обновляем временный результат
			l.processedOrders <- entities.OrderQueueEntry{
				OrderNumber: exrr.Order,
				OrderStatus: exrr.Status,
				Accrual:     exrr.Accrual,
			}

			// неокончательный результат
			if statusMap[exrr.Status] != string(utils.Processed) && statusMap[exrr.Status] != string(utils.Invalid) {

				noProcessedOrder.LastChecked = time.Now()
				noProcessedOrder.RetryAfter = 0

				l.noProcessedOrders <- noProcessedOrder
			}
		}
	}
	return nil
}
