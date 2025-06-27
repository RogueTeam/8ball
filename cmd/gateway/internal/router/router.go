package router

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"anarchy.ttfm/8ball/gateway"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Manages the entire setup of the Gateway service
type Router struct {
	// Process interval
	ProcessInterval time.Duration
	// Gateway controller
	Gateway *gateway.Controller
	// Base Gin Group to use for routing
	Base gin.IRoutes
}

const (
	IdParam            = "id"
	PaymentsPath       = "/payments"
	PaymentsPathWithId = PaymentsPath + "/:" + IdParam
)

func (r *Router) createPayment(ctx *gin.Context) {
	var receive Receive
	err := ctx.BindJSON(&receive)
	if err != nil {
		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	gatewayReceive, err := ReceiveToGateway(&receive)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	payment, err := r.Gateway.Receive(ctx, &gatewayReceive)
	switch {
	case err == nil:
		out := PaymentFromGateway(&payment)
		ctx.JSON(http.StatusCreated, &out)
	default:
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (r *Router) paymentStatus(ctx *gin.Context) {
	rawId := ctx.Param(IdParam)
	id, err := uuid.Parse(rawId)
	if err != nil {
		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	payment, err := r.Gateway.Query(ctx, id)
	switch {
	case err == nil:
		out := PaymentFromGateway(&payment)
		ctx.JSON(http.StatusCreated, &out)
	case errors.Is(err, gateway.ErrPaymentNotFound):
		ctx.AbortWithError(http.StatusNotFound, err)
	default:
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}
}

// Register routes in the Gin engine
func (r *Router) Register() {
	r.Base.POST(PaymentsPath, r.createPayment)
	r.Base.GET(PaymentsPathWithId, r.paymentStatus)

	go func() {
		ticker := time.NewTicker(r.ProcessInterval)
		defer ticker.Stop()

		for {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				processed, err := r.Gateway.ProcessPendingPayments()
				if err != nil {
					log.Println("ERROR|PROCESSING|PAYMENTS", err)
				}
				log.Println("INFO|PROCESSED|PAYMENTS", processed)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				processed, err := r.Gateway.ProcessPendingFees()
				if err != nil {
					log.Println("ERROR|PROCESSING|FEES", err)
				}
				log.Println("INFO|PROCESSED|FEES", processed)
			}()
			wg.Wait()
			<-ticker.C
		}
	}()
}
