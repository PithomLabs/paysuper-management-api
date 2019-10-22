package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/paylink"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	paylinksPath               = "/paylinks"
	paylinksIdPath             = "/paylinks/:id"
	paylinksUrlPath            = "/paylinks/:id/url"
	paylinksIdStatSummaryPath  = "/paylinks/:id/dashboard/summary"
	paylinksIdStatCountryPath  = "/paylinks/:id/dashboard/country"
	paylinksIdStatReferrerPath = "/paylinks/:id/dashboard/referrer"
	paylinksIdStatDatePath     = "/paylinks/:id/dashboard/date"
	paylinksIdStatUtmPath      = "/paylinks/:id/dashboard/utm"
)

type PayLinkRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPayLinkRoute(set common.HandlerSet, cfg *common.Config) *PayLinkRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PayLinkRoute"})
	return &PayLinkRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *PayLinkRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(paylinksPath, h.getPaylinksList)
	groups.AuthUser.GET(paylinksIdPath, h.getPaylink)
	groups.AuthUser.GET(paylinksUrlPath, h.getPaylinkUrl)
	groups.AuthUser.DELETE(paylinksIdPath, h.deletePaylink)
	groups.AuthUser.POST(paylinksPath, h.createPaylink)
	groups.AuthUser.PUT(paylinksIdPath, h.updatePaylink)
	groups.AuthUser.GET(paylinksIdStatSummaryPath, h.getPaylinkStatSummary)
	groups.AuthUser.GET(paylinksIdStatCountryPath, h.getPaylinkStatByCountry)
	groups.AuthUser.GET(paylinksIdStatReferrerPath, h.getPaylinkStatByReferrer)
	groups.AuthUser.GET(paylinksIdStatDatePath, h.getPaylinkStatByDate)
	groups.AuthUser.GET(paylinksIdStatUtmPath, h.getPaylinkStatByUtm)
}

// @Description Get list of paylinks for authenticated merchant
// @Example GET /admin/api/v1/paylinks?offset=0&limit=10
func (h *PayLinkRoute) getPaylinksList(ctx echo.Context) error {
	req := &grpc.GetPaylinksRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.ProjectId = ""

	if req.Limit == 0 {
		req.Limit = h.cfg.LimitDefault
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinks(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinks", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Data)
}

// @Description Get paylink, for authenticated merchant
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22
func (h *PayLinkRoute) getPaylink(ctx echo.Context) error {
	req := &grpc.PaylinkRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylink(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description paylink public url
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22/url?utm_source=3wefwe&utm_medium=njytrn&utm_campaign=bdfbh5
func (h *PayLinkRoute) getPaylinkUrl(ctx echo.Context) error {
	req := &grpc.GetPaylinkURLRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id
	req.UrlMask = pkg.PaylinkUrlDefaultMask

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkURL(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkURL", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Url)
}

// @Description Get paylink, for authenticated merchant
// @Example DELETE /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22
func (h *PayLinkRoute) deletePaylink(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req := &grpc.PaylinkRequest{
		Id:         ctx.Param(common.RequestParameterId),
		MerchantId: merchant.Item.Id,
	}
	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeletePaylink(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeletePaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Description Create paylink, for authenticated merchant
// @Example POST /admin/api/v1/paylinks
func (h *PayLinkRoute) createPaylink(ctx echo.Context) error {
	return h.createOrUpdatePaylink(ctx, "")
}

// @Description Update paylink, for authenticated merchant
// @Example PUT /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22
func (h *PayLinkRoute) updatePaylink(ctx echo.Context) error {
	return h.createOrUpdatePaylink(ctx, ctx.Param(common.RequestParameterId))
}

func (h *PayLinkRoute) createOrUpdatePaylink(ctx echo.Context, paylinkId string) error {
	req := &paylink.CreatePaylinkRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.MerchantId = merchant.Item.Id
	req.Id = paylinkId

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdatePaylink(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "CreateOrUpdatePaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description paylink stat summary
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22/stat/summary?period_from=1571671243&period_to=1571673643
func (h *PayLinkRoute) getPaylinkStatSummary(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatTotal(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatTotal", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description paylink stat by country
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22/stat/by/country?period_from=1571671243&period_to=1571673643
func (h *PayLinkRoute) getPaylinkStatByCountry(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByCountry(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByCountry", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description paylink stat by referrer
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22/stat/by/referrer?period_from=1571671243&period_to=1571673643
func (h *PayLinkRoute) getPaylinkStatByReferrer(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByReferrer(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByReferrer", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description paylink stat by date
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22/stat/by/date?period_from=1571671243&period_to=1571673643
func (h *PayLinkRoute) getPaylinkStatByDate(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByDate(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByDate", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description paylink stat by utm
// @Example GET /admin/api/v1/paylinks/21784001599a47e5a69ac28f7af2ec22/stat/by/utm?period_from=1571671243&period_to=1571673643
func (h *PayLinkRoute) getPaylinkStatByUtm(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByUtm(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByUtm", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
