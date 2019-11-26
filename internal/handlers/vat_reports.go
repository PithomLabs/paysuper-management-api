package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
	"strings"
)

const (
	vatReportsPath        = "/vat_reports"
	vatReportsCountryPath = "/vat_reports/country/:country"
	vatReportsDetailsPath = "/vat_reports/details/:id"
	vatReportsStatusPath  = "/vat_reports/status/:id"
)

type VatReportsRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewVatReportsRoute(set common.HandlerSet, cfg *common.Config) *VatReportsRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "VatReportsRoute"})
	return &VatReportsRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *VatReportsRoute) Route(groups *common.Groups) {
	groups.SystemUser.GET(vatReportsPath, h.getVatReportsDashboard)
	groups.SystemUser.GET(vatReportsCountryPath, h.getVatReportsForCountry)
	groups.SystemUser.GET(vatReportsDetailsPath, h.getVatReportTransactions)
	groups.SystemUser.POST(vatReportsStatusPath, h.updateVatReportStatus)
}

func (h *VatReportsRoute) getVatReportsDashboard(ctx echo.Context) error {

	res, err := h.dispatch.Services.Billing.GetVatReportsDashboard(ctx.Request().Context(), &grpc.EmptyRequest{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *VatReportsRoute) getVatReportsForCountry(ctx echo.Context) error {
	req := &grpc.VatReportsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	req.Country = strings.ToUpper(ctx.Param(common.RequestParameterCountry))

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetVatReportsForCountry(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *VatReportsRoute) getVatReportTransactions(ctx echo.Context) error {
	req := &grpc.VatTransactionsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	req.VatReportId = ctx.Param(common.RequestParameterId)

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetVatReportTransactions(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *VatReportsRoute) updateVatReportStatus(ctx echo.Context) error {

	req := &grpc.UpdateVatReportStatusRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	req.Id = ctx.Param(common.RequestParameterId)

	if err = h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.UpdateVatReportStatus(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}
