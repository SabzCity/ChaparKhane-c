/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"crypto/sha512"

	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/ganjine"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/otp"
	"../libgo/srpc"
	"../libgo/syllab"
)

var registerPersonNumberService = achaemenid.Service{
	ID:                756778065,
	IssueDate:         1602687948,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypePerson,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "RegisterPersonNumber",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"PersonNumber",
	},

	SRPCHandler: RegisterPersonNumberSRPC,
	HTTPHandler: RegisterPersonNumberHTTP,
}

// RegisterPersonNumberSRPC is sRPC handler of RegisterPersonNumber service.
func RegisterPersonNumberSRPC(st *achaemenid.Stream) {
	var req = &registerPersonNumberReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = registerPersonNumber(st, req, false)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// RegisterPersonNumberHTTP is HTTP handler of RegisterPersonNumber service.
func RegisterPersonNumberHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerPersonNumberReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = registerPersonNumber(st, req, false)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
}

type registerPersonNumberReq struct {
	PersonID    [32]byte `json:",string"` // UUID of registered user
	PhoneNumber uint64
	PhoneOTP    uint64
}

// unsafe must use internally by other services like register new person not network requests!
func registerPersonNumber(st *achaemenid.Stream, req *registerPersonNumberReq, unsafe bool) (err *er.Error) {
	if !unsafe {
		if st.Connection.UserID != req.PersonID {
			err = authorization.ErrUserNotAllow
			return
		}

		err = st.Authorize()
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
			return otp.ErrOTPWrongNumber
		}
	}

	var pn = datastore.PersonNumber{
		Number: req.PhoneNumber,
	}
	err = pn.GetLastByNumber()
	if err != nil {
		if err.Equal(ganjine.ErrRecordNotFound) {
			err = nil
		} else {
			err = ErrBadSituation
			return
		}
	}

	if pn.PersonID != [32]byte{} || pn.Status == datastore.PersonNumberRegister {
		err = ErrPersonNumberRegistered
		return
	} else if pn.Status == datastore.PersonNumberBlockedByJustice {
		err = ErrBlockedByJustice
		return
	}

	// TODO::: un-register last person number

	pn = datastore.PersonNumber{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		PersonID:         req.PersonID,
		Number:           req.PhoneNumber,
		Status:           datastore.PersonNumberRegister,
	}
	err = pn.SaveNew()
	return
}

func (req *registerPersonNumberReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.PersonID[:], buf[0:])
	req.PhoneNumber = syllab.GetUInt64(buf, 16)
	req.PhoneOTP = syllab.GetUInt64(buf, 24)
	return
}

func (req *registerPersonNumberReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *registerPersonNumberReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "PersonID":
			err = decoder.DecodeByteArrayAsBase64(req.PersonID[:])
		case "PhoneNumber":
			req.PhoneNumber, err = decoder.DecodeUInt64()
		case "PhoneOTP":
			req.PhoneOTP, err = decoder.DecodeUInt64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *registerPersonNumberReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"PersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.PersonID[:])

	encoder.EncodeString(`","PhoneNumber":`)
	encoder.EncodeUInt64(req.PhoneNumber)

	encoder.EncodeString(`,"PhoneOTP":`)
	encoder.EncodeUInt64(req.PhoneOTP)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerPersonNumberReq) jsonLen() (ln int) {
	ln = 125
	return
}
