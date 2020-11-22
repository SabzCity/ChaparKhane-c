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

var findWikiByOrgIDService = achaemenid.Service{
	ID:                2965780360,
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
		lang.EnglishLanguage: "Get Wiki IDs By Org ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: FindWikiByOrgIDSRPC,
	HTTPHandler: FindWikiByOrgIDHTTP,
}

// FindWikiByOrgIDSRPC is sRPC handler of FindWikiByOrgID service.
func FindWikiByOrgIDSRPC(st *achaemenid.Stream) {
	var req = &findWikiByOrgIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findWikiByOrgIDRes
	res, st.Err = findWikiByOrgID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindWikiByOrgIDHTTP is HTTP handler of FindWikiByOrgID service.
func FindWikiByOrgIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findWikiByOrgIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findWikiByOrgIDRes
	res, st.Err = findWikiByOrgID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findWikiByOrgIDReq struct {
	OrgID  [32]byte `json:",string"`
	Offset uint64
	Limit  uint64
}

type findWikiByOrgIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findWikiByOrgID(st *achaemenid.Stream, req *findWikiByOrgIDReq) (res *findWikiByOrgIDRes, err *er.Error) {
	var w = datastore.Wiki{
		OrgID: req.OrgID,
	}
	var indexRes [][32]byte
	indexRes, err = w.FindIDsByOrgIDByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findWikiByOrgIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findWikiByOrgIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.OrgID[:], buf[0:])
	req.Offset = syllab.GetUInt64(buf, 32)
	req.Limit = syllab.GetUInt64(buf, 40)
	return
}

func (req *findWikiByOrgIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.OrgID[:])
	syllab.SetUInt64(buf, 32, req.Offset)
	syllab.SetUInt64(buf, 40, req.Limit)
	return
}

func (req *findWikiByOrgIDReq) syllabStackLen() (ln uint32) {
	return 48
}

func (req *findWikiByOrgIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findWikiByOrgIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findWikiByOrgIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'O':
			switch decoder.Buf[1] {
			case 'r':
				decoder.SetFounded()
				decoder.Offset(8)
				err = decoder.DecodeByteArrayAsBase64(req.OrgID[:])
				if err != nil {
					return
				}
			case 'f':
				decoder.SetFounded()
				decoder.Offset(8)
				req.Offset, err = decoder.DecodeUInt64()
				if err != nil {
					return
				}
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

func (req *findWikiByOrgIDReq) jsonEncoder() (buf []byte) {
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

func (req *findWikiByOrgIDReq) jsonLen() (ln int) {
	ln = 114
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findWikiByOrgIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findWikiByOrgIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findWikiByOrgIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findWikiByOrgIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findWikiByOrgIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findWikiByOrgIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findWikiByOrgIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findWikiByOrgIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
