/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var unblockPersonService = achaemenid.Service{
	ID:                3479974432,
	URI:               "", // API services can set like "/apis?3479974432" but it is not efficient, find services by ID.
	Name:              "UnblockPerson",
	IssueDate:         1592391153,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"Just judges (justice service) can request to un-block a person and in many un-blocking level",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: UnblockPersonSRPC,
	HTTPHandler: UnblockPersonHTTP,
}

// UnblockPersonSRPC is sRPC handler of UnblockPerson service.
func UnblockPersonSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &unblockPersonReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *unblockPersonRes
	res, st.ReqRes.Err = unblockPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// UnblockPersonHTTP is HTTP handler of UnblockPerson service.
func UnblockPersonHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &unblockPersonReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *unblockPersonRes
	res, st.ReqRes.Err = unblockPerson(st, req)
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

type unblockPersonReq struct{}

type unblockPersonRes struct{}

func unblockPerson(st *achaemenid.Stream, req *unblockPersonReq) (res *unblockPersonRes, err error) {
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

	res = &unblockPersonRes{}

	return
}

func (req *unblockPersonReq) validator() (err error) {
	return
}

func (req *unblockPersonReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *unblockPersonReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *unblockPersonRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *unblockPersonRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
