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
	"../libgo/uuid"
)

var registerProductService = achaemenid.Service{
	ID:                1054113390,
	URI:               "", // API services can set like "/apis?1054113390" but it is not efficient, find services by ID.
	IssueDate:         1608124881,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Product",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Product",
	},

	SRPCHandler: RegisterProductSRPC,
	HTTPHandler: RegisterProductHTTP,
}

// RegisterProductSRPC is sRPC handler of RegisterProduct service.
func RegisterProductSRPC(st *achaemenid.Stream) {
	var req = &registerProductReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerProductRes
	res, st.Err = registerProduct(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterProductHTTP is HTTP handler of RegisterProduct service.
func RegisterProductHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerProductReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerProductRes
	res, st.Err = registerProduct(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerProductReq struct {
	QuiddityID   [32]byte      `json:",string"`
	Language     lang.Language // Just use to check quiddity exist and belong to requested org
	ProductionID [32]byte      `json:",string"`
	Number       uint64
}

type registerProductRes struct {
	IDs [][32]byte `json:",string"`
}

func registerProduct(st *achaemenid.Stream, req *registerProductReq) (res *registerProductRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	// Check quiddity exits and belong to this Org
	var getQuiddityReq = getQuiddityReq{
		ID:       req.QuiddityID,
		Language: req.Language,
	}
	var getQuiddityRes *getQuiddityRes
	getQuiddityRes, err = getQuiddity(st, &getQuiddityReq)
	if err != nil {
		return
	}
	if getQuiddityRes.OrgID != st.Connection.UserID {
		err = authorization.ErrUserNotAllow
		return
	}
	if getQuiddityRes.Status == datastore.QuiddityStatusBlocked {
		err = ErrBlockedByJustice
		return
	}

	res = &registerProductRes{
		IDs: make([][32]byte, req.Number),
	}

	var i uint64
	for i = 0; i < req.Number; i++ {
		var p = datastore.Product{
			AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
			UserConnectionID: st.Connection.ID,
			ID:               uuid.Random32Byte(),
			OwnerID:          st.Connection.UserID,

			// SellerID:         req.SellerID,
			QuiddityID:   req.QuiddityID,
			ProductionID: req.ProductionID,
			DCID:         st.Connection.UserID, // product must register first in product owner OrgID and update it later.
			// ProductAuctionID: req.ProductAuctionID,
			Status: datastore.ProductCreated,
		}

		res.IDs[i] = p.ID

		err = p.SaveNew()
		if err != nil {
			return
		}
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerProductReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))
	copy(req.ProductionID[:], buf[36:])
	req.Number = syllab.GetUInt64(buf, 68)
	return
}

func (req *registerProductReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))
	copy(buf[36:], req.ProductionID[:])
	syllab.SetUInt64(buf, 68, req.Number)
	return
}

func (req *registerProductReq) syllabStackLen() (ln uint32) {
	return 76
}

func (req *registerProductReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *registerProductReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerProductReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])
		case "Language":
			var num uint32
			num, err = decoder.DecodeUInt32()
			req.Language = lang.Language(num)
		case "ProductionID":
			err = decoder.DecodeByteArrayAsBase64(req.ProductionID[:])
		case "Number":
			req.Number, err = decoder.DecodeUInt64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *registerProductReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])

	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt32(uint32(req.Language))

	encoder.EncodeString(`,"ProductionID":"`)
	encoder.EncodeByteSliceAsBase64(req.ProductionID[:])

	encoder.EncodeString(`","Number":`)
	encoder.EncodeUInt64(req.Number)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerProductReq) jsonLen() (ln int) {
	ln = 169
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerProductRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 0)
	return
}

func (res *registerProductRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *registerProductRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *registerProductRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *registerProductRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerProductRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *registerProductRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *registerProductRes) jsonLen() (ln int) {
	ln += len(res.IDs) * 46
	ln += 10
	return
}
