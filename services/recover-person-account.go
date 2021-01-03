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

var recoverPersonAccountService = achaemenid.Service{
	ID:                1615946586,
	IssueDate:         1592390624,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "RecoverPersonAccount",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: `Recover user account due to lost password or lost all active devices with active connection!
If user authenticated successfully in st.Connection.PeerThingID one time before can recover just by OTP
Otherwise must send all information (OTP && SecurityKeyOTP) to recover!
It will change user status to PersonNotForceUse2Factor!`,
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: RecoverPersonAccountSRPC,
	HTTPHandler: RecoverPersonAccountHTTP,
}

// RecoverPersonAccountSRPC is sRPC handler of RecoverPersonAccount service.
func RecoverPersonAccountSRPC(st *achaemenid.Stream) {
	var req = &recoverPersonAccountReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *recoverPersonAccountRes
	res, st.Err = recoverPersonAccount(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RecoverPersonAccountHTTP is HTTP handler of RecoverPersonAccount service.
func RecoverPersonAccountHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &recoverPersonAccountReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *recoverPersonAccountRes
	res, st.Err = recoverPersonAccount(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type recoverPersonAccountReq struct {
	Username       string
	NewPassword    [32]byte `valid:"Password" json:",string"`
	OTP            uint32
	SecurityKeyOTP uint32
}

type recoverPersonAccountRes struct{}

func recoverPersonAccount(st *achaemenid.Stream, req *recoverPersonAccountReq) (res *recoverPersonAccountRes, err *er.Error) {
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

	res = &recoverPersonAccountRes{}

	return
}

func (req *recoverPersonAccountReq) validator() (err *er.Error) {
	return
}

func (req *recoverPersonAccountReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *recoverPersonAccountReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

func (res *recoverPersonAccountRes) syllabEncoder(buf []byte) {

}

func (res *recoverPersonAccountRes) syllabStackLen() (ln uint32) {
	return 0
}

func (res *recoverPersonAccountRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *recoverPersonAccountRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *recoverPersonAccountRes) jsonEncoder() (buf []byte) {
	buf, _ = json.Marshal(res)
	return
}
