/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var revokePersonPublicKeyService = achaemenid.Service{
	ID:                1775473172,
	URI:               "", // API services can set like "/apis?1775473172" but it is not efficient, find services by ID.
	Name:              "RevokePersonPublicKey",
	IssueDate:         1592390799,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		`Change status of public key and expire it. Due to legal history purpose never delete any public key.
		PersonID in st.Connection.OwnerID same as stored one in datastore!
		This service will check person authentication status force use OTP or not.`,
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: RevokePersonPublicKeySRPC,
	HTTPHandler: RevokePersonPublicKeyHTTP,
}

// RevokePersonPublicKeySRPC is sRPC handler of RevokePersonPublicKey service.
func RevokePersonPublicKeySRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &revokePersonPublicKeyReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *revokePersonPublicKeyRes
	res, st.ReqRes.Err = revokePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// RevokePersonPublicKeyHTTP is HTTP handler of RevokePersonPublicKey service.
func RevokePersonPublicKeyHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &revokePersonPublicKeyReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *revokePersonPublicKeyRes
	res, st.ReqRes.Err = revokePersonPublicKey(st, req)
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

type revokePersonPublicKeyReq struct {
	PublicKey [32]byte `valid:"PublicKey"`
	Password  [32]byte `valid:"Password"`
	OTP       uint32
}

type revokePersonPublicKeyRes struct{}

func revokePersonPublicKey(st *achaemenid.Stream, req *revokePersonPublicKeyReq) (res *revokePersonPublicKeyRes, err error) {
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

	res = &revokePersonPublicKeyRes{}

	return
}

func (req *revokePersonPublicKeyReq) validator() (err error) {
	return
}

func (req *revokePersonPublicKeyReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *revokePersonPublicKeyReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *revokePersonPublicKeyRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *revokePersonPublicKeyRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
