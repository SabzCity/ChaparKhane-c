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

var revokePersonPublicKeyService = achaemenid.Service{
	ID:                1775473172,
	IssueDate:         1592390799,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypePerson,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "RevokePersonPublicKey",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: `Change status of public key and expire it. Due to legal history purpose never delete any public key.
PersonID in st.Connection.OwnerID same as stored one in datastore!
This service will check person authentication status force use OTP or not.`,
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: RevokePersonPublicKeySRPC,
	HTTPHandler: RevokePersonPublicKeyHTTP,
}

// RevokePersonPublicKeySRPC is sRPC handler of RevokePersonPublicKey service.
func RevokePersonPublicKeySRPC(st *achaemenid.Stream) {
	var req = &revokePersonPublicKeyReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *revokePersonPublicKeyRes
	res, st.Err = revokePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RevokePersonPublicKeyHTTP is HTTP handler of RevokePersonPublicKey service.
func RevokePersonPublicKeyHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &revokePersonPublicKeyReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *revokePersonPublicKeyRes
	res, st.Err = revokePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type revokePersonPublicKeyReq struct {
	PublicKey [32]byte `valid:"PublicKey" json:",string"`
	Password  [32]byte `valid:"Password" json:",string"`
	OTP       uint32
}

type revokePersonPublicKeyRes struct{}

func revokePersonPublicKey(st *achaemenid.Stream, req *revokePersonPublicKeyReq) (res *revokePersonPublicKeyRes, err *er.Error) {
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

func (req *revokePersonPublicKeyReq) validator() (err *er.Error) {
	return
}

func (req *revokePersonPublicKeyReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *revokePersonPublicKeyReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

func (res *revokePersonPublicKeyRes) syllabEncoder(buf []byte) {
}

func (res *revokePersonPublicKeyRes) syllabStackLen() (ln uint32) {
	return 0
}

func (res *revokePersonPublicKeyRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *revokePersonPublicKeyRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *revokePersonPublicKeyRes) jsonEncoder() (buf []byte) {
	buf, _ = json.Marshal(res)
	return
}
