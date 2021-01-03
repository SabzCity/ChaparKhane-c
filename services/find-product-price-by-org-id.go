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

var findProductPriceByOrgIDService = achaemenid.Service{
	ID:                3010862396,
	IssueDate:         1608092848,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Product Price By Org ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductPrice",
	},

	SRPCHandler: FindProductPriceByOrgIDSRPC,
	HTTPHandler: FindProductPriceByOrgIDHTTP,
}

// FindProductPriceByOrgIDSRPC is sRPC handler of FindProductPriceByOrgID service.
func FindProductPriceByOrgIDSRPC(st *achaemenid.Stream) {
	var req = &findProductPriceByOrgIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findProductPriceByOrgIDRes
	res, st.Err = findProductPriceByOrgID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindProductPriceByOrgIDHTTP is HTTP handler of FindProductPriceByOrgID service.
func FindProductPriceByOrgIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findProductPriceByOrgIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findProductPriceByOrgIDRes
	res, st.Err = findProductPriceByOrgID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findProductPriceByOrgIDReq struct {
	OrgID  [32]byte `json:",string"`
	Offset uint64
	Limit  uint64
}

type findProductPriceByOrgIDRes struct {
	QuiddityIDs [][32]byte `json:",string"`
}

func findProductPriceByOrgID(st *achaemenid.Stream, req *findProductPriceByOrgIDReq) (res *findProductPriceByOrgIDRes, err *er.Error) {
	var pp = datastore.ProductPrice{
		OrgID: req.OrgID,
	}
	var indexRes [][32]byte
	indexRes, err = pp.FindQuiddityIDsByOrgID(req.Offset, req.Limit)
	if err != nil {
		return
	}

	res = &findProductPriceByOrgIDRes{
		QuiddityIDs: indexRes,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findProductPriceByOrgIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.OrgID[:], buf[0:])
	req.Offset = syllab.GetUInt64(buf, 32)
	req.Limit = syllab.GetUInt64(buf, 40)
	return
}

func (req *findProductPriceByOrgIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.OrgID[:])
	syllab.SetUInt64(buf, 32, req.Offset)
	syllab.SetUInt64(buf, 40, req.Limit)
	return
}

func (req *findProductPriceByOrgIDReq) syllabStackLen() (ln uint32) {
	return 48
}

func (req *findProductPriceByOrgIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *findProductPriceByOrgIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findProductPriceByOrgIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "OrgID":
			err = decoder.DecodeByteArrayAsBase64(req.OrgID[:])
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

func (req *findProductPriceByOrgIDReq) jsonEncoder() (buf []byte) {
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

func (req *findProductPriceByOrgIDReq) jsonLen() (ln int) {
	ln = 114
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findProductPriceByOrgIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.QuiddityIDs = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	return
}

func (res *findProductPriceByOrgIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.Set32ByteArrayArray(buf, res.QuiddityIDs, 0, hsi)
	return
}

func (res *findProductPriceByOrgIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *findProductPriceByOrgIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.QuiddityIDs) * 32)
	return
}

func (res *findProductPriceByOrgIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findProductPriceByOrgIDRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityIDs":
			res.QuiddityIDs, err = decoder.Decode32ByteArraySliceAsBase64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *findProductPriceByOrgIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityIDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.QuiddityIDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *findProductPriceByOrgIDRes) jsonLen() (ln int) {
	ln = len(res.QuiddityIDs) * 46
	ln += 14
	return
}
