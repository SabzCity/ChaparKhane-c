/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var blockPersonService = achaemenid.Service{
	ID:                4173689325,
	URI:               "", // API services can set like "/apis?4173689325" but it is not efficient, find services by ID.
	Name:              "BlockPerson",
	IssueDate:         1592390222,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"Just judges (justice service) can request to block a person and in many blocking level",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: BlockPersonSRPC,
	HTTPHandler: BlockPersonHTTP,
}

// BlockPersonSRPC is sRPC handler of BlockPerson service.
func BlockPersonSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &blockPersonReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *blockPersonRes
	res, st.ReqRes.Err = blockPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// BlockPersonHTTP is HTTP handler of BlockPerson service.
func BlockPersonHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &blockPersonReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *blockPersonRes
	res, st.ReqRes.Err = blockPerson(st, req)
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

type blockPersonReq struct{}

type blockPersonRes struct{}

func blockPerson(st *achaemenid.Stream, req *blockPersonReq) (res *blockPersonRes, err error) {
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

	// have difference level of blocking
	// just judges (justice service) can do this operation

	res = &blockPersonRes{}

	return
}

func (req *blockPersonReq) validator() (err error) {
	return
}

func (req *blockPersonReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *blockPersonReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *blockPersonRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *blockPersonRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
