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

var findUserAppConnectionService = achaemenid.Service{
	ID:                509105078,
	IssueDate:         1603793415,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find User App Connection",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"UserAppConnection",
	},

	SRPCHandler: FindUserAppConnectionSRPC,
	HTTPHandler: FindUserAppConnectionHTTP,
}

// FindUserAppConnectionSRPC is sRPC handler of FindUserAppConnection service.
func FindUserAppConnectionSRPC(st *achaemenid.Stream) {
	var req = &findUserAppConnectionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findUserAppConnectionRes
	res, st.Err = findUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindUserAppConnectionHTTP is HTTP handler of FindUserAppConnection service.
func FindUserAppConnectionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findUserAppConnectionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findUserAppConnectionRes
	res, st.Err = findUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findUserAppConnectionReq struct {
	Offset uint64
	Limit  uint64
}

type findUserAppConnectionRes struct {
	IDs [][32]byte `json:",string"`
}

func findUserAppConnection(st *achaemenid.Stream, req *findUserAppConnectionReq) (res *findUserAppConnectionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppConnection{
		UserID: st.Connection.UserID,
	}
	var indexRes [][32]byte
	indexRes, err = uac.FindIDsByUserID(req.Offset, req.Limit)
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = nil
	}
	if err != nil {
		return
	}

	res = &findUserAppConnectionRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findUserAppConnectionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Offset = syllab.GetUInt64(buf, 0)
	req.Limit = syllab.GetUInt64(buf, 8)
	return
}

func (req *findUserAppConnectionReq) syllabEncoder(buf []byte) {
	syllab.SetUInt64(buf, 0, req.Offset)
	syllab.SetUInt64(buf, 8, req.Limit)
	return
}

func (req *findUserAppConnectionReq) syllabStackLen() (ln uint32) {
	return 16
}

func (req *findUserAppConnectionReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findUserAppConnectionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findUserAppConnectionReq) jsonDecoder(buf []byte) (err *er.Error) {
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

func (req *findUserAppConnectionReq) jsonEncoder() (buf []byte) {
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

func (req *findUserAppConnectionReq) jsonLen() (ln int) {
	ln = 60
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findUserAppConnectionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 0)
	return
}

func (res *findUserAppConnectionRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findUserAppConnectionRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findUserAppConnectionRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.IDs) * 32)
	return
}

func (res *findUserAppConnectionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findUserAppConnectionRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findUserAppConnectionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findUserAppConnectionRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
