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

var blockPersonService = achaemenid.Service{
	ID:                4173689325,
	IssueDate:         1592390222,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "BlockPerson",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Just judges (justice service) can request to block a person and in many blocking level",
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: BlockPersonSRPC,
	HTTPHandler: BlockPersonHTTP,
}

// BlockPersonSRPC is sRPC handler of BlockPerson service.
func BlockPersonSRPC(st *achaemenid.Stream) {
	var req = &blockPersonReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *blockPersonRes
	res, st.Err = blockPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// BlockPersonHTTP is HTTP handler of BlockPerson service.
func BlockPersonHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &blockPersonReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *blockPersonRes
	res, st.Err = blockPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type blockPersonReq struct{}

type blockPersonRes struct{}

func blockPerson(st *achaemenid.Stream, req *blockPersonReq) (res *blockPersonRes, err *er.Error) {
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

func (req *blockPersonReq) validator() (err *er.Error) {
	return
}

func (req *blockPersonReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *blockPersonReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

func (res *blockPersonRes) syllabEncoder(buf []byte) {
	return
}

func (res *blockPersonRes) syllabStackLen() (ln uint32) {
	return 0
}

func (res *blockPersonRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *blockPersonRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *blockPersonRes) jsonEncoder() (buf []byte) {
	buf, _ = json.Marshal(res)
	return
}
