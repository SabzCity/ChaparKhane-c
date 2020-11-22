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

var findProductAuctionByGroupIDService = achaemenid.Service{
	ID:                2073537094,
	IssueDate:         1605202753,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Find Product Auction By Group ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByGroupIDSRPC,
	HTTPHandler: FindProductAuctionByGroupIDHTTP,
}

// FindProductAuctionByGroupIDSRPC is sRPC handler of FindProductAuctionByGroupID service.
func FindProductAuctionByGroupIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByGroupIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByGroupIDRes
	res, st.Err = findProductAuctionByGroupID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByGroupIDHTTP is HTTP handler of FindProductAuctionByGroupID service.
func FindProductAuctionByGroupIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByGroupIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByGroupIDRes
	res, st.Err = findProductAuctionByGroupID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByGroupIDReq struct {
	GroupID [32]byte `json:",string"`
	Offset  uint64
	Limit   uint64
}

type findProductAuctionByGroupIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByGroupID(st *achaemenid.Stream, req *findProductAuctionByGroupIDReq) (res *findProductAuctionByGroupIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		GroupID: req.GroupID,
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByGroupIDByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByGroupIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByGroupIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.GroupID[:], buf[0:])
	req.Offset = syllab.GetUInt64(buf, 32)
	req.Limit = syllab.GetUInt64(buf, 40)
	return
}

func (req *findProductAuctionByGroupIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.GroupID[:])
	syllab.SetUInt64(buf, 32, req.Offset)
	syllab.SetUInt64(buf, 40, req.Limit)
	return
}

func (req *findProductAuctionByGroupIDReq) syllabStackLen() (ln uint32) {
	return 48
}

func (req *findProductAuctionByGroupIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByGroupIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByGroupIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
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

func (req *findProductAuctionByGroupIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"GroupID":"`)
	encoder.EncodeByteSliceAsBase64(req.GroupID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByGroupIDReq) jsonLen() (ln int) {
	ln = 116
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByGroupIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByGroupIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByGroupIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByGroupIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByGroupIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByGroupIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByGroupIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByGroupIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
