package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"net/http"
)

const (
	projectsPath    = "/projects"
	projectsIdPath  = "/projects/:project_id"
	projectsSkuPath = "/projects/:project_id/sku"
)

type ProjectRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewProjectRoute(set common.HandlerSet, cfg *common.Config) *ProjectRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "ProjectRoute"})
	return &ProjectRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *ProjectRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(projectsPath, h.listProjects)
	groups.AuthUser.GET(projectsIdPath, h.getProject)
	groups.AuthUser.POST(projectsPath, h.createProject)
	groups.AuthUser.PATCH(projectsIdPath, h.updateProject)
	groups.AuthUser.DELETE(projectsIdPath, h.deleteProject)
	groups.AuthUser.POST(projectsSkuPath, h.checkSku)
}

func (h *ProjectRoute) createProject(ctx echo.Context) error {
	req := &billingpb.Project{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	if req.MerchantId != authUser.MerchantId {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectMerchantId)
	}

	if len(req.CallbackProtocol) == 0 {
		req.CallbackProtocol = billingpb.ProjectCallbackProtocolEmpty
	}

	// vat payer is seller by default on project creation
	req.VatPayer = billingpb.VatPayerSeller

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ChangeProject(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != billingpb.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusCreated, res.Item)
}

func (h *ProjectRoute) updateProject(ctx echo.Context) error {
	req := &billingpb.Project{}
	binder := common.NewChangeProjectRequestBinder(h.dispatch, h.cfg)

	if err := binder.Bind(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	authUser := common.ExtractUserContext(ctx)
	if req.MerchantId != authUser.MerchantId {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectMerchantId)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ChangeProject(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != billingpb.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *ProjectRoute) getProject(ctx echo.Context) error {
	req := &billingpb.GetProjectRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetProject(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != billingpb.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ProjectRoute) listProjects(ctx echo.Context) error {
	req := &billingpb.ListProjectsRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = h.cfg.LimitDefault
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListProjects(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ProjectRoute) deleteProject(ctx echo.Context) error {
	req := &billingpb.GetProjectRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeleteProject(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != billingpb.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}
func (h *ProjectRoute) checkSku(ctx echo.Context) error {
	req := &billingpb.CheckSkuAndKeyProjectRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CheckSkuAndKeyProject(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != billingpb.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}
