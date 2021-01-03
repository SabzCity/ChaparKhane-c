/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"crypto/sha256"
	"crypto/sha512"

	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/otp"
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/uuid"
)

var registerPersonService = achaemenid.Service{
	ID:                304427766,
	IssueDate:         1592316187,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Person",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "register a new real person user in SabzCity platform.",
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: RegisterPersonSRPC,
	HTTPHandler: RegisterPersonHTTP,
}

// RegisterPersonSRPC is sRPC handler of RegisterPerson service.
func RegisterPersonSRPC(st *achaemenid.Stream) {
	var req = &registerPersonReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerPersonRes
	res, st.Err = registerPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterPersonHTTP is HTTP handler of RegisterPerson service.
func RegisterPersonHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerPersonReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerPersonRes
	res, st.Err = registerPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerPersonReq struct {
	ThingID [32]byte `json:",string"`

	PhoneNumber uint64
	PhoneOTP    uint64
	CaptchaID   [16]byte `json:",string"`

	PasswordHash  [32]byte `json:",string"`
	OTPAdditional int32
}

type registerPersonRes struct {
	PersonID    [32]byte `json:",string"` // UUID of registered user
	OTPPattern  [32]byte `json:",string"`
	SecurityKey [32]byte `json:",string"`
}

func registerPerson(st *achaemenid.Stream, req *registerPersonReq) (res *registerPersonRes, err *er.Error) {
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

	if st.Connection.UserType == authorization.UserTypeGuest && req.PasswordHash != [32]byte{} {
		// Prevent DDos attack by do some easy process for user e.g. captcha is not good way!
		err = phraseCaptchas.Check(req.CaptchaID)
		if err != nil {
			return
		}

		if req.PhoneOTP != timeOTP {
			return nil, otp.ErrOTPWrongNumber
		}
	}

	var pa = datastore.PersonAuthentication{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		PersonID:         uuid.Random32Byte(),
		ReferentPersonID: st.Connection.UserID,
		Status:           datastore.PersonAuthenticationNotForceUse2Factor,
		PasswordHash:     req.PasswordHash,
		OTPPattern:       uuid.Random32Byte(),
		OTPAdditional:    req.OTPAdditional,
		SecurityKey:      uuid.Random32Byte(),
	}
	if req.OTPAdditional != 0 {
		pa.Status = datastore.PersonAuthenticationForceUse2Factor
	}

	if req.PasswordHash == [32]byte{} {
		syllab.SetUInt64(pa.PasswordHash[:], 0, timeOTP)
		pa.PasswordHash = sha256.Sum256(pa.PasswordHash[:])
		// send otp as password to user
		var sendOtpReq = sendOtpReq{
			PhoneNumber: req.PhoneNumber,
			Language:    lang.LanguagePersian,
		}
		sendOtp(st, &sendOtpReq)
		// can't do anything here if otp not deliver to user! user must recover account later!
	}

	var registerPersonNumberReq = registerPersonNumberReq{
		PersonID:    pa.PersonID,
		PhoneNumber: req.PhoneNumber,
	}
	err = registerPersonNumber(st, &registerPersonNumberReq, true)
	if err != nil {
		return
	}

	err = pa.SaveNew()
	if err != nil {
		// TODO::: can't easily return due to person number registered successfully!
		return
	}
	if st.Connection.UserType == authorization.UserTypeGuest {
		st.Connection.UserID = pa.PersonID
		st.Connection.UserType = authorization.UserTypePerson
		st.Connection.SetThingID(req.ThingID)
	}

	res = &registerPersonRes{
		PersonID:    pa.PersonID,
		OTPPattern:  pa.OTPPattern,
		SecurityKey: pa.SecurityKey,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerPersonReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ThingID[:], buf[0:])
	req.PhoneNumber = syllab.GetUInt64(buf, 32)
	req.PhoneOTP = syllab.GetUInt64(buf, 40)
	copy(req.CaptchaID[:], buf[48:])
	copy(req.PasswordHash[:], buf[64:])
	req.OTPAdditional = syllab.GetInt32(buf, 96)
	return
}

func (req *registerPersonReq) syllabStackLen() (ln uint32) {
	return 100
}

func (req *registerPersonReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ThingID":
			err = decoder.DecodeByteArrayAsBase64(req.ThingID[:])
		case "PhoneNumber":
			req.PhoneNumber, err = decoder.DecodeUInt64()
		case "PhoneOTP":
			req.PhoneOTP, err = decoder.DecodeUInt64()
		case "CaptchaID":
			err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
		case "PasswordHash":
			err = decoder.DecodeByteArrayAsBase64(req.PasswordHash[:])
		case "OTPAdditional":
			req.OTPAdditional, err = decoder.DecodeInt32()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerPersonRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.PersonID[:])
	copy(buf[32:], res.OTPPattern[:])
	copy(buf[64:], res.SecurityKey[:])
	return
}

func (res *registerPersonRes) syllabStackLen() (ln uint32) {
	return 96
}

func (res *registerPersonRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerPersonRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerPersonRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"PersonID":"`)
	encoder.EncodeByteSliceAsBase64(res.PersonID[:])

	encoder.EncodeString(`","OTPPattern":"`)
	encoder.EncodeByteSliceAsBase64(res.OTPPattern[:])

	encoder.EncodeString(`","SecurityKey":"`)
	encoder.EncodeByteSliceAsBase64(res.SecurityKey[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerPersonRes) jsonLen() (ln int) {
	ln = 177
	return
}
