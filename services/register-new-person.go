/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"crypto/rand"
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

var registerNewPersonService = achaemenid.Service{
	ID:                956555232,
	URI:               "", // API services can set like "/apis?956555232" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDCreate,
	IssueDate:         1592316187,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "RegisterNewPerson",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "register a new user in SabzCity platform.",
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: RegisterNewPersonSRPC,
	HTTPHandler: RegisterNewPersonHTTP,
}

// RegisterNewPersonSRPC is sRPC handler of RegisterNewPerson service.
func RegisterNewPersonSRPC(st *achaemenid.Stream) {
	var req = &registerNewPersonReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerNewPersonRes
	res, st.Err = registerNewPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterNewPersonHTTP is HTTP handler of RegisterNewPerson service.
func RegisterNewPersonHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerNewPersonReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerNewPersonRes
	res, st.Err = registerNewPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerNewPersonReq struct {
	PhoneNumber uint64
	PhoneOTP    uint64
	CaptchaID   [16]byte `json:",string"`

	PasswordHash  [32]byte `json:",string"`
	OTPAdditional int32
}

type registerNewPersonRes struct {
	PersonID    [32]byte `json:",string"` // UUID of registered user
	OTPPattern  [32]byte `json:",string"`
	SecurityKey [32]byte `json:",string"`
}

func registerNewPerson(st *achaemenid.Stream, req *registerNewPersonReq) (res *registerNewPersonRes, err *er.Error) {
	var goErr error

	// Prevent DDos attack by do some easy process for user e.g. captcha is not good way!
	err = phraseCaptchas.Check(req.CaptchaID)
	if err != nil {
		return
	}

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
	if req.PhoneOTP != timeOTP {
		return nil, otp.ErrOTPWrongNumber
	}

	var OTPPattern = make([]byte, 32)
	_, goErr = rand.Read(OTPPattern)
	// Note that err == nil only if we read len(OTPPattern) bytes.
	if goErr != nil {
		err = ErrPlatformBadSituation
		return
	}

	var SecurityKey = make([]byte, 32)
	_, goErr = rand.Read(SecurityKey)
	// Note that err == nil only if we read len(SecurityKey) bytes.
	if goErr != nil {
		err = ErrPlatformBadSituation
		return
	}

	var pa = datastore.PersonAuthentication{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		PersonID:         uuid.Random32Byte(),
		ReferentPersonID: st.Connection.UserID,
		Status:           datastore.PersonAuthenticationNotForceUse2Factor,
		PasswordHash:     req.PasswordHash,
		OTPAdditional:    req.OTPAdditional,
	}
	copy(pa.OTPPattern[:], OTPPattern[:])
	copy(pa.SecurityKey[:], SecurityKey[:])
	if req.OTPAdditional != 0 {
		pa.Status = datastore.PersonAuthenticationForceUse2Factor
	}

	var registerPersonNumberReq = registerPersonNumberReq{
		PersonID:    pa.PersonID,
		PhoneNumber: req.PhoneNumber,
	}
	err = registerPersonNumber(st, &registerPersonNumberReq, true)
	if err != nil {
		return
	}

	err = pa.Set()
	if err != nil {
		// TODO::: can't easily return due to person number registered successfully!
		return
	}

	pa.IndexPersonID()
	pa.IndexPersonIDforRegisterTime()
	if st.Connection.UserType != achaemenid.UserTypeGuest {
		pa.IndexPersonIDforReferentPersonID()
	} else {
		st.Connection.UserID = pa.PersonID
		st.Connection.UserType = achaemenid.UserTypePerson
	}

	res = &registerNewPersonRes{
		PersonID:    pa.PersonID,
		OTPPattern:  pa.OTPPattern,
		SecurityKey: pa.SecurityKey,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerNewPersonReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.PhoneNumber = syllab.GetUInt64(buf, 0)
	req.PhoneOTP = syllab.GetUInt64(buf, 8)
	copy(req.CaptchaID[:], buf[16:])
	copy(req.PasswordHash[:], buf[32:])
	req.OTPAdditional = syllab.GetInt32(buf, 64)
	return
}

func (req *registerNewPersonReq) syllabStackLen() (ln uint32) {
	return 68
}

func (req *registerNewPersonReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'C':
			decoder.SetFounded()
			decoder.Offset(12)
			err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
			if err != nil {
				return
			}
		case 'P':
			switch decoder.Buf[5] {
			case 'N':
				decoder.SetFounded()
				decoder.Offset(13)
				req.PhoneNumber, err = decoder.DecodeUInt64()
				if err != nil {
					return
				}
			case 'O':
				decoder.SetFounded()
				decoder.Offset(10)
				req.PhoneOTP, err = decoder.DecodeUInt64()
				if err != nil {
					return
				}
			case 'o':
				decoder.SetFounded()
				decoder.Offset(15)
				err = decoder.DecodeByteArrayAsBase64(req.PasswordHash[:])
				if err != nil {
					return
				}
			}
		case 'O':
			decoder.SetFounded()
			decoder.Offset(15)
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			req.OTPAdditional = int32(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerNewPersonRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.PersonID[:])
	copy(buf[16:], res.OTPPattern[:])
	copy(buf[48:], res.SecurityKey[:])
	return
}

func (res *registerNewPersonRes) syllabStackLen() (ln uint32) {
	return 80
}

func (res *registerNewPersonRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerNewPersonRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerNewPersonRes) jsonEncoder() (buf []byte) {
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

func (res *registerNewPersonRes) jsonLen() (ln int) {
	ln = 177
	return
}
