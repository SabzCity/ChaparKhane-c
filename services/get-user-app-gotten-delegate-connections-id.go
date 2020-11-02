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

var getUserAppGottenDelegateConnectionsIDService = achaemenid.Service{
	ID:                2101786453,
	URI:               "", // API services can set like "/apis?2101786453" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDRead,
	IssueDate:         1603815653,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get User App Gotten Delegate Connections ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"UserAppsConnection",
	},

	SRPCHandler: GetUserAppGottenDelegateConnectionsIDSRPC,
	HTTPHandler: GetUserAppGottenDelegateConnectionsIDHTTP,
}

// GetUserAppGottenDelegateConnectionsIDSRPC is sRPC handler of GetUserAppGottenDelegateConnectionsID service.
func GetUserAppGottenDelegateConnectionsIDSRPC(st *achaemenid.Stream) {
	var req = &getUserAppGottenDelegateConnectionsIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getUserAppGottenDelegateConnectionsIDRes
	res, st.Err = getUserAppGottenDelegateConnectionsID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetUserAppGottenDelegateConnectionsIDHTTP is HTTP handler of GetUserAppGottenDelegateConnectionsID service.
func GetUserAppGottenDelegateConnectionsIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getUserAppGottenDelegateConnectionsIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getUserAppGottenDelegateConnectionsIDRes
	res, st.Err = getUserAppGottenDelegateConnectionsID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getUserAppGottenDelegateConnectionsIDReq struct {
	Offset uint64
	Limit  uint64
}

type getUserAppGottenDelegateConnectionsIDRes struct {
	IDs [][32]byte `json:",string"`
}

func getUserAppGottenDelegateConnectionsID(st *achaemenid.Stream, req *getUserAppGottenDelegateConnectionsIDReq) (res *getUserAppGottenDelegateConnectionsIDRes, err *er.Error) {
	if st.Connection.UserType == achaemenid.UserTypeGuest {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppsConnection{
		DelegateUserID: st.Connection.UserID,
	}
	var indexRes [][32]byte
	indexRes, err = uac.GetIDsByGottenDelegate(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &getUserAppGottenDelegateConnectionsIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getUserAppGottenDelegateConnectionsIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Offset = syllab.GetUInt64(buf, 0)
	req.Limit = syllab.GetUInt64(buf, 8)
	return
}

func (req *getUserAppGottenDelegateConnectionsIDReq) syllabEncoder(buf []byte) {
	syllab.SetUInt64(buf, 0, req.Offset)
	syllab.SetUInt64(buf, 8, req.Limit)
	return
}

func (req *getUserAppGottenDelegateConnectionsIDReq) syllabStackLen() (ln uint32) {
	return 16
}

func (req *getUserAppGottenDelegateConnectionsIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getUserAppGottenDelegateConnectionsIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getUserAppGottenDelegateConnectionsIDReq) jsonDecoder(buf []byte) (err *er.Error) {
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

func (req *getUserAppGottenDelegateConnectionsIDReq) jsonEncoder() (buf []byte) {
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

func (req *getUserAppGottenDelegateConnectionsIDReq) jsonLen() (ln int) {
	ln = 60
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getUserAppGottenDelegateConnectionsIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArrayArray(buf, 0)
	return
}

func (res *getUserAppGottenDelegateConnectionsIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *getUserAppGottenDelegateConnectionsIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getUserAppGottenDelegateConnectionsIDRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.IDs))
	return
}

func (res *getUserAppGottenDelegateConnectionsIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getUserAppGottenDelegateConnectionsIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *getUserAppGottenDelegateConnectionsIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *getUserAppGottenDelegateConnectionsIDRes) jsonLen() (ln int) {
	ln += len(res.IDs) * 46
	ln += 10
	return
}
