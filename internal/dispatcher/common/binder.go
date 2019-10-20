package common

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/globalsign/mgo/bson"
	"github.com/gurukami/typ/v2"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-payment-link/proto/paylink"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

)

var (
	SystemBinderDefault   = &SystemBinder{}
	MerchantBinderDefault = &MerchantBinder{}
	BinderDefault         = &Binder{}
	EchoBinderDefault     = &echo.DefaultBinder{}
)

const (
	MerchantIdField    = "MerchantId"
	MerchantSliceField = "Merchant"
	ParamTag           = "param"
)

type SystemBinder struct{}

func (b *SystemBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	rv := reflect.ValueOf(i)

	if rv.Type().Kind() != reflect.Ptr || rv.IsNil() {
		return ErrorInternal
	}

	irv := rv.Elem()
	irt := irv.Type()

	if irt.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < irv.NumField(); i++ {

		rv := irv.Field(i)
		tf := irt.Field(i)

		if v, ok := tf.Tag.Lookup(ParamTag); ok {
			rv.Set(reflect.ValueOf(ctx.Param(v)))
		}

		if strings.EqualFold(tf.Name, MerchantIdField) {
			if tf.Type.Kind() != reflect.String {
				return ErrorInternal
			}
			rv.Set(reflect.ValueOf(ctx.Param(RequestParameterMerchantId)))
		}

		if strings.EqualFold(tf.Name, MerchantSliceField) {
			if tf.Type.Kind() != reflect.Slice {
				return ErrorInternal
			}
			if rv.Type().Elem().Kind() == reflect.String {
				rv.Set(reflect.ValueOf([]string{ctx.Param(RequestParameterMerchantId)}))
			}
		}
	}

	return nil
}

type MerchantBinder struct{}

func (b *MerchantBinder) Bind(i interface{}, ctx echo.Context) (err error) {

	rv := reflect.ValueOf(i)

	if rv.Type().Kind() != reflect.Ptr || rv.IsNil() {
		return ErrorInternal
	}

	irv := rv.Elem()
	irt := irv.Type()

	if irt.Kind() != reflect.Struct {
		return nil
	}

	u := ExtractUserContext(ctx)

	for i := 0; i < irv.NumField(); i++ {

		rv := irv.Field(i)
		tf := irt.Field(i)

		if v, ok := tf.Tag.Lookup(ParamTag); ok {
			rv.Set(reflect.ValueOf(ctx.Param(v)))
		}

		if strings.EqualFold(tf.Name, MerchantIdField) {
			if tf.Type.Kind() != reflect.String {
				return ErrorInternal
			}
			rv.Set(reflect.ValueOf(u.MerchantId))
		}

		if strings.EqualFold(tf.Name, MerchantSliceField) {
			if tf.Type.Kind() != reflect.Slice {
				return ErrorInternal
			}
			if rv.Type().Elem().Kind() == reflect.String {
				rv.Set(reflect.ValueOf([]string{u.MerchantId}))
			}
		}
	}

	return nil
}

type Binder struct {
	LimitDefault, OffsetDefault, LimitMax int32
}

func (b *Binder) Bind(i interface{}, ctx echo.Context) (err error) {
	if err := EchoBinderDefault.Bind(i, ctx); err != nil {
		return err
	}
	//
	params := ctx.QueryParams()
	limit := params.Get(RequestParameterLimit)
	if len(limit) > 0 {
		if ta := typ.StringInt32(limit); ta.Err() != nil {
			return ta.Err()
		} else if ta.V() < 0 {
			ta.Set(b.LimitDefault)
		} else if ta.V() > b.LimitMax {
			ta.Set(b.LimitMax)
		}
	}

	offset := params.Get(RequestParameterOffset)
	if len(offset) > 0 {
		if ta := typ.StringInt32(offset); ta.Err() != nil {
			return ta.Err()
		}
	}
	//
	if binder := ExtractBinderContext(ctx); binder != nil {
		return binder.Bind(i, ctx)
	}
	return nil
}

