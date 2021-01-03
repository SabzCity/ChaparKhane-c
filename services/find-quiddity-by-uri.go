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

var findQuiddityByURIService = achaemenid.Service{
	ID:                4276179545,
	IssueDate:         1605026726,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Quiddity By URI",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: FindQuiddityByURISRPC,
	HTTPHandler: FindQuiddityByURIHTTP,
}

// FindQuiddityByURISRPC is sRPC handler of FindQuiddityByURI service.
func FindQuiddityByURISRPC(st *achaemenid.Stream) {
	var req = &findQuiddityByURIReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findQuiddityByURIRes
	res, st.Err = findQuiddityByURI(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindQuiddityByURIHTTP is HTTP handler of FindQuiddityByURI service.
func FindQuiddityByURIHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findQuiddityByURIReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findQuiddityByURIRes
	res, st.Err = findQuiddityByURI(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findQuiddityByURIReq struct {
	URI    string
	Offset uint64
	Limit  uint64
}

type findQuiddityByURIRes struct {
	IDs [][32]byte `json:",string"`
}

func findQuiddityByURI(st *achaemenid.Stream, req *findQuiddityByURIReq) (res *findQuiddityByURIRes, err *er.Error) {
	var q = datastore.Quiddity{
		URI: req.URI,
	}
	var indexRes [][32]byte
	indexRes, err = q.FindIDsByURI(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findQuiddityByURIRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findQuiddityByURIReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.URI = syllab.UnsafeGetString(buf, 0)
	req.Offset = syllab.GetUInt64(buf, 8)
	req.Limit = syllab.GetUInt64(buf, 16)
	return
}

func (req *findQuiddityByURIReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.URI, 0, hsi)
	syllab.SetUInt64(buf, 8, req.Offset)
	syllab.SetUInt64(buf, 16, req.Limit)
	return
}

func (req *findQuiddityByURIReq) syllabStackLen() (ln uint32) {
	return 24
}

func (req *findQuiddityByURIReq) syllabHeapLen() (ln uint32) {
	ln = uint32(len(req.URI))
	return
}

func (req *findQuiddityByURIReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findQuiddityByURIReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "URI":
			req.URI, err = decoder.DecodeString()
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

func (req *findQuiddityByURIReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"URI":"`)
	encoder.EncodeString(req.URI)

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findQuiddityByURIReq) jsonLen() (ln int) {
	ln = len(req.URI)
	ln += 69
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findQuiddityByURIRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findQuiddityByURIRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findQuiddityByURIRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findQuiddityByURIRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findQuiddityByURIRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findQuiddityByURIRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findQuiddityByURIRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findQuiddityByURIRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
