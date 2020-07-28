/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var changePersonPasswordService = achaemenid.Service{
	ID:                4266514739,
	URI:               "", // API services can set like "/apis?4266514739" but it is not efficient, find services by ID.
	Name:              "ChangePersonPassword",
	IssueDate:         1592389987,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"Change password in active person connection not to recover account!",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: ChangePersonPasswordSRPC,
	HTTPHandler: ChangePersonPasswordHTTP,
}

// ChangePersonPasswordSRPC is sRPC handler of ChangePersonPassword service.
func ChangePersonPasswordSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &changePersonPasswordReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *changePersonPasswordRes
	res, st.ReqRes.Err = changePersonPassword(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// ChangePersonPasswordHTTP is HTTP handler of ChangePersonPassword service.
func ChangePersonPasswordHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &changePersonPasswordReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *changePersonPasswordRes
	res, st.ReqRes.Err = changePersonPassword(st, req)
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

type changePersonPasswordReq struct{
	OldPassword [32]byte `valid:"Password"`
	NewPassword [32]byte `valid:"Password"`
	OTP         uint32
}

type changePersonPasswordRes struct{}

func changePersonPassword(st *achaemenid.Stream, req *changePersonPasswordReq) (res *changePersonPasswordRes, err error) {
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

	// this service can use just for active user so we use st.Connection.OwnerID for personID
	// This service can't use on Delegate connection so check st.Connection.DelegateUserID

	// check OTP if it activated by user OTPPattern + OTPAdditional

	res = &changePersonPasswordRes{}

	return
}

func (req *changePersonPasswordReq) validator() (err error) {
	return
}

func (req *changePersonPasswordReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *changePersonPasswordReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *changePersonPasswordRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *changePersonPasswordRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