type OrderFormBinder struct{}
type OrderJsonBinder struct{}
type PaymentCreateProcessBinder struct{}
type OnboardingMerchantListingBinder struct {
	LimitDefault, OffsetDefault int32
}
type OnboardingNotificationsListBinder struct {
	LimitDefault, OffsetDefault int32
}
type PaylinksListBinder struct {
	LimitDefault, OffsetDefault int32
}
type PaylinksUrlBinder struct{}
type PaylinksCreateBinder struct{}
type PaylinksUpdateBinder struct{}
type ProductsCreateProductBinder struct{}
type ProductsUpdateProductBinder struct{}

// ChangeMerchantDataRequestBinder
type ChangeMerchantDataRequestBinder struct {
	dispatch HandlerSet
	provider.LMT
	cfg Config
}

type OrderListRefundsBinder struct {
	cfg Config
}

// NewChangeMerchantDataRequestBinder
func NewChangeMerchantDataRequestBinder(set HandlerSet, cfg Config) *ChangeMerchantDataRequestBinder {
	return &ChangeMerchantDataRequestBinder{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

// ChangeProjectRequestBinder
type ChangeProjectRequestBinder struct {
	dispatch HandlerSet
	provider.LMT
	cfg Config
	Binder
}

// NewChangeProjectRequestBinder
func NewChangeProjectRequestBinder(set HandlerSet, cfg Config) *ChangeProjectRequestBinder {
	return &ChangeProjectRequestBinder{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

// Bind
func (cb *OrderFormBinder) Bind(i interface{}, ctx echo.Context) (err error) {

	if err = BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	params, err := ctx.FormParams()
	addParams := make(map[string]string)
	rawParams := make(map[string]string)

	if err != nil {
		return err
	}

	o := i.(*billing.OrderCreateRequest)

	for key, value := range params {
		if _, ok := OrderReservedWords[key]; !ok {
			addParams[key] = value[0]
		}

		rawParams[key] = value[0]
	}

	o.Other = addParams
	o.RawParams = rawParams

	return
}

// Bind
func (cb *OrderJsonBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	var buf []byte

	if ctx.Request().Body != nil {
		buf, err = ioutil.ReadAll(ctx.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))

		if err != nil {
			return err
		}

		ctx.Request().Body = rdr
	}

	if err = BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	structure := i.(*billing.OrderCreateRequest)
	structure.RawBody = string(buf)

	return
}

// Bind
func (cb *PaymentCreateProcessBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	untypedData := make(map[string]interface{})

	if err = BinderDefault.Bind(&untypedData, ctx); err != nil {
		return
	}

	data := i.(map[string]string)

	for k, v := range untypedData {
		switch sv := v.(type) {
		case bool:
			data[k] = "0"

			if sv == true {
				data[k] = "1"
			}
			break
		default:
			data[k] = fmt.Sprintf("%v", sv)
		}
	}

	return
}

// Bind
func (cb *OnboardingMerchantListingBinder) Bind(i interface{}, ctx echo.Context) (err error) {

	if err = BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	params := ctx.QueryParams()
	structure := i.(*grpc.MerchantListingRequest)

	if structure.Limit <= 0 {
		structure.Limit = cb.LimitDefault
	}

	if v, ok := params[RequestParameterIsSigned]; ok {
		if v[0] == "0" || v[0] == "false" {
			structure.IsSigned = 2
		} else {
			if v[0] == "1" || v[0] == "true" {
				structure.IsSigned = 2
			} else {
				return ErrorRequestParamsIncorrect
			}
		}
	}

	return
}

// Bind
func (cb *OnboardingNotificationsListBinder) Bind(i interface{}, ctx echo.Context) error {

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	params := ctx.QueryParams()
	structure := i.(*grpc.ListingNotificationRequest)

	if structure.Limit <= 0 {
		structure.Limit = cb.LimitDefault
	}

	// TODO: to remove
	if v, ok := params[RequestParameterIsSystem]; ok {
		if v[0] == "0" || v[0] == "false" {
			structure.IsSystem = 1
		} else {
			structure.IsSystem = 2
		}
	}

	return nil
}

// Bind
func (b *PaylinksListBinder) Bind(i interface{}, ctx echo.Context) error {

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	params := ctx.QueryParams()
	structure := i.(*paylink.GetPaylinksRequest)

	structure.Limit = uint32(b.LimitDefault)
	structure.Offset = uint32(b.OffsetDefault)

	if v, ok := params[RequestParameterLimit]; ok {
		i, err := strconv.ParseInt(v[0], 10, 32)
		if err != nil {
			return err
		}
		structure.Limit = uint32(i)
	}

	if v, ok := params[RequestParameterOffset]; ok {
		i, err := strconv.ParseInt(v[0], 10, 32)
		if err != nil {
			return err
		}
		structure.Offset = uint32(i)
	}

	structure.ProjectId = ctx.Param(RequestParameterProjectId)

	return nil
}

// Bind
func (b *PaylinksUrlBinder) Bind(i interface{}, ctx echo.Context) error {

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	params := ctx.QueryParams()
	structure := i.(*paylink.GetPaylinkURLRequest)

	id := ctx.Param(RequestParameterId)
	structure.Id = id

	if v, ok := params[RequestParameterUtmSource]; ok {
		structure.UtmSource = v[0]
	}

	if v, ok := params[RequestParameterUtmMedium]; ok {
		structure.UtmMedium = v[0]
	}

	if v, ok := params[RequestParameterUtmCampaign]; ok {
		structure.UtmCampaign = v[0]
	}

	return nil
}

// Bind
func (b *PaylinksCreateBinder) Bind(i interface{}, ctx echo.Context) error {

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	structure := i.(*paylink.CreatePaylinkRequest)
	structure.Id = ""

	return nil
}

// Bind
func (b *PaylinksUpdateBinder) Bind(i interface{}, ctx echo.Context) error {
	id := ctx.Param(RequestParameterId)
	if id == "" {
		return ErrorIncorrectPaylinkId
	}

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	structure := i.(*paylink.CreatePaylinkRequest)
	structure.Id = id

	return nil
}

// Bind
func (b *ProductsCreateProductBinder) Bind(i interface{}, ctx echo.Context) error {

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	structure := i.(*grpc.Product)
	structure.Id = ""

	return nil
}

// Bind
func (b *ProductsUpdateProductBinder) Bind(i interface{}, ctx echo.Context) error {
	id := ctx.Param(RequestParameterId)
	if id == "" || bson.IsObjectIdHex(id) == false {
		return ErrorIncorrectProductId
	}

	if err := BinderDefault.Bind(i, ctx); err != nil {
		return err
	}

	structure := i.(*grpc.Product)
	structure.Id = id

	return nil
}

// Bind
func (b *ChangeMerchantDataRequestBinder) Bind(i interface{}, ctx echo.Context) error {

	// TODO: to remove  a whole method, need make check in billing

	req := make(map[string]interface{})

	if err := BinderDefault.Bind(&req, ctx); err != nil {
		return err
	}

	merchantId := ctx.Param(RequestParameterId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return ErrorIncorrectMerchantId
	}

	mReq := &grpc.GetMerchantByRequest{MerchantId: merchantId}
	mRsp, err := b.dispatch.Services.Billing.GetMerchantBy(context.Background(), mReq)

	if err != nil {
		b.L().Error(`Call billing server method "GetMerchantBy" failed`, logger.Args("error", err.Error(), "request", mReq))
		return ErrorUnknown
	}

	if mRsp.Status != pkg.ResponseStatusOk {
		return mRsp.Message
	}

	structure := i.(*grpc.ChangeMerchantDataRequest)
	structure.MerchantId = merchantId
	structure.HasMerchantSignature = mRsp.Item.HasMerchantSignature
	structure.HasPspSignature = mRsp.Item.HasPspSignature

	if v, ok := req[RequestParameterHasMerchantSignature]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageHasMerchantSignatureIncorrectType
		} else {
			structure.HasMerchantSignature = tv
		}
	}

	if v, ok := req[RequestParameterHasPspSignature]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageHasPspSignatureIncorrectType
		} else {
			structure.HasPspSignature = tv
		}
	}

	return nil
}

