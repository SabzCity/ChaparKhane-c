/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"crypto/sha512"
	"strconv"

	"../libgo/achaemenid"
	"../libgo/asanak.com"
	"../libgo/authorization"
	"../libgo/convert"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/log"
	"../libgo/otp"
	"../libgo/srpc"
	"../libgo/syllab"
)

// https://tools.ietf.org/html/rfc6238
var sendOtpService = achaemenid.Service{
	ID:                633216246,
	URI:               "", // API services can set like "/apis?633216246" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDRead,
	IssueDate:         1592374531,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "SendOtp",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: `Request to get approve code for given phone or email.
It can use for many purpose e.g. to recover person, improve account security by force use OTP in some very dangerous operation`,
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: SendOtpSRPC,
	HTTPHandler: SendOtpHTTP,
}

// SendOtpSRPC is sRPC handler of SendOtp service.
func SendOtpSRPC(st *achaemenid.Stream) {
	var req = &sendOtpReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *sendOtpRes
	res, st.Err = sendOtp(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// SendOtpHTTP is HTTP handler of SendOtp service.
func SendOtpHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &sendOtpReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *sendOtpRes
	res, st.Err = sendOtp(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

const (
	smsOTPDigits = 8
	smsOTPPeriod = 30 * 60 // 30 minutes

	englishSMSTemplate = "SabzCity\n\nYour OTP\n"
	persianSMSTemplate = "شهرسبز\nرمز یکبار مصرف\n\n"
)

type sendOtpReq struct {
	CaptchaID   [16]byte `json:",string"`
	Email       string   `valid:"Email"`
	PhoneNumber uint64   `valid:"PhoneNumber"`
	PhoneType   uint8    // 0:SMS 1:call
	Language    lang.Language
}

type sendOtpRes struct {
	OTPID uint64
}

func sendOtp(st *achaemenid.Stream, req *sendOtpReq) (res *sendOtpRes, err *er.Error) {
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	// TODO::: Prevent DDos attack by do some easy process for user e.g. captcha is not good way!
	err = phraseCaptchas.Check(req.CaptchaID)
	if err != nil {
		return
	}

	res = &sendOtpRes{}

	if req.Email != "" {
		var otpReq = otp.GenerateTimeOTPReq{
			Hasher:     sha512.New512_256(),
			SecretKey:  smsOTPSecurityKey,
			Additional: convert.UnsafeStringToByteSlice(req.Email),
			Period:     smsOTPPeriod,
			Digits:     smsOTPDigits,
		}
		var timeOTP uint64
		timeOTP, err = otp.GenerateTimeOTP(&otpReq)
		if err != nil {
			return
		}

		if log.DebugMode {
			log.Debug("Desire Email OTP:", timeOTP)
		} else {
			// TODO::: send OTP to desire email address
		}
	}

	if req.PhoneNumber > 0 {
		var otpReq = otp.GenerateTimeOTPReq{
			Hasher:     sha512.New512_256(),
			SecretKey:  smsOTPSecurityKey,
			Additional: make([]byte, 8),
			Period:     smsOTPPeriod,
			Digits:     smsOTPDigits,
		}
		syllab.SetUInt64(otpReq.Additional, 0, req.PhoneNumber)
		var timeOTP uint64
		timeOTP, err = otp.GenerateTimeOTP(&otpReq)
		if err != nil {
			return
		}

		if log.DevMode {
			log.Debug("Desire Phone OTP:", timeOTP)
		} else {
			if req.PhoneType == 0 {
				var SendSMSReq = asanak.SendSMSReq{
					Destination: []string{strconv.FormatUint(req.PhoneNumber, 10)},
				}
				switch req.Language {
				case lang.PersianLanguage:
					SendSMSReq.Message = make([]byte, 0, len(persianSMSTemplate)+smsOTPDigits+len(server.Manifest.DomainName)+smsOTPDigits+5)
					SendSMSReq.Message = append(SendSMSReq.Message, persianSMSTemplate...)
				default:
					SendSMSReq.Message = make([]byte, 0, len(englishSMSTemplate)+smsOTPDigits+len(server.Manifest.DomainName)+smsOTPDigits+5)
					SendSMSReq.Message = append(SendSMSReq.Message, englishSMSTemplate...)
				}
				SendSMSReq.Message = strconv.AppendUint(SendSMSReq.Message, timeOTP, 10)
				// Add web-otp
				SendSMSReq.Message = append(SendSMSReq.Message, "\n\n@"+server.Manifest.DomainName+" #"...)
				SendSMSReq.Message = strconv.AppendUint(SendSMSReq.Message, timeOTP, 10)

				var SendSMSRes asanak.SendSMSRes
				SendSMSRes, err = smsProvider.SendSMS(&SendSMSReq)
				if err != nil {
					return
				}
				res.OTPID = SendSMSRes[0]
			}
		}
	}

	return
}

func (req *sendOtpReq) validator() (err *er.Error) {
	return
}

func (req *sendOtpReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.CaptchaID[:], buf[0:])
	req.Email = syllab.UnsafeGetString(buf, 16)
	req.PhoneNumber = syllab.GetUInt64(buf, 24)
	req.PhoneType = syllab.GetUInt8(buf, 32)
	req.Language = lang.Language(syllab.GetUInt8(buf, 33))
	return
}

func (req *sendOtpReq) syllabStackLen() (ln uint32) {
	return 37
}

func (req *sendOtpReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'C':
			decoder.SetFounded()
			decoder.Offset(11)
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
			if err != nil {
				return
			}
		case 'E':
			decoder.SetFounded()
			decoder.Offset(7)
			decoder.Offset(1)
			req.Email = decoder.DecodeString()
		case 'P':
			switch decoder.Buf[5] {
			case 'N':
				decoder.SetFounded()
				decoder.Offset(13)
				req.PhoneNumber, err = decoder.DecodeUInt64()
				if err != nil {
					return
				}
			case 'T':
				decoder.SetFounded()
				decoder.Offset(11)
				var num uint64
				num, err = decoder.DecodeUInt64()
				if err != nil {
					return
				}
				req.PhoneType = uint8(num)
			}
		case 'L':
			decoder.SetFounded()
			decoder.Offset(10)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Language = lang.Language(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *sendOtpRes) syllabEncoder(buf []byte) {
	syllab.SetUInt64(buf, 0, res.OTPID)
}

func (res *sendOtpRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *sendOtpRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *sendOtpRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *sendOtpRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"OTPID":`)
	encoder.EncodeUInt64(uint64(res.OTPID))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *sendOtpRes) jsonLen() (ln int) {
	ln += 20
	ln += 10
	return
}
