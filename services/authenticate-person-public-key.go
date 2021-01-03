/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
)

var authenticatePersonPublicKeyService = achaemenid.Service{
	ID:                846072085,
	IssueDate:         1592379653,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "AuthenticatePersonPublicKey",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Authenticate user and store given PublicKey to use by user for any purpose!",
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: AuthenticatePersonPublicKeySRPC,
	HTTPHandler: AuthenticatePersonPublicKeyHTTP,
}

// AuthenticatePersonPublicKeySRPC is sRPC handler of AuthenticatePersonPublicKey service.
func AuthenticatePersonPublicKeySRPC(st *achaemenid.Stream) {
	var req = &authenticatePersonPublicKeyReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *authenticatePersonPublicKeyRes
	res, st.Err = authenticatePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// AuthenticatePersonPublicKeyHTTP is HTTP handler of AuthenticatePersonPublicKey service.
func AuthenticatePersonPublicKeyHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &authenticatePersonPublicKeyReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *authenticatePersonPublicKeyRes
	res, st.Err = authenticatePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.Body, st.Err = res.jsonEncoder()
	// st.Err make occur on just memory full!

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
}

type authenticatePersonPublicKeyReq struct {
	PersonID     [32]byte `json:",string"`
	PasswordHash [32]byte `json:",string"`
	OTP          uint32
	PublicKey    [32]byte `json:",string"`
	ExpireAt     uint64
	CaptchaID    [16]byte `json:",string"`
}

type authenticatePersonPublicKeyRes struct{}

func authenticatePersonPublicKey(st *achaemenid.Stream, req *authenticatePersonPublicKeyReq) (res *authenticatePersonPublicKeyRes, err *er.Error) {
	// TODO::: Authenticate request first by service policy.

	err = st.Authorize()
	if err != nil {
		return
	}

	// Request should send with UserID! Otherwise use identification||information service to find UserID.

	// get ThingID from st.Connection.PeerThingID

	// Person can have one PublicKey on same thing! so if other active public key is exist for requested thing,
	// we should notify user and get confirmed. If user not access to any their active connection, user must do recovery proccess first.

	// Notify user about failed authenticated proccess!

	res = &authenticatePersonPublicKeyRes{}

	return
}

/*
	Request Encoders & Decoders
*/

func (req *authenticatePersonPublicKeyReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *authenticatePersonPublicKeyReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

/*
	Response Encoders & Decoders
*/

func (res *authenticatePersonPublicKeyRes) syllabEncoder(buf []byte) {
	return
}

func (res *authenticatePersonPublicKeyRes) syllabStackLen() (ln uint32) {
	return 0
}

func (res *authenticatePersonPublicKeyRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *authenticatePersonPublicKeyRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *authenticatePersonPublicKeyRes) jsonEncoder() (buf []byte, err *er.Error) {
	buf, err = json.Marshal(res)
	return
}
