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

var findProductAuctionByWikiIDGroupIDService = achaemenid.Service{
	ID:                719110014,
	IssueDate:         1605202810,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Find Product Auction By Wiki ID Group ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByWikiIDGroupIDSRPC,
	HTTPHandler: FindProductAuctionByWikiIDGroupIDHTTP,
}

// FindProductAuctionByWikiIDGroupIDSRPC is sRPC handler of FindProductAuctionByWikiIDGroupID service.
func FindProductAuctionByWikiIDGroupIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByWikiIDGroupIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByWikiIDGroupIDRes
	res, st.Err = findProductAuctionByWikiIDGroupID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByWikiIDGroupIDHTTP is HTTP handler of FindProductAuctionByWikiIDGroupID service.
func FindProductAuctionByWikiIDGroupIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByWikiIDGroupIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByWikiIDGroupIDRes
	res, st.Err = findProductAuctionByWikiIDGroupID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByWikiIDGroupIDReq struct {
	WikiID  [32]byte `json:",string"`
	GroupID [32]byte `json:",string"`
	Offset  uint64
	Limit   uint64
}

type findProductAuctionByWikiIDGroupIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByWikiIDGroupID(st *achaemenid.Stream, req *findProductAuctionByWikiIDGroupIDReq) (res *findProductAuctionByWikiIDGroupIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		WikiID:  req.WikiID,
		GroupID: req.GroupID,
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByWikiIDGroupIDByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByWikiIDGroupIDRes{
		IDs: indexRes,
	}
	return
}

func (req *findProductAuctionByWikiIDGroupIDReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByWikiIDGroupIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.WikiID[:], buf[0:])
	copy(req.GroupID[:], buf[32:])
	req.Offset = syllab.GetUInt64(buf, 64)
	req.Limit = syllab.GetUInt64(buf, 72)
	return
}

func (req *findProductAuctionByWikiIDGroupIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.WikiID[:])
	copy(buf[32:], req.GroupID[:])
	syllab.SetUInt64(buf, 64, req.Offset)
	syllab.SetUInt64(buf, 72, req.Limit)
	return
}

func (req *findProductAuctionByWikiIDGroupIDReq) syllabStackLen() (ln uint32) {
	return 80
}

func (req *findProductAuctionByWikiIDGroupIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByWikiIDGroupIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByWikiIDGroupIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'W':
			decoder.SetFounded()
			decoder.Offset(9)
			err = decoder.DecodeByteArrayAsBase64(req.WikiID[:])
			if err != nil {
				return
			}
		case 'G':
			decoder.SetFounded()
			decoder.Offset(10)
			err = decoder.DecodeByteArrayAsBase64(req.GroupID[:])
			if err != nil {
				return
			}
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

func (req *findProductAuctionByWikiIDGroupIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"WikiID":"`)
	encoder.EncodeByteSliceAsBase64(req.WikiID[:])

	encoder.EncodeString(`","GroupID":"`)
	encoder.EncodeByteSliceAsBase64(req.GroupID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByWikiIDGroupIDReq) jsonLen() (ln int) {
	ln = 171
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByWikiIDGroupIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByWikiIDGroupIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByWikiIDGroupIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByWikiIDGroupIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByWikiIDGroupIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByWikiIDGroupIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByWikiIDGroupIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByWikiIDGroupIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
