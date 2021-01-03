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
	IssueDate:         1603174921,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "AuthenticateAppConnection",
		lang.LanguagePersian: "تایید هوییت ارتباطی برنامه",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: `Authenticate the person active app connection, can't use to authenticate connection for delegate purpose!
Usually use in HTTP protocol!`,
		lang.LanguagePersian: "تایید هوییت ارتباطی برنامه",
	},
	TAGS: []string{
		"UserAppConnection", "Authentication",
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
			Name:     achaemenid.HTTPCookieNameBaseConnID,
			Value:    base64.RawStdEncoding.EncodeToString(st.Connection.ID[:]),
			MaxAge:   "630720000", // = 20 year = 20*365*24*60*60
			Secure:   true,
			HTTPOnly: true,
			SameSite: "Lax",
		}, http.SetCookie{
			Name:     achaemenid.HTTPCookieNameBaseUserID,
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
	ThingID      [32]byte `json:",string"`
	CaptchaID    [16]byte `json:",string"`
	PersonID     [32]byte `json:",string"`
	PasswordHash [32]byte `json:",string"`
	OTP          uint32
}

func authenticateAppConnection(st *achaemenid.Stream, req *authenticateAppConnectionReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
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
		if err.Equal(ganjine.ErrRecordNotFound) {
			return
		}
		err = ErrBadSituation
		return
	}

	if req.PasswordHash != pa.PasswordHash {
		err = ErrBadPasswordOrOTP
		return
	}
	if pa.Status == datastore.PersonAuthenticationForceUse2Factor {
		// TODO::: check OTP
		// err = ErrBadPasswordOrOTP
	}

	if req.ThingID != [32]byte{} {
		var conn = achaemenid.Server.Connections.GetConnByUserIDThingID(pa.PersonID, req.ThingID)
		if conn != nil {
			st.Connection = conn
			achaemenid.Server.Connections.RegisterConnection(conn)
			return
		}
	}

	st.Connection.UserID = pa.PersonID
	st.Connection.UserType = authorization.UserTypePerson
	st.Connection.State = achaemenid.StateNew

	achaemenid.Server.Connections.SaveConn(st.Connection)
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

	copy(req.ThingID[:], buf[0:])
	copy(req.CaptchaID[:], buf[32:])
	copy(req.PersonID[:], buf[48:])
	copy(req.PasswordHash[:], buf[80:])
	req.OTP = syllab.GetUInt32(buf, 112)
	return
}

func (req *authenticateAppConnectionReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ThingID[:])
	copy(buf[32:], req.CaptchaID[:])
	copy(buf[48:], req.PersonID[:])
	copy(buf[80:], req.PasswordHash[:])
	syllab.SetUInt32(buf, 112, req.OTP)
	return
}

func (req *authenticateAppConnectionReq) syllabStackLen() (ln uint32) {
	return 116
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ThingID":
			err = decoder.DecodeByteArrayAsBase64(req.ThingID[:])
		case "CaptchaID":
			err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
		case "PersonID":
			err = decoder.DecodeByteArrayAsBase64(req.PersonID[:])
		case "PasswordHash":
			err = decoder.DecodeByteArrayAsBase64(req.PasswordHash[:])
		case "OTP":
			var num uint64
			num, err = decoder.DecodeUInt64()
			req.OTP = uint32(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *authenticateAppConnectionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ThingID":"`)
	encoder.EncodeByteSliceAsBase64(req.ThingID[:])

	encoder.EncodeString(`","CaptchaID":"`)
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
	ln = 239
	return
}
