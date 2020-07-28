/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var approvePersonPublicKeyService = achaemenid.Service{
	ID:                157272426,
	URI:               "", // API services can set like "/apis?157272426" but it is not efficient, find services by ID.
	Name:              "ApprovePersonPublicKey",
	IssueDate:         1592380051,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"Approve given key if given data valid or return error",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: ApprovePersonPublicKeySRPC,
	HTTPHandler: ApprovePersonPublicKeyHTTP,
}

// ApprovePersonPublicKeySRPC is sRPC handler of ApprovePersonPublicKey service.
func ApprovePersonPublicKeySRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &approvePersonPublicKeyReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *approvePersonPublicKeyRes
	res, st.ReqRes.Err = approvePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// ApprovePersonPublicKeyHTTP is HTTP handler of ApprovePersonPublicKey service.
func ApprovePersonPublicKeyHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &approvePersonPublicKeyReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *approvePersonPublicKeyRes
	res, st.ReqRes.Err = approvePersonPublicKey(st, req)
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

type approvePersonPublicKeyReq struct {
	PersonID  [16]byte
	PublicKey [32]byte
	ThingID   [16]byte
}

type approvePersonPublicKeyRes struct {
	Status   uint8
	IssueAt  uint64
	ExpireAt uint64
}

func approvePersonPublicKey(st *achaemenid.Stream, req *approvePersonPublicKeyReq) (res *approvePersonPublicKeyRes, err error) {
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

	// Just can call by registered org

	res = &approvePersonPublicKeyRes{}

	return
}

func (req *approvePersonPublicKeyReq) validator() (err error) {
	return
}

func (req *approvePersonPublicKeyReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *approvePersonPublicKeyReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *approvePersonPublicKeyRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *approvePersonPublicKeyRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
