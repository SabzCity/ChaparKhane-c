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

var findWikiByTitleService = achaemenid.Service{
	ID:                4187610614,
	IssueDate:         1605026748,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Find Wiki By Title",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "Find Wiki IDs by given title",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: FindWikiByTitleSRPC,
	HTTPHandler: FindWikiByTitleHTTP,
}

// FindWikiByTitleSRPC is sRPC handler of FindWikiByTitle service.
func FindWikiByTitleSRPC(st *achaemenid.Stream) {
	var req = &findWikiByTitleReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findWikiByTitleRes
	res, st.Err = findWikiByTitle(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindWikiByTitleHTTP is HTTP handler of FindWikiByTitle service.
func FindWikiByTitleHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findWikiByTitleReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findWikiByTitleRes
	res, st.Err = findWikiByTitle(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findWikiByTitleReq struct {
	Title  string
	Offset uint64
	Limit  uint64
}

type findWikiByTitleRes struct {
	IDs [][32]byte `json:",string"`
}

func findWikiByTitle(st *achaemenid.Stream, req *findWikiByTitleReq) (res *findWikiByTitleRes, err *er.Error) {
	var w = datastore.Wiki{
		Title: req.Title,
	}
	var indexRes [][32]byte
	indexRes, err = w.GetIDsByTitleByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findWikiByTitleRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findWikiByTitleReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Title = syllab.UnsafeGetString(buf, 0)
	req.Offset = syllab.GetUInt64(buf, 8)
	req.Limit = syllab.GetUInt64(buf, 16)
	return
}

func (req *findWikiByTitleReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.Title, 0, hsi)
	syllab.SetUInt64(buf, 8, req.Offset)
	syllab.SetUInt64(buf, 16, req.Limit)
	return
}

func (req *findWikiByTitleReq) syllabStackLen() (ln uint32) {
	return 24
}

func (req *findWikiByTitleReq) syllabHeapLen() (ln uint32) {
	ln = uint32(len(req.Title))
	return
}

func (req *findWikiByTitleReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findWikiByTitleReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'T':
			decoder.SetFounded()
			decoder.Offset(8)
			req.Title = decoder.DecodeString()
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

func (req *findWikiByTitleReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Title":"`)
	encoder.EncodeString(req.Title)

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findWikiByTitleReq) jsonLen() (ln int) {
	ln = len(req.Title)
	ln += 71
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findWikiByTitleRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findWikiByTitleRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findWikiByTitleRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findWikiByTitleRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findWikiByTitleRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findWikiByTitleRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findWikiByTitleRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findWikiByTitleRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
