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

var getUserAppGivenDelegateConnectionsIDService = achaemenid.Service{
	ID:                469133894,
	URI:               "", // API services can set like "/apis?469133894" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDRead,
	IssueDate:         1603815639,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get User App Given Delegate Connections ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"UserAppsConnection",
	},

	SRPCHandler: GetUserAppGivenDelegateConnectionsIDSRPC,
	HTTPHandler: GetUserAppGivenDelegateConnectionsIDHTTP,
}

// GetUserAppGivenDelegateConnectionsIDSRPC is sRPC handler of GetUserAppGivenDelegateConnectionsID service.
func GetUserAppGivenDelegateConnectionsIDSRPC(st *achaemenid.Stream) {
	var req = &getUserAppGivenDelegateConnectionsIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getUserAppGivenDelegateConnectionsIDRes
	res, st.Err = getUserAppGivenDelegateConnectionsID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetUserAppGivenDelegateConnectionsIDHTTP is HTTP handler of GetUserAppGivenDelegateConnectionsID service.
func GetUserAppGivenDelegateConnectionsIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getUserAppGivenDelegateConnectionsIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getUserAppGivenDelegateConnectionsIDRes
	res, st.Err = getUserAppGivenDelegateConnectionsID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getUserAppGivenDelegateConnectionsIDReq struct {
	Offset uint64
	Limit  uint64
}

type getUserAppGivenDelegateConnectionsIDRes struct {
	IDs [][32]byte `json:",string"`
}

func getUserAppGivenDelegateConnectionsID(st *achaemenid.Stream, req *getUserAppGivenDelegateConnectionsIDReq) (res *getUserAppGivenDelegateConnectionsIDRes, err *er.Error) {
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
	indexRes, err = uac.GetIDsByGivenDelegate(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &getUserAppGivenDelegateConnectionsIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getUserAppGivenDelegateConnectionsIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Offset = syllab.GetUInt64(buf, 0)
	req.Limit = syllab.GetUInt64(buf, 8)
	return
}

func (req *getUserAppGivenDelegateConnectionsIDReq) syllabEncoder(buf []byte) {
	syllab.SetUInt64(buf, 0, req.Offset)
	syllab.SetUInt64(buf, 8, req.Limit)
	return
}

func (req *getUserAppGivenDelegateConnectionsIDReq) syllabStackLen() (ln uint32) {
	return 16
}

func (req *getUserAppGivenDelegateConnectionsIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getUserAppGivenDelegateConnectionsIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getUserAppGivenDelegateConnectionsIDReq) jsonDecoder(buf []byte) (err *er.Error) {
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

func (req *getUserAppGivenDelegateConnectionsIDReq) jsonEncoder() (buf []byte) {
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

func (req *getUserAppGivenDelegateConnectionsIDReq) jsonLen() (ln int) {
	ln = 60
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getUserAppGivenDelegateConnectionsIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArrayArray(buf, 0)
	return
}

func (res *getUserAppGivenDelegateConnectionsIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *getUserAppGivenDelegateConnectionsIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getUserAppGivenDelegateConnectionsIDRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.IDs))
	return
}

func (res *getUserAppGivenDelegateConnectionsIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getUserAppGivenDelegateConnectionsIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *getUserAppGivenDelegateConnectionsIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *getUserAppGivenDelegateConnectionsIDRes) jsonLen() (ln int) {
	ln += len(res.IDs) * 46
	ln += 10
	return
}
