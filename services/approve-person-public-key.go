/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var approvePersonPublicKeyService = achaemenid.Service{
	ID:                157272426,
	IssueDate:         1592380051,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "ApprovePersonPublicKey",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Approve given key if given data valid or return error",
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: ApprovePersonPublicKeySRPC,
	HTTPHandler: ApprovePersonPublicKeyHTTP,
}

// ApprovePersonPublicKeySRPC is sRPC handler of ApprovePersonPublicKey service.
func ApprovePersonPublicKeySRPC(st *achaemenid.Stream) {
	var req = &approvePersonPublicKeyReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *approvePersonPublicKeyRes
	res, st.Err = approvePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// ApprovePersonPublicKeyHTTP is HTTP handler of ApprovePersonPublicKey service.
func ApprovePersonPublicKeyHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &approvePersonPublicKeyReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *approvePersonPublicKeyRes
	res, st.Err = approvePersonPublicKey(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type approvePersonPublicKeyReq struct {
	PersonID  [32]byte `json:",string"`
	PublicKey [32]byte `json:",string"`
	ThingID   [32]byte `json:",string"`
}

type approvePersonPublicKeyRes struct {
	Status datastore.PersonPublicKeyStatus
}

func approvePersonPublicKey(st *achaemenid.Stream, req *approvePersonPublicKeyReq) (res *approvePersonPublicKeyRes, err *er.Error) {
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

func (req *approvePersonPublicKeyReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *approvePersonPublicKeyReq) syllabDecoder(buf []byte) (err *er.Error) {
	return
}

func (req *approvePersonPublicKeyReq) jsonDecoder(buf []byte) (err *er.Error) {
	err = json.UnMarshal(buf, req)
	return
}

/*
	Response Encoders & Decoders
*/

func (res *approvePersonPublicKeyRes) syllabEncoder(buf []byte) {
	syllab.SetUInt8(buf, 0, uint8(res.Status))
}

func (res *approvePersonPublicKeyRes) syllabStackLen() (ln uint32) {
	return 1
}

func (res *approvePersonPublicKeyRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *approvePersonPublicKeyRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *approvePersonPublicKeyRes) jsonEncoder() (buf []byte) {
	buf, _ = json.Marshal(res)
	return
}
