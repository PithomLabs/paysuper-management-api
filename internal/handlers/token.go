package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	tokenPath = "/tokens"
)

type TokenRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

type TokenCreationResponse struct {
	// The secure string which contains encrypted information about your customer and sales option data.
	Token string `json:"token"`
	// The PaySuper-hosted URL of a payment form.
	PaymentFormUrl string `json:"payment_form_url"`
}

func NewTokenRoute(set common.HandlerSet, cfg *common.Config) *TokenRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "TokenRoute"})
	return &TokenRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *TokenRoute) Route(groups *common.Groups) {
	groups.Common.POST(tokenPath, h.createToken)
}

// @summary Create a token
// @desc Create a token that encrypts details of your customer, the game and purchase parameters.
// @id tokenPathCreateToken
// @tag Token
// @accept application/json
// @produce application/json
// @body grpc.TokenRequest
// @success 200 {object} TokenCreationResponse Returns the token string and the PaySuper-hosted URL for a payment form
// @failure 400 {object} grpc.ResponseErrorMessage Invalid request data
// @failure 404 {object} grpc.ResponseErrorMessage Not found
// @failure 500 {object} grpc.ResponseErrorMessage Internal Server Error
// @router /api/v1/tokens [post]
func (h *TokenRoute) createToken(ctx echo.Context) error {
	req := &grpc.TokenRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	err = common.CheckProjectAuthRequestSignature(h.dispatch, ctx, req.Settings.ProjectId)

	if err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.CreateToken(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	response := map[string]string{
		"token":            res.Token,
		"payment_form_url": h.cfg.OrderInlineFormUrlMask + "?token=" + res.Token,
	}

	return ctx.JSON(http.StatusOK, response)
}
