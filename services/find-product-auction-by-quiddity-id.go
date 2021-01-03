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

var findProductAuctionByQuiddityIDService = achaemenid.Service{
	ID:                4261544056,
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
		lang.LanguageEnglish: "Find Product Auction By Quiddity ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByQuiddityIDSRPC,
	HTTPHandler: FindProductAuctionByQuiddityIDHTTP,
}

// FindProductAuctionByQuiddityIDSRPC is sRPC handler of FindProductAuctionByQuiddityID service.
func FindProductAuctionByQuiddityIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByQuiddityIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByQuiddityIDRes
	res, st.Err = findProductAuctionByQuiddityID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByQuiddityIDHTTP is HTTP handler of FindProductAuctionByQuiddityID service.
func FindProductAuctionByQuiddityIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByQuiddityIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByQuiddityIDRes
	res, st.Err = findProductAuctionByQuiddityID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByQuiddityIDReq struct {
	QuiddityID [32]byte `json:",string"`
	Offset uint64
	Limit  uint64
}

type findProductAuctionByQuiddityIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByQuiddityID(st *achaemenid.Stream, req *findProductAuctionByQuiddityIDReq) (res *findProductAuctionByQuiddityIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		QuiddityID: req.QuiddityID,
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByQuiddityID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByQuiddityIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByQuiddityIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])
	req.Offset = syllab.GetUInt64(buf, 32)
	req.Limit = syllab.GetUInt64(buf, 40)
	return
}

func (req *findProductAuctionByQuiddityIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])
	syllab.SetUInt64(buf, 32, req.Offset)
	syllab.SetUInt64(buf, 40, req.Limit)
	return
}

func (req *findProductAuctionByQuiddityIDReq) syllabStackLen() (ln uint32) {
	return 48
}

func (req *findProductAuctionByQuiddityIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByQuiddityIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByQuiddityIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])
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

func (req *findProductAuctionByQuiddityIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])

	encoder.EncodeString(`,"Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByQuiddityIDReq) jsonLen() (ln int) {
	ln = 115
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByQuiddityIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByQuiddityIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByQuiddityIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByQuiddityIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByQuiddityIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByQuiddityIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByQuiddityIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByQuiddityIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
