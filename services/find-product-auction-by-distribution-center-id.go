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

var findProductAuctionByDistributionCenterIDService = achaemenid.Service{
	ID:                1692737536,
	IssueDate:         1605202730,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Product Auction By Distribution Center ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByDistributionCenterIDSRPC,
	HTTPHandler: FindProductAuctionByDistributionCenterIDHTTP,
}

// FindProductAuctionByDistributionCenterIDSRPC is sRPC handler of FindProductAuctionByDistributionCenterID service.
func FindProductAuctionByDistributionCenterIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByDistributionCenterIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByDistributionCenterIDRes
	res, st.Err = findProductAuctionByDistributionCenterID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByDistributionCenterIDHTTP is HTTP handler of FindProductAuctionByDistributionCenterID service.
func FindProductAuctionByDistributionCenterIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByDistributionCenterIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByDistributionCenterIDRes
	res, st.Err = findProductAuctionByDistributionCenterID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByDistributionCenterIDReq struct {
	DistributionCenterID [32]byte `json:",string"`
	Offset               uint64
	Limit                uint64
}

type findProductAuctionByDistributionCenterIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByDistributionCenterID(st *achaemenid.Stream, req *findProductAuctionByDistributionCenterIDReq) (res *findProductAuctionByDistributionCenterIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		Authorization: authorization.Product{
			AllowUserID: req.DistributionCenterID,
		},
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByAllowUserID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByDistributionCenterIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByDistributionCenterIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.DistributionCenterID[:], buf[0:])
	req.Offset = syllab.GetUInt64(buf, 32)
	req.Limit = syllab.GetUInt64(buf, 40)
	return
}

func (req *findProductAuctionByDistributionCenterIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.DistributionCenterID[:])
	syllab.SetUInt64(buf, 32, req.Offset)
	syllab.SetUInt64(buf, 40, req.Limit)
	return
}

func (req *findProductAuctionByDistributionCenterIDReq) syllabStackLen() (ln uint32) {
	return 48
}

func (req *findProductAuctionByDistributionCenterIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByDistributionCenterIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByDistributionCenterIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
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

func (req *findProductAuctionByDistributionCenterIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"DistributionCenterID":"`)
	encoder.EncodeByteSliceAsBase64(req.DistributionCenterID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByDistributionCenterIDReq) jsonLen() (ln int) {
	ln = 129
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByDistributionCenterIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByDistributionCenterIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByDistributionCenterIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByDistributionCenterIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByDistributionCenterIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByDistributionCenterIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByDistributionCenterIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByDistributionCenterIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
