/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"strconv"

	"../libgo/achaemenid"
	"../libgo/asanak.com"
	"../libgo/captcha"
	"../libgo/errors"
	"../libgo/http"
	"../libgo/json"
)

var sendOtpService = achaemenid.Service{
	ID:                633216246,
	URI:               "", // API services can set like "/apis?633216246" but it is not efficient, find services by ID.
	Name:              "SendOtp",
	IssueDate:         1592374531,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		`Request to get approve code for given phone or email.
		It can use for many purpose e.g. to recover person, improve account security by force use OTP in some very dangerous operation`,
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: SendOtpSRPC,
	HTTPHandler: SendOtpHTTP,
}

// services errors
var (
	ErrCaptchaNotExist = errors.New("CaptchaNotExist", "Given captcha not found or expired")
	ErrCaptchaNotSolved = errors.New("CaptchaNotSolved", "Given captcha not solved before call this service")
	ErrSMSProviderError = errors.New("SMSProviderError", "Our SMS provider API can't proccess send OTP message")
)

// SendOtpSRPC is sRPC handler of SendOtp service.
func SendOtpSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &sendOtpReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *sendOtpRes
	res, st.ReqRes.Err = sendOtp(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// SendOtpHTTP is HTTP handler of SendOtp service.
func SendOtpHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &sendOtpReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *sendOtpRes
	res, st.ReqRes.Err = sendOtp(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.Body, st.ReqRes.Err = res.jsonEncoder()
	// st.ReqRes.Err make occur on just memory full!

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.SetValue(http.HeaderKeyContentType, "application/json")
}

type sendOtpReq struct {
	UserID      [16]byte
	Email       string `valid:"Email,optional"`
	PhoneNumber uint64 `valid:"UserNumber,optional"`
	PhoneType   uint8  // 0:SMS 1:call
	CaptchaID   [16]byte
}

type sendOtpRes struct {
	OTPID uint64
}

func sendOtp(st *achaemenid.Stream, req *sendOtpReq) (res *sendOtpRes, err error) {
	// TODO::: Authenticate request first by service policy.

	err = st.Authorize()
	if err != nil {
		return
	}

	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	var pc = phraseCaptchas.Get(req.CaptchaID)
	if pc == nil {
		return nil, ErrCaptchaNotExist
	}
	if pc.State != captcha.StateSolved {
		return nil, ErrCaptchaNotSolved
	}

	res = &sendOtpRes{}

	if req.Email != "" {

	}

	if req.PhoneNumber > 0 && req.PhoneType == 0 {
		var SendSMSReq = asanak.SendSMSReq{
			Destination: []string{strconv.FormatUint(req.PhoneNumber, 10)},
			Message:     smsTemplate + "",
		}
		var SendSMSRes asanak.SendSMSRes
		SendSMSRes, err = smsProvider.SendSMS(&SendSMSReq)
		if err != nil {
			return nil, ErrSMSProviderError
		}
		res.OTPID = SendSMSRes[0]
	}

	return
}

func (req *sendOtpReq) validator() (err error) {
	return
}

func (req *sendOtpReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *sendOtpReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *sendOtpRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *sendOtpRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}

const smsTemplate = "رمز یکبار مصرف شما \n" + "Your OTP \n"
