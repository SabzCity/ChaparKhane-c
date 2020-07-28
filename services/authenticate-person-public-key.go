/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var authenticatePersonPublicKeyService = achaemenid.Service{
	ID:                846072085,
	URI:               "", // API services can set like "/apis?846072085" but it is not efficient, find services by ID.
	Name:              "AuthenticatePersonPublicKey",
	IssueDate:         1592379653,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"Authenticate user and store given PublicKey to use by user for any purpose!",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: AuthenticatePersonPublicKeySRPC,
	HTTPHandler: AuthenticatePersonPublicKeyHTTP,
}

// AuthenticatePersonPublicKeySRPC is sRPC handler of AuthenticatePersonPublicKey service.
func AuthenticatePersonPublicKeySRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &authenticatePersonPublicKeyReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		st.Connection.FailedPacketsReceived++
		// Attack??
		return
	}

	var res *authenticatePersonPublicKeyRes
	res, st.ReqRes.Err = authenticatePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		st.Connection.FailedServiceCall++
		// Attack??
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// AuthenticatePersonPublicKeyHTTP is HTTP handler of AuthenticatePersonPublicKey service.
func AuthenticatePersonPublicKeyHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &authenticatePersonPublicKeyReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *authenticatePersonPublicKeyRes
	res, st.ReqRes.Err = authenticatePersonPublicKey(st, req)
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

type authenticatePersonPublicKeyReq struct {
	PersonID     [16]byte `valid:"optional"`
	Username     string   `valid:"Username,optional"`
	PasswordHash [32]byte
	OTP          uint32
	PublicKey    [32]byte
	ExpireAt     uint64
	CaptchaID    [16]byte
}

type authenticatePersonPublicKeyRes struct{}

func authenticatePersonPublicKey(st *achaemenid.Stream, req *authenticatePersonPublicKeyReq) (res *authenticatePersonPublicKeyRes, err error) {
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

	// Request should send with UserID! Otherwise use identification||information service to find UserID.

	// get ThingID from st.Connection.PeerThingID

	// Person can have one PublicKey on same thing! so if other active public key is exist for requested thing,
	// we should notify user and get confirmed. If user not access to any their active connection, user must do recovery proccess first.

	// Notify user about failed authenticated proccess!

	res = &authenticatePersonPublicKeyRes{}

	return
}

func (req *authenticatePersonPublicKeyReq) validator() (err error) {
	return
}

func (req *authenticatePersonPublicKeyReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *authenticatePersonPublicKeyReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *authenticatePersonPublicKeyRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *authenticatePersonPublicKeyRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
