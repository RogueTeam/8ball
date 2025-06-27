package router

import (
	"errors"
	"net/http"

	"anarchy.ttfm/8ball/gateway"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Manages the entire setup of the Gateway service
type Router struct {
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
func (r *Router) Register() (err error) {
	r.Base.POST(PaymentsPath, r.createPayment)
	r.Base.GET(PaymentsPathWithId, r.paymentStatus)
	return nil
}
