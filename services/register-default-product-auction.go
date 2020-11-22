/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/ganjine"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/math"
	"../libgo/price"
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/uuid"
	"../libgo/validators"
)

var registerDefaultProductAuctionService = achaemenid.Service{
	ID:                152569757,
	IssueDate:         1605189667,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Register Default Product Auction",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: RegisterDefaultProductAuctionSRPC,
	HTTPHandler: RegisterDefaultProductAuctionHTTP,
}

// RegisterDefaultProductAuctionSRPC is sRPC handler of RegisterDefaultProductAuction service.
func RegisterDefaultProductAuctionSRPC(st *achaemenid.Stream) {
	var req = &registerDefaultProductAuctionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerDefaultProductAuctionRes
	res, st.Err = registerDefaultProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterDefaultProductAuctionHTTP is HTTP handler of RegisterDefaultProductAuction service.
func RegisterDefaultProductAuctionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerDefaultProductAuctionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerDefaultProductAuctionRes
	res, st.Err = registerDefaultProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerDefaultProductAuctionReq struct {
	WikiID   [32]byte      `json:",string"`
	Language lang.Language // Just use to check wiki exist and belong to requested org

	Currency                     price.Currency
	SuggestPrice                 price.Amount
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad
	Discount                     math.PerMyriad

	Description string `valid:"text[0:50]"`
	Type        datastore.ProductAuctionType
}

type registerDefaultProductAuctionRes struct {
	ID [32]byte `json:",string"`
}

func registerDefaultProductAuction(st *achaemenid.Stream, req *registerDefaultProductAuctionReq) (res *registerDefaultProductAuctionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	// Check wiki exits and belong to this Org
	var getWikiByIDReq = getWikiByIDReq{
		ID:       req.WikiID,
		Language: req.Language,
	}
	var getWikiByIDRes *getWikiByIDRes
	getWikiByIDRes, err = getWikiByID(st, &getWikiByIDReq)
	if err != nil {
		return
	}
	if getWikiByIDRes.OrgID != st.Connection.UserID {
		err = authorization.ErrUserNotAllow
		return
	}
	if getWikiByIDRes.Status == datastore.WikiStatusBlocked {
		err = ErrBlockedByJustice
		return
	}

	var pa = datastore.ProductAuction{
		WikiID:   req.WikiID,
		Currency: req.Currency,
	}
	var IDs [][32]byte
	IDs, err = pa.FindIDsByWikiIDCurrencyByHashIndex(0, 1)
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = nil
	}
	if err != nil {
		return
	}
	if len(IDs) == 1 {
		err = ErrProductAuctionRegistered
		return
	}

	pa = datastore.ProductAuction{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            st.Connection.UserID,
		ID:               uuid.Random32Byte(),
		WikiID:           req.WikiID,

		Currency:                     req.Currency,
		SuggestPrice:                 req.SuggestPrice,
		DistributionCenterCommission: req.DistributionCenterCommission,
		SellerCommission:             req.SellerCommission,
		Discount:                     req.Discount,

		AllowWeekdays: etime.WeekdaysAll,
		AllowDayhours: etime.DayhoursAll,

		Description: req.Description,
		Type:        req.Type,
		Status:      datastore.ProductAuctionRegistered,
	}
	err = pa.SaveNew()
	if err != nil {
		return
	}

	res = &registerDefaultProductAuctionRes{
		ID: pa.ID,
	}

	return
}

func (req *registerDefaultProductAuctionReq) validator() (err *er.Error) {
	err = validators.ValidateText(req.Description, 0, 50)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerDefaultProductAuctionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.WikiID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))

	req.Currency = price.Currency(syllab.GetUInt16(buf, 36))
	req.SuggestPrice = price.Amount(syllab.GetUInt64(buf, 38))
	req.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 46))
	req.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 48))
	req.Discount = math.PerMyriad(syllab.GetUInt16(buf, 50))

	req.Description = syllab.UnsafeGetString(buf, 52)
	req.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 60))
	return
}

func (req *registerDefaultProductAuctionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.WikiID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))

	syllab.SetUInt16(buf, 36, uint16(req.Currency))
	syllab.SetUInt64(buf, 38, uint64(req.SuggestPrice))
	syllab.SetUInt16(buf, 46, uint16(req.DistributionCenterCommission))
	syllab.SetUInt16(buf, 48, uint16(req.SellerCommission))
	syllab.SetUInt16(buf, 50, uint16(req.Discount))

	syllab.SetString(buf, req.Description, 52, hsi)
	syllab.SetUInt8(buf, 60, uint8(req.Type))
	return
}

func (req *registerDefaultProductAuctionReq) syllabStackLen() (ln uint32) {
	return 61
}

func (req *registerDefaultProductAuctionReq) syllabHeapLen() (ln uint32) {
	ln = uint32(len(req.Description))
	return
}

func (req *registerDefaultProductAuctionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerDefaultProductAuctionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	var keyName string
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		keyName = decoder.DecodeKey()
		switch keyName {
		case "WikiID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(req.WikiID[:])
			if err != nil {
				return
			}
		case "Language":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			req.Language = lang.Language(num)
		case "Currency":
			decoder.SetFounded()
			var num uint16
			num, err = decoder.DecodeUInt16()
			if err != nil {
				return
			}
			req.Currency = price.Currency(num)
		case "SuggestPrice":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.SuggestPrice = price.Amount(num)
		case "DistributionCenterCommission":
			decoder.SetFounded()
			var num uint16
			num, err = decoder.DecodeUInt16()
			if err != nil {
				return
			}
			req.DistributionCenterCommission = math.PerMyriad(num)
		case "SellerCommission":
			decoder.SetFounded()
			var num uint16
			num, err = decoder.DecodeUInt16()
			if err != nil {
				return
			}
			req.SellerCommission = math.PerMyriad(num)
		case "Discount":
			decoder.SetFounded()
			var num uint16
			num, err = decoder.DecodeUInt16()
			if err != nil {
				return
			}
			req.Discount = math.PerMyriad(num)
		case "Description":
			decoder.SetFounded()
			decoder.Offset(1)
			req.Description = decoder.DecodeString()
		case "Type":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			req.Type = datastore.ProductAuctionType(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *registerDefaultProductAuctionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"WikiID":"`)
	encoder.EncodeByteSliceAsBase64(req.WikiID[:])

	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt64(uint64(req.Language))

	encoder.EncodeString(`,"Currency":`)
	encoder.EncodeUInt64(uint64(req.Currency))

	encoder.EncodeString(`,"SuggestPrice":`)
	encoder.EncodeUInt64(uint64(req.SuggestPrice))

	encoder.EncodeString(`,"DistributionCenterCommission":`)
	encoder.EncodeUInt64(uint64(req.DistributionCenterCommission))

	encoder.EncodeString(`,"SellerCommission":`)
	encoder.EncodeUInt64(uint64(req.SellerCommission))

	encoder.EncodeString(`,"Discount":`)
	encoder.EncodeUInt64(uint64(req.Discount))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(req.Description)

	encoder.EncodeString(`","Type":`)
	encoder.EncodeUInt8(uint8(req.Type))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerDefaultProductAuctionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += 334
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerDefaultProductAuctionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerDefaultProductAuctionRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerDefaultProductAuctionRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerDefaultProductAuctionRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerDefaultProductAuctionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerDefaultProductAuctionRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
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

func (res *registerDefaultProductAuctionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerDefaultProductAuctionRes) jsonLen() (ln int) {
	ln = 52
	return
}
