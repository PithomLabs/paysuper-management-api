package api

import "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"

func newManagementApiResponseError(code, msg string, details ...string) *grpc.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return &grpc.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}

func newValidationError(details string) *grpc.ResponseErrorMessage {
	return newManagementApiResponseError(errorValidationFailed.Code, errorValidationFailed.Message, details)
}

var (
	errorUnknown                                      = newManagementApiResponseError("ma000001", "unknown error. try request later")
	errorValidationFailed                             = newManagementApiResponseError("ma000002", "validation failed")
	errorInternal                                     = newManagementApiResponseError("ma000003", "internal error")
	errorMessageAccessDenied                          = newManagementApiResponseError("ma000004", "access denied")
	errorIdIsEmpty                                    = newManagementApiResponseError("ma000005", "identifier can't be empty")
	errorIncorrectMerchantId                          = newManagementApiResponseError("ma000006", "incorrect merchant identifier")
	errorIncorrectNotificationId                      = newManagementApiResponseError("ma000007", "incorrect notification identifier")
	errorIncorrectOrderId                             = newManagementApiResponseError("ma000008", "incorrect order identifier")
	errorIncorrectProductId                           = newManagementApiResponseError("ma000009", "incorrect product identifier")
	errorIncorrectCountryIdentifier                   = newManagementApiResponseError("ma000010", "incorrect country identifier")
	errorIncorrectCurrencyIdentifier                  = newManagementApiResponseError("ma000011", "incorrect currency identifier")
	errorMessageOrdersNotFound                        = newManagementApiResponseError("ma000012", "orders not found")
	errorCountryNotFound                              = newManagementApiResponseError("ma000013", "country not found")
	errorCurrencyNotFound                             = newManagementApiResponseError("ma000014", "currency not found")
	errorNotificationNotFound                         = newManagementApiResponseError("ma000015", "notification not found")
	errorMerchantNotFound                             = newManagementApiResponseError("ma000016", "merchant not found")
	errorMerchantCreateFailed                         = newManagementApiResponseError("ma000017", "merchant create failed")
	errorMerchantDeleteFailed                         = newManagementApiResponseError("ma000018", "merchant delete failed")
	errorMerchantUpdateFailed                         = newManagementApiResponseError("ma000019", "merchant update failed")
	errorMessageAgreementCanNotBeGenerate             = newManagementApiResponseError("ma000020", "agreement can't be generated for not checked merchant data")
	errorMessageAgreementNotGenerated                 = newManagementApiResponseError("ma000021", "agreement for merchant not generated early")
	errorMessageSignatureHeaderIsEmpty                = newManagementApiResponseError("ma000022", "header with request signature can't be empty")
	errorRequestParamsIncorrect                       = newManagementApiResponseError("ma000023", "incorrect request parameters")
	errorEmailFieldIsRequired                         = newManagementApiResponseError("ma000024", "email field is required")
	errorRequestDataInvalid                           = newManagementApiResponseError("ma000026", "request data invalid")
	errorCountriesListError                           = newManagementApiResponseError("ma000027", "countries list error")
	errorAgreementFileNotExist                        = newManagementApiResponseError("ma000028", "file for the specified key does not exist")
	errorNotMultipartForm                             = newManagementApiResponseError("ma000029", "no multipart boundary param in Content-Type")
	errorUploadFailed                                 = newManagementApiResponseError("ma000030", "upload failed")
	errorIncorrectProjectId                           = newManagementApiResponseError("ma000031", "incorrect project identifier")
	errorIncorrectPaymentMethodId                     = newManagementApiResponseError("ma000032", "incorrect payment method identifier")
	errorIncorrectPaylinkId                           = newManagementApiResponseError("ma000033", "incorrect paylink identifier")
	errorMessageAuthorizationHeaderNotFound           = newManagementApiResponseError("ma000034", "authorization header not found")
	errorMessageAuthorizationTokenNotFound            = newManagementApiResponseError("ma000035", "authorization token not found")
	errorMessageAuthorizedUserNotFound                = newManagementApiResponseError("ma000036", "information about authorized user not found")
	errorMessageStatusIncorrectType                   = newManagementApiResponseError("ma000037", "status parameter has incorrect type")
	errorMessageAgreementNotFound                     = newManagementApiResponseError("ma000038", "agreement for merchant not found")
	errorMessageAgreementUploadMaxSize                = newManagementApiResponseError("ma000039", "agreement document max upload size exceeded")
	errorMessageAgreementContentType                  = newManagementApiResponseError("ma000040", "agreement document type must be a pdf")
	errorMessageAgreementTypeIncorrectType            = newManagementApiResponseError("ma000041", "agreement type parameter have incorrect type")
	errorMessageHasMerchantSignatureIncorrectType     = newManagementApiResponseError("ma000042", "merchant signature parameter has incorrect type")
	errorMessageHasPspSignatureIncorrectType          = newManagementApiResponseError("ma000043", "paysuper signature parameter has incorrect type")
	errorMessageAgreementSentViaMailIncorrectType     = newManagementApiResponseError("ma000044", "agreement sent via email parameter has incorrect type")
	errorMessageMailTrackingLinkIncorrectType         = newManagementApiResponseError("ma000045", "mail tracking link parameter has incorrect type")
	errorMessageNameIncorrectType                     = newManagementApiResponseError("ma000046", "name parameter has incorrect type")
	errorMessageImageIncorrectType                    = newManagementApiResponseError("ma000047", "image parameter has incorrect type")
	errorMessageCallbackCurrencyIncorrectType         = newManagementApiResponseError("ma000048", "callback currency parameter has incorrect type")
	errorMessageCallbackProtocolIncorrectType         = newManagementApiResponseError("ma000049", "callback protocol parameter has incorrect type")
	errorMessageCreateOrderAllowedUrlsIncorrectType   = newManagementApiResponseError("ma000050", "create order allowed urls parameter has incorrect type")
	errorMessageAllowDynamicNotifyUrlsIncorrectType   = newManagementApiResponseError("ma000051", "allow dynamic notify urls parameter has incorrect type")
	errorMessageAllowDynamicRedirectUrlsIncorrectType = newManagementApiResponseError("ma000052", "allow dynamic redirect urls parameter has incorrect type")
	errorMessageLimitsCurrencyIncorrectType           = newManagementApiResponseError("ma000053", "limits currency parameter has incorrect type")
	errorMessageMinPaymentAmountIncorrectType         = newManagementApiResponseError("ma000054", "min payment amount parameter has incorrect type")
	errorMessageMaxPaymentAmountIncorrectType         = newManagementApiResponseError("ma000055", "max payment amount parameter has incorrect type")
	errorMessageNotifyEmailsIncorrectType             = newManagementApiResponseError("ma000056", "notify emails parameter has incorrect type")
	errorMessageIsProductsCheckoutIncorrectType       = newManagementApiResponseError("ma000057", "is products checkout parameter has incorrect type")
	errorMessageSecretKeyIncorrectType                = newManagementApiResponseError("ma000058", "secret key parameter has incorrect type")
	errorMessageSignatureRequiredIncorrectType        = newManagementApiResponseError("ma000059", "signature required parameter has incorrect type")
	errorMessageSendNotifyEmailIncorrectType          = newManagementApiResponseError("ma000060", "send notify email parameter has incorrect type")
	errorMessageUrlCheckAccountIncorrectType          = newManagementApiResponseError("ma000061", "url check account parameter has incorrect type")
	errorMessageUrlProcessPaymentIncorrectType        = newManagementApiResponseError("ma000062", "url process payment parameter has incorrect type")
	errorMessageUrlRedirectFailIncorrectType          = newManagementApiResponseError("ma000063", "url redirect fail parameter has incorrect type")
	errorMessageUrlRedirectSuccessIncorrectType       = newManagementApiResponseError("ma000064", "url redirect success parameter has incorrect type")
	errorMessageUrlChargebackPayment                  = newManagementApiResponseError("ma000065", "url chargeback payment parameter has incorrect type")
	errorMessageUrlCancelPayment                      = newManagementApiResponseError("ma000066", "url cancel payment parameter has incorrect type")
	errorMessageUrlFraudPayment                       = newManagementApiResponseError("ma000067", "url fraud payment parameter has incorrect type")
	errorMessageUrlRefundPayment                      = newManagementApiResponseError("ma000068", "url refund payment parameter has incorrect type")
)