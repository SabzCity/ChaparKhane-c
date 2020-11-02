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

var getUserAppConnectionsIDService = achaemenid.Service{
	ID:                452721290,
	URI:               "", // API services can set like "/apis?452721290" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDRead,
	IssueDate:         1603793415,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get User App Connections ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"UserAppsConnection",
	},

	SRPCHandler: GetUserAppConnectionsIDSRPC,
	HTTPHandler: GetUserAppConnectionsIDHTTP,
}

// GetUserAppConnectionsIDSRPC is sRPC handler of GetUserAppConnectionsID service.
func GetUserAppConnectionsIDSRPC(st *achaemenid.Stream) {
	var req = &getUserAppConnectionsIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getUserAppConnectionsIDRes
	res, st.Err = getUserAppConnectionsID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetUserAppConnectionsIDHTTP is HTTP handler of GetUserAppConnectionsID service.
func GetUserAppConnectionsIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getUserAppConnectionsIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getUserAppConnectionsIDRes
	res, st.Err = getUserAppConnectionsID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getUserAppConnectionsIDReq struct {
	Offset uint64
	Limit  uint64
}

type getUserAppConnectionsIDRes struct {
	IDs [][32]byte `json:",string"`
}

func getUserAppConnectionsID(st *achaemenid.Stream, req *getUserAppConnectionsIDReq) (res *getUserAppConnectionsIDRes, err *er.Error) {
	if st.Connection.UserType == achaemenid.UserTypeGuest {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppsConnection{
		UserID: st.Connection.UserID,
	}
	var indexRes [][32]byte
	indexRes, err = uac.GetIDsByUserID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &getUserAppConnectionsIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getUserAppConnectionsIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Offset = syllab.GetUInt64(buf, 0)
	req.Limit = syllab.GetUInt64(buf, 8)
	return
}

func (req *getUserAppConnectionsIDReq) syllabEncoder(buf []byte) {
	syllab.SetUInt64(buf, 0, req.Offset)
	syllab.SetUInt64(buf, 8, req.Limit)
	return
}

func (req *getUserAppConnectionsIDReq) syllabStackLen() (ln uint32) {
	return 16
}

func (req *getUserAppConnectionsIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getUserAppConnectionsIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getUserAppConnectionsIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'O':
			decoder.SetFounded()
			decoder.Offset(8)
			req.Offset, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
		case 'L':
			decoder.SetFounded()
			decoder.Offset(7)
			req.Limit, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getUserAppConnectionsIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Offset":`)
	encoder.EncodeUInt64(uint64(req.Offset))

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(uint64(req.Limit))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *getUserAppConnectionsIDReq) jsonLen() (ln int) {
	ln = 60
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getUserAppConnectionsIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArrayArray(buf, 0)
	return
}

func (res *getUserAppConnectionsIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *getUserAppConnectionsIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getUserAppConnectionsIDRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.IDs))
	return
}

func (res *getUserAppConnectionsIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getUserAppConnectionsIDRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(6)
			res.IDs, err = decoder.Decode32ByteArraySliceAsBase64()
			if err != nil {
				return
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getUserAppConnectionsIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *getUserAppConnectionsIDRes) jsonLen() (ln int) {
	ln += len(res.IDs) * 46
	ln += 10
	return
}
