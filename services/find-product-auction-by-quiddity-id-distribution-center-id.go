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

var findProductAuctionByQuiddityIDDistributionCenterIDService = achaemenid.Service{
	ID:                1699303145,
	IssueDate:         1605202788,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Product Auction By Quiddity ID Distribution Center ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByQuiddityIDDistributionCenterIDSRPC,
	HTTPHandler: FindProductAuctionByQuiddityIDDistributionCenterIDHTTP,
}

// FindProductAuctionByQuiddityIDDistributionCenterIDSRPC is sRPC handler of FindProductAuctionByQuiddityIDDistributionCenterID service.
func FindProductAuctionByQuiddityIDDistributionCenterIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByQuiddityIDDistributionCenterIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByQuiddityIDDistributionCenterIDRes
	res, st.Err = findProductAuctionByQuiddityIDDistributionCenterID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByQuiddityIDDistributionCenterIDHTTP is HTTP handler of FindProductAuctionByQuiddityIDDistributionCenterID service.
func FindProductAuctionByQuiddityIDDistributionCenterIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByQuiddityIDDistributionCenterIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByQuiddityIDDistributionCenterIDRes
	res, st.Err = findProductAuctionByQuiddityIDDistributionCenterID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByQuiddityIDDistributionCenterIDReq struct {
	QuiddityID           [32]byte `json:",string"`
	DistributionCenterID [32]byte `json:",string"`
	Offset               uint64
	Limit                uint64
}

type findProductAuctionByQuiddityIDDistributionCenterIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByQuiddityIDDistributionCenterID(st *achaemenid.Stream, req *findProductAuctionByQuiddityIDDistributionCenterIDReq) (res *findProductAuctionByQuiddityIDDistributionCenterIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		QuiddityID: req.QuiddityID,
		Authorization: authorization.Product{
			AllowUserID: req.DistributionCenterID,
		},
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByQuiddityIDAllowUserID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByQuiddityIDDistributionCenterIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])
	copy(req.DistributionCenterID[:], buf[32:])
	req.Offset = syllab.GetUInt64(buf, 64)
	req.Limit = syllab.GetUInt64(buf, 72)
	return
}

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])
	copy(buf[32:], req.DistributionCenterID[:])
	syllab.SetUInt64(buf, 64, req.Offset)
	syllab.SetUInt64(buf, 72, req.Limit)
	return
}

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) syllabStackLen() (ln uint32) {
	return 80
}

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])
		case "DistributionCenterID":
			err = decoder.DecodeByteArrayAsBase64(req.DistributionCenterID[:])
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

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])

	encoder.EncodeString(`","DistributionCenterID":"`)
	encoder.EncodeByteSliceAsBase64(req.DistributionCenterID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByQuiddityIDDistributionCenterIDReq) jsonLen() (ln int) {
	ln = 184
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByQuiddityIDDistributionCenterIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
