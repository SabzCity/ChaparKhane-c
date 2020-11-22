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

var findWikiByURIService = achaemenid.Service{
	ID:                3272467714,
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
		lang.EnglishLanguage: "Get Wiki By URI",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: FindWikiByURISRPC,
	HTTPHandler: FindWikiByURIHTTP,
}

// FindWikiByURISRPC is sRPC handler of FindWikiByURI service.
func FindWikiByURISRPC(st *achaemenid.Stream) {
	var req = &findWikiByURIReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findWikiByURIRes
	res, st.Err = findWikiByURI(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindWikiByURIHTTP is HTTP handler of FindWikiByURI service.
func FindWikiByURIHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findWikiByURIReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findWikiByURIRes
	res, st.Err = findWikiByURI(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findWikiByURIReq struct {
	URI    string
	Offset uint64
	Limit  uint64
}

type findWikiByURIRes struct {
	IDs [][32]byte `json:",string"`
}

func findWikiByURI(st *achaemenid.Stream, req *findWikiByURIReq) (res *findWikiByURIRes, err *er.Error) {
	var w = datastore.Wiki{
		URI: req.URI,
	}
	var indexRes [][32]byte
	indexRes, err = w.GetIDsByURIByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findWikiByURIRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findWikiByURIReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.URI = syllab.UnsafeGetString(buf, 0)
	req.Offset = syllab.GetUInt64(buf, 8)
	req.Limit = syllab.GetUInt64(buf, 16)
	return
}

func (req *findWikiByURIReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.URI, 0, hsi)
	syllab.SetUInt64(buf, 8, req.Offset)
	syllab.SetUInt64(buf, 16, req.Limit)
	return
}

func (req *findWikiByURIReq) syllabStackLen() (ln uint32) {
	return 24
}

func (req *findWikiByURIReq) syllabHeapLen() (ln uint32) {
	ln = uint32(len(req.URI))
	return
}

func (req *findWikiByURIReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findWikiByURIReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'U':
			decoder.SetFounded()
			decoder.Offset(6)
			req.URI = decoder.DecodeString()
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

func (req *findWikiByURIReq) jsonEncoder() (buf []byte) {
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

func (req *findWikiByURIReq) jsonLen() (ln int) {
	ln += 0 + len(req.URI)
	ln += 69
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findWikiByURIRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findWikiByURIRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findWikiByURIRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findWikiByURIRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findWikiByURIRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findWikiByURIRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findWikiByURIRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findWikiByURIRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
