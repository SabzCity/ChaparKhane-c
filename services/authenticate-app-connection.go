/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"encoding/base64"

	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/ganjine"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/log"
	"../libgo/srpc"
	"../libgo/syllab"
)

var authenticateAppConnectionService = achaemenid.Service{
	ID:                528205152,
	URI:               "", // API services can set like "/apis?528205152" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDUpdate,
	IssueDate:         1603174921,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "AuthenticateAppConnection",
		lang.PersianLanguage: "تایید هوییت ارتباطی برنامه",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: `Authenticate the person active app connection, can't use to authenticate connection for delegate purpose!
Usually use in HTTP protocol!`,
		lang.PersianLanguage: "تایید هوییت ارتباطی برنامه",
	},
	TAGS: []string{
		"UserAppsConnection", "Authentication",
	},

	SRPCHandler: AuthenticateAppConnectionSRPC,
	HTTPHandler: AuthenticateAppConnectionHTTP,
}

// AuthenticateAppConnectionSRPC is sRPC handler of AuthenticateAppConnection service.
func AuthenticateAppConnectionSRPC(st *achaemenid.Stream) {
	var req = &authenticateAppConnectionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = authenticateAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// AuthenticateAppConnectionHTTP is HTTP handler of AuthenticateAppConnection service.
func AuthenticateAppConnectionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &authenticateAppConnectionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = authenticateAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	var cookies = []http.SetCookie{
		http.SetCookie{
			Name:     achaemenid.HTTPCookieNameConnectionID,
			Value:    base64.RawStdEncoding.EncodeToString(st.Connection.ID[:]),
			MaxAge:   "630720000", // = 20 year = 20*365*24*60*60
			Secure:   true,
			HTTPOnly: true,
			SameSite: "Lax",
		}, http.SetCookie{
			Name:     achaemenid.HTTPCookieNameUserID,
			Value:    base64.RawStdEncoding.EncodeToString(st.Connection.UserID[:]),
			MaxAge:   "630720000", // = 20 year = 20*365*24*60*60
			Secure:   true,
			HTTPOnly: false,
			SameSite: "Lax",
		},
	}
	if log.DevMode {
		cookies[0].Secure = false
		cookies[1].Secure = false
	}
	httpRes.Header.SetSetCookies(cookies)
}

type authenticateAppConnectionReq struct {
	CaptchaID    [16]byte `json:",string"`
	PersonID     [32]byte `json:",string"`
	PasswordHash [32]byte `json:",string"`
	OTP          uint32
}

func authenticateAppConnection(st *achaemenid.Stream, req *authenticateAppConnectionReq) (err *er.Error) {
	if st.Connection.UserType != achaemenid.UserTypeGuest {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	// Prevent DDos attack by do some easy process for user e.g. captcha is not good way!
	err = phraseCaptchas.Check(req.CaptchaID)
	if err != nil {
		return
	}

	var pa = datastore.PersonAuthentication{
		PersonID: req.PersonID,
	}
	err = pa.GetLastByPersonID()
	if err != nil {
		if err == ganjine.ErrGanjineRecordNotFound {
			return
		}
		err = ErrPlatformBadSituation
		return
	}

	if req.PasswordHash != pa.PasswordHash {
		err = ErrPlatformBadPasswordOrOTP
		return
	}
	if pa.Status == datastore.PersonAuthenticationForceUse2Factor {
		// TODO::: check OTP
		// err = ErrPlatformBadPasswordOrOTP
	}

	st.Connection.UserID = pa.PersonID
	st.Connection.UserType = achaemenid.UserTypePerson
	st.Connection.State = achaemenid.StateNew

	server.Connections.SaveConn(st.Connection)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *authenticateAppConnectionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.CaptchaID[:], buf[0:])
	copy(req.PersonID[:], buf[16:])
	copy(req.PasswordHash[:], buf[32:])
	req.OTP = syllab.GetUInt32(buf, 64)
	return
}

func (req *authenticateAppConnectionReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.CaptchaID[:])
	copy(buf[16:], req.PersonID[:])
	copy(buf[32:], req.PasswordHash[:])
	syllab.SetUInt32(buf, 64, req.OTP)
	return
}

func (req *authenticateAppConnectionReq) syllabStackLen() (ln uint32) {
	return 68
}

func (req *authenticateAppConnectionReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *authenticateAppConnectionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *authenticateAppConnectionReq) jsonDecoder(buf []byte) (err *er.Error) {
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
			switch decoder.Buf[1] {
			case 'e':
				decoder.SetFounded()
				decoder.Offset(11)
				err = decoder.DecodeByteArrayAsBase64(req.PersonID[:])
				if err != nil {
					return
				}
			case 'a':
				decoder.SetFounded()
				decoder.Offset(15)
				err = decoder.DecodeByteArrayAsBase64(req.PasswordHash[:])
				if err != nil {
					return
				}
			}
		case 'O':
			decoder.SetFounded()
			decoder.Offset(5)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.OTP = uint32(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *authenticateAppConnectionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"CaptchaID":"`)
	encoder.EncodeByteSliceAsBase64(req.CaptchaID[:])

	encoder.EncodeString(`","PersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.PersonID[:])

	encoder.EncodeString(`","PasswordHash":"`)
	encoder.EncodeByteSliceAsBase64(req.PasswordHash[:])

	encoder.EncodeString(`","OTP":`)
	encoder.EncodeUInt64(uint64(req.OTP))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *authenticateAppConnectionReq) jsonLen() (ln int) {
	ln = 183
	return
}
