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
	URI:               "", // API services can set like "/apis?756778065" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDCreate,
	IssueDate:         1602687948,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "RegisterPersonNumber",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
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
			err = authorization.ErrAuthorizationUserNotAllow
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
		if err == ganjine.ErrGanjineRecordNotFound {
			err = nil
		} else {
			err = ErrPlatformBadSituation
			return
		}
	}

	if pn.PersonID != [32]byte{} || pn.Status == datastore.PersonNumberRegister {
		err = ErrPlatformPersonNumberRegistered
		return
	} else if pn.Status == datastore.PersonNumberBlockedByJustice {
		err = ErrPlatformBlockedByJustice
		return
	}

	// TODO::: un-register last person number

	pn = datastore.PersonNumber{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		PersonID:         req.PersonID,
		Number:           req.PhoneNumber,
		Status:           datastore.PersonNumberRegister,
	}
	err = pn.Set()
	if err != nil {
		// TODO:::
		return
	}

	pn.IndexPersonID()
	pn.IndexNumber()
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
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[5] {
		case 'n':
			decoder.SetFounded()
			decoder.Offset(11)
			err = decoder.DecodeByteArrayAsBase64(req.PersonID[:])
			if err != nil {
				return
			}
		case 'N':
			decoder.SetFounded()
			decoder.Offset(13)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.PhoneNumber = uint64(num)
		case 'O':
			decoder.SetFounded()
			decoder.Offset(10)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.PhoneOTP = uint64(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}