// Bind
func (b *ChangeProjectRequestBinder) Bind(i interface{}, ctx echo.Context) error {
	req := make(map[string]interface{})

	if err := b.Binder.Bind(&req, ctx); err != nil {
		return err
	}

	projectId := ctx.Param(RequestParameterId)

	if projectId == "" || bson.IsObjectIdHex(projectId) == false {
		return ErrorIncorrectProjectId
	}

	pReq := &grpc.GetProjectRequest{ProjectId: projectId}
	pRsp, err := b.dispatch.Services.Billing.GetProject(context.Background(), pReq)

	if err != nil {
		b.L().Error(`Call billing server method "GetProject" failed`, logger.Args("error", err.Error(), "request", pReq))
		return ErrorUnknown
	}

	if pRsp.Status != pkg.ResponseStatusOk {
		return pRsp.Message
	}

	structure := i.(*billing.Project)
	structure.Id = projectId
	structure.MerchantId = pRsp.Item.MerchantId
	structure.Name = pRsp.Item.Name
	structure.Image = pRsp.Item.Image
	structure.CallbackCurrency = pRsp.Item.CallbackCurrency
	structure.CallbackProtocol = pRsp.Item.CallbackProtocol
	structure.CreateOrderAllowedUrls = pRsp.Item.CreateOrderAllowedUrls
	structure.AllowDynamicNotifyUrls = pRsp.Item.AllowDynamicNotifyUrls
	structure.AllowDynamicRedirectUrls = pRsp.Item.AllowDynamicRedirectUrls
	structure.LimitsCurrency = pRsp.Item.LimitsCurrency
	structure.MinPaymentAmount = pRsp.Item.MinPaymentAmount
	structure.MaxPaymentAmount = pRsp.Item.MaxPaymentAmount
	structure.NotifyEmails = pRsp.Item.NotifyEmails
	structure.IsProductsCheckout = pRsp.Item.IsProductsCheckout
	structure.SecretKey = pRsp.Item.SecretKey
	structure.SignatureRequired = pRsp.Item.SignatureRequired
	structure.SendNotifyEmail = pRsp.Item.SendNotifyEmail
	structure.UrlCheckAccount = pRsp.Item.UrlCheckAccount
	structure.UrlProcessPayment = pRsp.Item.UrlProcessPayment
	structure.UrlRedirectFail = pRsp.Item.UrlRedirectFail
	structure.UrlRedirectSuccess = pRsp.Item.UrlRedirectSuccess
	structure.Status = pRsp.Item.Status

	if v, ok := req[RequestParameterName]; ok {
		tv, ok := v.(map[string]interface{})

		if !ok || len(tv) <= 0 {
			return ErrorMessageNameIncorrectType
		}

		for k, tvv := range tv {
			structure.Name[k] = tvv.(string)
		}
	}

	if v, ok := req[RequestParameterImage]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageImageIncorrectType
		} else {
			structure.Image = tv
		}
	}

	if v, ok := req[RequestParameterCallbackCurrency]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageCallbackCurrencyIncorrectType
		} else {
			structure.CallbackCurrency = tv
		}
	}

	if v, ok := req[RequestParameterCallbackProtocol]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageCallbackProtocolIncorrectType
		} else {
			structure.CallbackProtocol = tv
		}
	}

	if v, ok := req[RequestParameterCreateOrderAllowedUrls]; ok {
		tv, ok := v.([]interface{})

		if !ok {
			return ErrorMessageCreateOrderAllowedUrlsIncorrectType
		}

		structure.CreateOrderAllowedUrls = []string{}

		for _, tvv := range tv {
			structure.CreateOrderAllowedUrls = append(structure.CreateOrderAllowedUrls, tvv.(string))
		}
	}

	if v, ok := req[RequestParameterAllowDynamicNotifyUrls]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageAllowDynamicNotifyUrlsIncorrectType
		} else {
			structure.AllowDynamicNotifyUrls = tv
		}
	}

	if v, ok := req[RequestParameterAllowDynamicRedirectUrls]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageAllowDynamicRedirectUrlsIncorrectType
		} else {
			structure.AllowDynamicRedirectUrls = tv
		}
	}

	if v, ok := req[RequestParameterLimitsCurrency]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageLimitsCurrencyIncorrectType
		} else {
			structure.LimitsCurrency = tv
		}
	}

	if v, ok := req[RequestParameterMinPaymentAmount]; ok {
		if tv, ok := v.(float64); !ok {
			return ErrorMessageMinPaymentAmountIncorrectType
		} else {
			structure.MinPaymentAmount = tv
		}
	}

	if v, ok := req[RequestParameterMaxPaymentAmount]; ok {
		if tv, ok := v.(float64); !ok {
			return ErrorMessageMaxPaymentAmountIncorrectType
		} else {
			structure.MaxPaymentAmount = tv
		}
	}

	if v, ok := req[RequestParameterNotifyEmails]; ok {
		tv, ok := v.([]interface{})

		if !ok {
			return ErrorMessageNotifyEmailsIncorrectType
		}

		structure.NotifyEmails = []string{}

		for _, tvv := range tv {
			structure.NotifyEmails = append(structure.NotifyEmails, tvv.(string))
		}
	}

	if v, ok := req[RequestParameterIsProductsCheckout]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageIsProductsCheckoutIncorrectType
		} else {
			structure.IsProductsCheckout = tv
		}
	}

	if v, ok := req[RequestParameterSecretKey]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageSecretKeyIncorrectType
		} else {
			structure.SecretKey = tv
		}
	}

	if v, ok := req[RequestParameterSignatureRequired]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageSignatureRequiredIncorrectType
		} else {
			structure.SignatureRequired = tv
		}
	}

	if v, ok := req[RequestParameterSendNotifyEmail]; ok {
		if tv, ok := v.(bool); !ok {
			return ErrorMessageSendNotifyEmailIncorrectType
		} else {
			structure.SendNotifyEmail = tv
		}
	}

	if v, ok := req[RequestParameterUrlCheckAccount]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlCheckAccountIncorrectType
		} else {
			structure.UrlCheckAccount = tv
		}
	}

	if v, ok := req[RequestParameterUrlProcessPayment]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlProcessPaymentIncorrectType
		} else {
			structure.UrlProcessPayment = tv
		}
	}

	if v, ok := req[RequestParameterUrlRedirectFail]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlRedirectFailIncorrectType
		} else {
			structure.UrlRedirectFail = tv
		}
	}

	if v, ok := req[RequestParameterUrlRedirectSuccess]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlRedirectSuccessIncorrectType
		} else {
			structure.UrlRedirectSuccess = tv
		}
	}

	if v, ok := req[RequestParameterStatus]; ok {
		if tv, ok := v.(float64); !ok {
			return ErrorMessageStatusIncorrectType
		} else {
			structure.Status = int32(tv)
		}
	}

	if v, ok := req[RequestParameterUrlChargebackPayment]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlChargebackPayment
		} else {
			structure.UrlChargebackPayment = tv
		}
	}

	if v, ok := req[RequestParameterUrlCancelPayment]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlCancelPayment
		} else {
			structure.UrlCancelPayment = tv
		}
	}

	if v, ok := req[RequestParameterUrlFraudPayment]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlFraudPayment
		} else {
			structure.UrlFraudPayment = tv
		}
	}

	if v, ok := req[RequestParameterUrlRefundPayment]; ok {
		if tv, ok := v.(string); !ok {
			return ErrorMessageUrlRefundPayment
		} else {
			structure.UrlRefundPayment = tv
		}
	}

	return nil
}

func (b *OrderListRefundsBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)

	if err != nil {
		return err
	}

	structure := i.(*grpc.ListRefundsRequest)
	structure.OrderId = ctx.Param(RequestParameterOrderId)

	if structure.Limit <= 0 {
		structure.Limit = b.cfg.LimitDefault
	}

	return nil
}
