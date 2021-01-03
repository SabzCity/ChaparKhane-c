/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/ganjine"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var findUserAppConnectionByGottenDelegateService = achaemenid.Service{
	ID:                248852948,
	IssueDate:         1603815653,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find User App Connection By Gotten Delegate",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"UserAppConnection",
	},

	SRPCHandler: FindUserAppConnectionByGottenDelegateSRPC,
	HTTPHandler: FindUserAppConnectionByGottenDelegateHTTP,
}

// FindUserAppConnectionByGottenDelegateSRPC is sRPC handler of FindUserAppConnectionByGottenDelegate service.
func FindUserAppConnectionByGottenDelegateSRPC(st *achaemenid.Stream) {
	var req = &findUserAppConnectionByGottenDelegateReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findUserAppConnectionByGottenDelegateRes
	res, st.Err = findUserAppConnectionByGottenDelegate(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindUserAppConnectionByGottenDelegateHTTP is HTTP handler of FindUserAppConnectionByGottenDelegate service.
func FindUserAppConnectionByGottenDelegateHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findUserAppConnectionByGottenDelegateReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findUserAppConnectionByGottenDelegateRes
	res, st.Err = findUserAppConnectionByGottenDelegate(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findUserAppConnectionByGottenDelegateReq struct {
	Offset uint64
	Limit  uint64
}

type findUserAppConnectionByGottenDelegateRes struct {
	IDs [][32]byte `json:",string"`
}

func findUserAppConnectionByGottenDelegate(st *achaemenid.Stream, req *findUserAppConnectionByGottenDelegateReq) (res *findUserAppConnectionByGottenDelegateRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppConnection{
		DelegateUserID: st.Connection.UserID,
	}
	var indexRes [][32]byte
	indexRes, err = uac.FindIDsByGottenDelegate(req.Offset, req.Limit)
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = nil
	}
	if err != nil {
		return
	}

	res = &findUserAppConnectionByGottenDelegateRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findUserAppConnectionByGottenDelegateReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Offset = syllab.GetUInt64(buf, 0)
	req.Limit = syllab.GetUInt64(buf, 8)
	return
}

func (req *findUserAppConnectionByGottenDelegateReq) syllabEncoder(buf []byte) {
	syllab.SetUInt64(buf, 0, req.Offset)
	syllab.SetUInt64(buf, 8, req.Limit)
	return
}

func (req *findUserAppConnectionByGottenDelegateReq) syllabStackLen() (ln uint32) {
	return 16
}

func (req *findUserAppConnectionByGottenDelegateReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findUserAppConnectionByGottenDelegateReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findUserAppConnectionByGottenDelegateReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "Offset":
			req.Offset, err = decoder.DecodeUInt64()
		case "Limit":
			req.Limit, err = decoder.DecodeUInt64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *findUserAppConnectionByGottenDelegateReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findUserAppConnectionByGottenDelegateReq) jsonLen() (ln int) {
	ln = 60
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findUserAppConnectionByGottenDelegateRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 0)
	return
}

func (res *findUserAppConnectionByGottenDelegateRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findUserAppConnectionByGottenDelegateRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findUserAppConnectionByGottenDelegateRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.IDs) * 32)
	return
}

func (res *findUserAppConnectionByGottenDelegateRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findUserAppConnectionByGottenDelegateRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "IDs":
			res.IDs, err = decoder.Decode32ByteArraySliceAsBase64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *findUserAppConnectionByGottenDelegateRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findUserAppConnectionByGottenDelegateRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
