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

var findProductAuctionByWikiIDDistributionCenterIDService = achaemenid.Service{
	ID:                2532138649,
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
		lang.EnglishLanguage: "Find Product Auction By Wiki ID Distribution Center ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: FindProductAuctionByWikiIDDistributionCenterIDSRPC,
	HTTPHandler: FindProductAuctionByWikiIDDistributionCenterIDHTTP,
}

// FindProductAuctionByWikiIDDistributionCenterIDSRPC is sRPC handler of FindProductAuctionByWikiIDDistributionCenterID service.
func FindProductAuctionByWikiIDDistributionCenterIDSRPC(st *achaemenid.Stream) {
	var req = &findProductAuctionByWikiIDDistributionCenterIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductAuctionByWikiIDDistributionCenterIDRes
	res, st.Err = findProductAuctionByWikiIDDistributionCenterID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductAuctionByWikiIDDistributionCenterIDHTTP is HTTP handler of FindProductAuctionByWikiIDDistributionCenterID service.
func FindProductAuctionByWikiIDDistributionCenterIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductAuctionByWikiIDDistributionCenterIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductAuctionByWikiIDDistributionCenterIDRes
	res, st.Err = findProductAuctionByWikiIDDistributionCenterID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductAuctionByWikiIDDistributionCenterIDReq struct {
	WikiID               [32]byte `json:",string"`
	DistributionCenterID [32]byte `json:",string"`
	Offset               uint64
	Limit                uint64
}

type findProductAuctionByWikiIDDistributionCenterIDRes struct {
	IDs [][32]byte `json:",string"`
}

func findProductAuctionByWikiIDDistributionCenterID(st *achaemenid.Stream, req *findProductAuctionByWikiIDDistributionCenterIDReq) (res *findProductAuctionByWikiIDDistributionCenterIDRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		WikiID:               req.WikiID,
		DistributionCenterID: req.DistributionCenterID,
	}
	var indexRes [][32]byte
	indexRes, err = pa.FindIDsByWikiIDDistributionCenterIDByHashIndex(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductAuctionByWikiIDDistributionCenterIDRes{
		IDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.WikiID[:], buf[0:])
	copy(req.DistributionCenterID[:], buf[32:])
	req.Offset = syllab.GetUInt64(buf, 64)
	req.Limit = syllab.GetUInt64(buf, 72)
	return
}

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.WikiID[:])
	copy(buf[32:], req.DistributionCenterID[:])
	syllab.SetUInt64(buf, 64, req.Offset)
	syllab.SetUInt64(buf, 72, req.Limit)
	return
}

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) syllabStackLen() (ln uint32) {
	return 80
}

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		case 'D':
			decoder.SetFounded()
			decoder.Offset(23)
			err = decoder.DecodeByteArrayAsBase64(req.DistributionCenterID[:])
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

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"WikiID":"`)
	encoder.EncodeByteSliceAsBase64(req.WikiID[:])

	encoder.EncodeString(`","DistributionCenterID":"`)
	encoder.EncodeByteSliceAsBase64(req.DistributionCenterID[:])

	encoder.EncodeString(`","Offset":`)
	encoder.EncodeUInt64(req.Offset)

	encoder.EncodeString(`,"Limit":`)
	encoder.EncodeUInt64(req.Limit)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findProductAuctionByWikiIDDistributionCenterIDReq) jsonLen() (ln int) {
	ln = 184
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductAuctionByWikiIDDistributionCenterIDRes) jsonLen() (ln int) {
	ln = len(res.IDs) * 46
	ln += 8
	return
}
