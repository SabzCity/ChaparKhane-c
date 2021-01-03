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

var changePersonPasswordService = achaemenid.Service{
	ID:                4266514739,
	IssueDate:         1592389987,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "ChangePersonPassword",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Change password in active person connection not to recover account!",
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: ChangePersonPasswordSRPC,
	HTTPHandler: ChangePersonPasswordHTTP,
}

// ChangePersonPasswordSRPC is sRPC handler of ChangePersonPassword service.
func ChangePersonPasswordSRPC(st *achaemenid.Stream) {
	var req = &changePersonPasswordReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *changePersonPasswordRes
	res, st.Err = changePersonPassword(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// ChangePersonPasswordHTTP is HTTP handler of ChangePersonPassword service.
func ChangePersonPasswordHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &changePersonPasswordReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *changePersonPasswordRes
	res, st.Err = changePersonPassword(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type changePersonPasswordReq struct {
	OldPassword [32]byte `valid:"Password" json:",string"`
	NewPassword [32]byte `valid:"Password" json:",string"`
	OTP         uint32
}

type changePersonPasswordRes struct{}

func changePersonPassword(st *achaemenid.Stream, req *changePersonPasswordReq) (res *changePersonPasswordRes, err *er.Error) {
	// TODO::: Authenticate & Authorizing request first by service policy.

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

func (req *changePersonPasswordReq) validator() (err *er.Error) {
	return
}

func (req *changePersonPasswordReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *changePersonPasswordReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

func (res *changePersonPasswordRes) syllabEncoder(buf []byte) {
	return
}

func (res *changePersonPasswordRes) syllabStackLen() (ln uint32) {
	return 0
}

func (res *changePersonPasswordRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *changePersonPasswordRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *changePersonPasswordRes) jsonEncoder() (buf []byte) {
	buf, _ = json.Marshal(res)
	return
}
