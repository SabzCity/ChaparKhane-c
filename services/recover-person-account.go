/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
)

var recoverPersonAccountService = achaemenid.Service{
	ID:                1615946586,
	URI:               "", // API services can set like "/apis?1615946586" but it is not efficient, find services by ID.
	Name:              "RecoverPersonAccount",
	IssueDate:         1592390624,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		`Recover user account due to lost password or lost all active devices with active connection!
		If user authenticated successfully in st.Connection.PeerThingID one time before can recover just by OTP
		Otherwise must send all information (OTP && RecoveryCode && SecurityQuestion) to recover!
		It will change user status to PersonNotForceUse2Factor!
		`,
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: RecoverPersonAccountSRPC,
	HTTPHandler: RecoverPersonAccountHTTP,
}

// RecoverPersonAccountSRPC is sRPC handler of RecoverPersonAccount service.
func RecoverPersonAccountSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &recoverPersonAccountReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *recoverPersonAccountRes
	res, st.ReqRes.Err = recoverPersonAccount(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// RecoverPersonAccountHTTP is HTTP handler of RecoverPersonAccount service.
func RecoverPersonAccountHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &recoverPersonAccountReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *recoverPersonAccountRes
	res, st.ReqRes.Err = recoverPersonAccount(st, req)
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

type recoverPersonAccountReq struct {
	Username         string
	NewPassword      [32]byte `valid:"Password"`
	OTP              uint32
	RecoveryCode     [128]byte
	SecurityQuestion uint16
	SecurityAnswer   string
}

type recoverPersonAccountRes struct{}

func recoverPersonAccount(st *achaemenid.Stream, req *recoverPersonAccountReq) (res *recoverPersonAccountRes, err error) {
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

func (req *recoverPersonAccountReq) validator() (err error) {
	return
}

func (req *recoverPersonAccountReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *recoverPersonAccountReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *recoverPersonAccountRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *recoverPersonAccountRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
