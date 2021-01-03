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

var getPersonStatusService = achaemenid.Service{
	ID:                4031021741,
	IssueDate:         1604744631,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest, // TODO::: allow to user circles get person status??
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Person Status",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"PersonAuthentication",
	},

	SRPCHandler: GetPersonStatusSRPC,
	HTTPHandler: GetPersonStatusHTTP,
}

// GetPersonStatusSRPC is sRPC handler of GetPersonStatus service.
func GetPersonStatusSRPC(st *achaemenid.Stream) {
	var req = &getPersonStatusReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getPersonStatusRes
	res, st.Err = getPersonStatus(st, req, false)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetPersonStatusHTTP is HTTP handler of GetPersonStatus service.
func GetPersonStatusHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getPersonStatusReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getPersonStatusRes
	res, st.Err = getPersonStatus(st, req, false)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getPersonStatusReq struct {
	PersonID [32]byte `json:",string"`
}

type getPersonStatusRes struct {
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`

	ReferentPersonID [32]byte `json:",string"`
	Status           datastore.PersonAuthenticationStatus
}

func getPersonStatus(st *achaemenid.Stream, req *getPersonStatusReq, unsafe bool) (res *getPersonStatusRes, err *er.Error) {
	if !unsafe {
		err = st.Authorize()
		if err != nil {
			return
		}
	}

	var pa = datastore.PersonAuthentication{
		PersonID: req.PersonID,
	}
	err = pa.GetLastByPersonID()
	if err != nil {
		return
	}

	res = &getPersonStatusRes{
		AppInstanceID: pa.AppInstanceID,
		// UserConnectionID: pa.UserConnectionID, TODO::: Due to HTTP use ConnectionID to authenticate connections can't enable it now!!
		ReferentPersonID: pa.ReferentPersonID,
		Status:           pa.Status,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getPersonStatusReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.PersonID[:], buf[0:])
	return
}

func (req *getPersonStatusReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.PersonID[:])
	return
}

func (req *getPersonStatusReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getPersonStatusReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getPersonStatusReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getPersonStatusReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "PersonID":
			err = decoder.DecodeByteArrayAsBase64(req.PersonID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *getPersonStatusReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"PersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.PersonID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getPersonStatusReq) jsonLen() (ln int) {
	ln = 58
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getPersonStatusRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.AppInstanceID[:], buf[0:])
	copy(res.UserConnectionID[:], buf[32:])
	copy(res.ReferentPersonID[:], buf[64:])
	res.Status = datastore.PersonAuthenticationStatus(syllab.GetUInt8(buf, 96))
	return
}

func (res *getPersonStatusRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.AppInstanceID[:])
	copy(buf[32:], res.UserConnectionID[:])
	copy(buf[64:], res.ReferentPersonID[:])
	syllab.SetUInt8(buf, 96, uint8(res.Status))
	return
}

func (res *getPersonStatusRes) syllabStackLen() (ln uint32) {
	return 97
}

func (res *getPersonStatusRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getPersonStatusRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getPersonStatusRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "AppInstanceID":
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
		case "UserConnectionID":
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
		case "ReferentPersonID":
			err = decoder.DecodeByteArrayAsBase64(res.ReferentPersonID[:])
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Status = datastore.PersonAuthenticationStatus(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getPersonStatusRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","ReferentPersonID":"`)
	encoder.EncodeByteSliceAsBase64(res.ReferentPersonID[:])

	encoder.EncodeString(`","Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getPersonStatusRes) jsonLen() (ln int) {
	ln = 206
	return
}
