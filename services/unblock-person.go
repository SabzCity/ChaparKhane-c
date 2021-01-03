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

var unblockPersonService = achaemenid.Service{
	ID:                3479974432,
	IssueDate:         1592391153,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypePerson,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "UnblockPerson",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Just judges (justice service) can request to un-block a person and in many un-blocking level",
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: UnblockPersonSRPC,
	HTTPHandler: UnblockPersonHTTP,
}

// UnblockPersonSRPC is sRPC handler of UnblockPerson service.
func UnblockPersonSRPC(st *achaemenid.Stream) {
	var req = &unblockPersonReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *unblockPersonRes
	res, st.Err = unblockPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// UnblockPersonHTTP is HTTP handler of UnblockPerson service.
func UnblockPersonHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &unblockPersonReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *unblockPersonRes
	res, st.Err = unblockPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type unblockPersonReq struct{}

type unblockPersonRes struct{}

func unblockPerson(st *achaemenid.Stream, req *unblockPersonReq) (res *unblockPersonRes, err *er.Error) {
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

func (req *unblockPersonReq) validator() (err *er.Error) {
	return
}

func (req *unblockPersonReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *unblockPersonReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

func (res *unblockPersonRes) syllabEncoder(buf []byte) {

}

func (res *unblockPersonRes) syllabStackLen() (ln uint32) {
	return 0
}

func (res *unblockPersonRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *unblockPersonRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *unblockPersonRes) jsonEncoder() (buf []byte) {
	buf, _ = json.Marshal(res)
	return
}
