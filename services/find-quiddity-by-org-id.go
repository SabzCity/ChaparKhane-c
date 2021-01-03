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

var findQuiddityByOrgIDService = achaemenid.Service{
	ID:                2290794554,
	IssueDate:         1605027929,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Quiddity By Org ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: FindQuiddityByOrgIDSRPC,
	HTTPHandler: FindQuiddityByOrgIDHTTP,
}

// FindQuiddityByOrgIDSRPC is sRPC handler of FindQuiddityByOrgID service.
func FindQuiddityByOrgIDSRPC(st *achaemenid.Stream) {
	var req = &findQuiddityByOrgIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findQuiddityByOrgIDRes
	res, st.Err = findQuiddityByOrgID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindQuiddityByOrgIDHTTP is HTTP handler of FindQuiddityByOrgID service.
func FindQuiddityByOrgIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findQuiddityByOrgIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findQuiddityByOrgIDRes
	res, st.Err = findQuiddityByOrgID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findQuiddityByOrgIDReq struct {
	OrgID  [32]byte `json:",string"`
	Offset uint64
	Limit  uint64
}

type findQuiddityByOrgIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findQuiddityByOrgID(st *achaemenid.Stream, req *findQuiddityByOrgIDReq) (res *findQuiddityByOrgIDRes, err *er.Error) {
	var w = datastore.Quiddity{
		OrgID: req.OrgID,
	}
	var indexRes [][32]byte
	indexRes, err = w.FindIDsByOrgID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findQuiddityByOrgIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findQuiddityByOrgIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.OrgID[:], buf[0:])
	req.Offset = syllab.GetUInt64(buf, 32)
	req.Limit = syllab.GetUInt64(buf, 40)
	return
}

func (req *findQuiddityByOrgIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.OrgID[:])
	syllab.SetUInt64(buf, 32, req.Offset)
	syllab.SetUInt64(buf, 40, req.Limit)
	return
}

func (req *findQuiddityByOrgIDReq) syllabStackLen() (ln uint32) {
	return 48
}

func (req *findQuiddityByOrgIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findQuiddityByOrgIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findQuiddityByOrgIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "OrgID":
			err = decoder.DecodeByteArrayAsBase64(req.OrgID[:])
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

func (req *findQuiddityByOrgIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"OrgID":"`)
	encoder.EncodeByteSliceAsBase64(req.OrgID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findQuiddityByOrgIDReq) jsonLen() (ln int) {
	ln = 114
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findQuiddityByOrgIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findQuiddityByOrgIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findQuiddityByOrgIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findQuiddityByOrgIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findQuiddityByOrgIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findQuiddityByOrgIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findQuiddityByOrgIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findQuiddityByOrgIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
