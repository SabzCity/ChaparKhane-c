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

var findProductAuctionByQuiddityIDGroupIDService = achaemenid.Service{
	ID:                79942143,
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
		lang.LanguageEnglish: "Find Product Auction By Quiddity ID Group ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByQuiddityIDGroupIDSRPC,
	HTTPHandler: FindProductAuctionByQuiddityIDGroupIDHTTP,
}

// FindProductAuctionByQuiddityIDGroupIDSRPC is sRPC handler of FindProductAuctionByQuiddityIDGroupID service.
func FindProductAuctionByQuiddityIDGroupIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByQuiddityIDGroupIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByQuiddityIDGroupIDRes
	res, st.Err = findProductAuctionByQuiddityIDGroupID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByQuiddityIDGroupIDHTTP is HTTP handler of FindProductAuctionByQuiddityIDGroupID service.
func FindProductAuctionByQuiddityIDGroupIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByQuiddityIDGroupIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByQuiddityIDGroupIDRes
	res, st.Err = findProductAuctionByQuiddityIDGroupID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByQuiddityIDGroupIDReq struct {
	QuiddityID [32]byte `json:",string"`
	GroupID    [32]byte `json:",string"`
	Offset     uint64
	Limit      uint64
}

type findProductAuctionByQuiddityIDGroupIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByQuiddityIDGroupID(st *achaemenid.Stream, req *findProductAuctionByQuiddityIDGroupIDReq) (res *findProductAuctionByQuiddityIDGroupIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		QuiddityID: req.QuiddityID,
		Authorization: authorization.Product{
			GroupID: req.GroupID,
		},
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByQuiddityIDGroupID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByQuiddityIDGroupIDRes{
		IDs: indexRes,
	}
	return
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByQuiddityIDGroupIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])
	copy(req.GroupID[:], buf[32:])
	req.Offset = syllab.GetUInt64(buf, 64)
	req.Limit = syllab.GetUInt64(buf, 72)
	return
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])
	copy(buf[32:], req.GroupID[:])
	syllab.SetUInt64(buf, 64, req.Offset)
	syllab.SetUInt64(buf, 72, req.Limit)
	return
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) syllabStackLen() (ln uint32) {
	return 80
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])
		case "GroupID":
			err = decoder.DecodeByteArrayAsBase64(req.GroupID[:])
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

func (req *findProductAuctionByQuiddityIDGroupIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])

	encoder.EncodeString(`","GroupID":"`)
	encoder.EncodeByteSliceAsBase64(req.GroupID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByQuiddityIDGroupIDReq) jsonLen() (ln int) {
	ln = 171
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByQuiddityIDGroupIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByQuiddityIDGroupIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByQuiddityIDGroupIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByQuiddityIDGroupIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByQuiddityIDGroupIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByQuiddityIDGroupIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByQuiddityIDGroupIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByQuiddityIDGroupIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
