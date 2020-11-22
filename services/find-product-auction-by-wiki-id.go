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
	"../libgo/price"
	"../libgo/srpc"
	"../libgo/syllab"
)

var findProductAuctionByWikiIDService = achaemenid.Service{
	ID:                1429674706,
	IssueDate:         1605202824,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Find Product Auction By Wiki ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByWikiIDSRPC,
	HTTPHandler: FindProductAuctionByWikiIDHTTP,
}

// FindProductAuctionByWikiIDSRPC is sRPC handler of FindProductAuctionByWikiID service.
func FindProductAuctionByWikiIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByWikiIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByWikiIDRes
	res, st.Err = findProductAuctionByWikiID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByWikiIDHTTP is HTTP handler of FindProductAuctionByWikiID service.
func FindProductAuctionByWikiIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByWikiIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByWikiIDRes
	res, st.Err = findProductAuctionByWikiID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByWikiIDReq struct {
	WikiID   [32]byte `json:",string"`
	Currency price.Currency
	Offset   uint64
	Limit    uint64
}

type findProductAuctionByWikiIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByWikiID(st *achaemenid.Stream, req *findProductAuctionByWikiIDReq) (res *findProductAuctionByWikiIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		WikiID:   req.WikiID,
		Currency: req.Currency,
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByWikiIDCurrencyByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByWikiIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByWikiIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.WikiID[:], buf[0:])
	req.Currency = price.Currency(syllab.GetUInt16(buf, 32))
	req.Offset = syllab.GetUInt64(buf, 34)
	req.Limit = syllab.GetUInt64(buf, 42)
	return
}

func (req *findProductAuctionByWikiIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.WikiID[:])
	syllab.SetUInt16(buf, 32, uint16(req.Currency))
	syllab.SetUInt64(buf, 34, req.Offset)
	syllab.SetUInt64(buf, 42, req.Limit)
	return
}

func (req *findProductAuctionByWikiIDReq) syllabStackLen() (ln uint32) {
	return 50
}

func (req *findProductAuctionByWikiIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByWikiIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByWikiIDReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		case 'C':
			decoder.SetFounded()
			decoder.Offset(10)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Currency = price.Currency(num)
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

func (req *findProductAuctionByWikiIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"WikiID":"`)
	encoder.EncodeByteSliceAsBase64(req.WikiID[:])

	encoder.EncodeString(`","Currency":`)
	encoder.EncodeUInt64(uint64(req.Currency))

	encoder.EncodeString(`,"Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByWikiIDReq) jsonLen() (ln int) {
	ln = 147
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByWikiIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByWikiIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByWikiIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByWikiIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByWikiIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByWikiIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByWikiIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByWikiIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
