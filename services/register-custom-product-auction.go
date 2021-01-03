/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/ganjine"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/math"
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/uuid"
	"../libgo/validators"
)

var registerCustomProductAuctionService = achaemenid.Service{
	ID:                2675795419,
	IssueDate:         1605189695,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Custom Product Auction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: RegisterCustomProductAuctionSRPC,
	HTTPHandler: RegisterCustomProductAuctionHTTP,
}

// RegisterCustomProductAuctionSRPC is sRPC handler of RegisterCustomProductAuction service.
func RegisterCustomProductAuctionSRPC(st *achaemenid.Stream) {
	var req = &registerCustomProductAuctionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerCustomProductAuctionRes
	res, st.Err = registerCustomProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterCustomProductAuctionHTTP is HTTP handler of RegisterCustomProductAuction service.
func RegisterCustomProductAuctionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerCustomProductAuctionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerCustomProductAuctionRes
	res, st.Err = registerCustomProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerCustomProductAuctionReq struct {
	ID [32]byte `json:",string"`

	// Price
	Discount                     math.PerMyriad
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad

	// Authorization
	Authorization authorization.Product

	Description string `valid:"text[0:50]"`
	Type        datastore.ProductAuctionType
}

type registerCustomProductAuctionRes struct {
	ID [32]byte `json:",string"`
}

func registerCustomProductAuction(st *achaemenid.Stream, req *registerCustomProductAuctionReq) (res *registerCustomProductAuctionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	var pa = datastore.ProductAuction{
		ID: req.ID,
	}
	err = pa.GetLastByID()
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = ErrProductAuctionDefaultNotRegistered
		return
	}
	if err != nil {
		return
	}
	if pa.OrgID != st.Connection.UserID {
		err = authorization.ErrUserNotAllow
		return
	}
	if pa.Status == datastore.QuiddityStatusBlocked {
		err = ErrBlockedByJustice
		return
	}

	pa = datastore.ProductAuction{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            st.Connection.UserID,
		ID:               uuid.Random32Byte(),
		QuiddityID:           pa.QuiddityID,

		Discount:         req.Discount,
		DCCommission:     req.DistributionCenterCommission,
		SellerCommission: req.SellerCommission,

		// Authorization
		Authorization: req.Authorization,

		Description: req.Description,
		Type:        req.Type,
		Status:      datastore.ProductAuctionRegistered,
	}
	err = pa.SaveNew()
	if err != nil {
		return
	}

	res = &registerCustomProductAuctionRes{
		ID: pa.ID,
	}
	return
}

func (req *registerCustomProductAuctionReq) validator() (err *er.Error) {
	err = validators.ValidateText(req.Description, 0, 50)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerCustomProductAuctionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])

	req.Discount = math.PerMyriad(syllab.GetUInt16(buf, 32))
	req.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 34))
	req.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 36))

	req.Authorization.SyllabDecoder(buf, 38)

	req.Description = syllab.UnsafeGetString(buf, 38+req.Authorization.SyllabStackLen())
	req.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 46+req.Authorization.SyllabStackLen()))
	return
}

func (req *registerCustomProductAuctionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])

	syllab.SetUInt16(buf, 32, uint16(req.Discount))
	syllab.SetUInt16(buf, 34, uint16(req.DistributionCenterCommission))
	syllab.SetUInt16(buf, 36, uint16(req.SellerCommission))

	hsi = req.Authorization.SyllabEncoder(buf, 38, hsi)

	hsi = syllab.SetString(buf, req.Description, 38+req.Authorization.SyllabStackLen(), hsi)
	syllab.SetUInt8(buf, 46+req.Authorization.SyllabStackLen(), uint8(req.Type))
	return
}

func (req *registerCustomProductAuctionReq) syllabStackLen() (ln uint32) {
	return 47 + req.Authorization.SyllabStackLen()
}

func (req *registerCustomProductAuctionReq) syllabHeapLen() (ln uint32) {
	ln += req.Authorization.SyllabHeapLen()
	ln += uint32(len(req.Description))
	return
}

func (req *registerCustomProductAuctionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerCustomProductAuctionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])

		case "Discount":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.Discount = math.PerMyriad(num)
		case "DistributionCenterCommission":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.DistributionCenterCommission = math.PerMyriad(num)
		case "SellerCommission":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.SellerCommission = math.PerMyriad(num)

		case "Authorization":
			err = req.Authorization.JSONDecoder(decoder)

		case "Description":
			req.Description, err = decoder.DecodeString()
		case "Type":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.Type = datastore.ProductAuctionType(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *registerCustomProductAuctionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Discount":`)
	encoder.EncodeUInt16(uint16(req.Discount))
	encoder.EncodeString(`,"DistributionCenterCommission":`)
	encoder.EncodeUInt16(uint16(req.DistributionCenterCommission))
	encoder.EncodeString(`,"SellerCommission":`)
	encoder.EncodeUInt16(uint16(req.SellerCommission))

	encoder.EncodeString(`,"Authorization":`)
	req.Authorization.JSONEncoder(encoder)

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(req.Description)
	encoder.EncodeString(`","Type":`)
	encoder.EncodeUInt8(uint8(req.Type))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerCustomProductAuctionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += req.Authorization.JSONLen()
	ln += 179
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerCustomProductAuctionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerCustomProductAuctionRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerCustomProductAuctionRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerCustomProductAuctionRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerCustomProductAuctionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerCustomProductAuctionRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *registerCustomProductAuctionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerCustomProductAuctionRes) jsonLen() (ln int) {
	ln = 52
	return
}
