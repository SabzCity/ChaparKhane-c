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
		lang.EnglishLanguage: "Register Custom Product Auction",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
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
	SuggestPrice                 price.Amount
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad
	Discount                     math.PerMyriad

	// Authorization
	DistributionCenterID [32]byte `json:",string"`
	GroupID              [32]byte `json:",string"`
	MinNumBuy            uint64
	StockNumber          uint64
	LiveUntil            etime.Time
	AllowWeekdays        etime.Weekdays
	AllowDayhours        etime.Dayhours

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
	err = pa.GetLastByIDByHashIndex()
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
	if pa.Status == datastore.WikiStatusBlocked {
		err = ErrBlockedByJustice
		return
	}

	pa = datastore.ProductAuction{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            st.Connection.UserID,
		ID:               uuid.Random32Byte(),
		WikiID:           pa.WikiID,

		Currency:                     pa.Currency,
		SuggestPrice:                 req.SuggestPrice,
		DistributionCenterCommission: req.DistributionCenterCommission,
		SellerCommission:             req.SellerCommission,
		Discount:                     req.Discount,

		// Authorization
		DistributionCenterID: req.DistributionCenterID,
		GroupID:              req.GroupID,
		MinNumBuy:            req.MinNumBuy,
		StockNumber:          req.StockNumber,
		LiveUntil:            req.LiveUntil,
		AllowWeekdays:        req.AllowWeekdays,
		AllowDayhours:        req.AllowDayhours,

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

	req.SuggestPrice = price.Amount(syllab.GetUInt64(buf, 32))
	req.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 40))
	req.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 42))
	req.Discount = math.PerMyriad(syllab.GetUInt16(buf, 44))

	copy(req.DistributionCenterID[:], buf[46:])
	copy(req.GroupID[:], buf[78:])
	req.MinNumBuy = syllab.GetUInt64(buf, 110)
	req.StockNumber = syllab.GetUInt64(buf, 118)
	req.LiveUntil = etime.Time(syllab.GetInt64(buf, 126))
	req.AllowWeekdays = etime.Weekdays(syllab.GetUInt8(buf, 134))
	req.AllowDayhours = etime.Dayhours(syllab.GetUInt32(buf, 135))

	req.Description = syllab.UnsafeGetString(buf, 139)
	req.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 147))
	return
}

func (req *registerCustomProductAuctionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])

	syllab.SetUInt64(buf, 32, uint64(req.SuggestPrice))
	syllab.SetUInt16(buf, 40, uint16(req.DistributionCenterCommission))
	syllab.SetUInt16(buf, 42, uint16(req.SellerCommission))
	syllab.SetUInt16(buf, 44, uint16(req.Discount))

	copy(buf[46:], req.DistributionCenterID[:])
	copy(buf[78:], req.GroupID[:])
	syllab.SetUInt64(buf, 110, req.MinNumBuy)
	syllab.SetUInt64(buf, 118, req.StockNumber)
	syllab.SetInt64(buf, 126, int64(req.LiveUntil))
	syllab.SetUInt8(buf, 134, uint8(req.AllowWeekdays))
	syllab.SetUInt32(buf, 135, uint32(req.AllowDayhours))

	hsi = syllab.SetString(buf, req.Description, 139, hsi)
	syllab.SetUInt8(buf, 147, uint8(req.Type))
	return
}

func (req *registerCustomProductAuctionReq) syllabStackLen() (ln uint32) {
	return 148
}

func (req *registerCustomProductAuctionReq) syllabHeapLen() (ln uint32) {
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
	var keyName string
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
			if err != nil {
				return
			}
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
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.DistributionCenterCommission = math.PerMyriad(num)
		case "SellerCommission":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.SellerCommission = math.PerMyriad(num)
		case "Discount":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Discount = math.PerMyriad(num)
		case "DistributionCenterID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(req.DistributionCenterID[:])
			if err != nil {
				return
			}
		case "GroupID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(req.GroupID[:])
			if err != nil {
				return
			}
		case "MinNumBuy":
			decoder.SetFounded()
			req.MinNumBuy, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
		case "StockNumber":
			decoder.SetFounded()
			req.StockNumber, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
		case "LiveUntil":
			decoder.SetFounded()
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			req.LiveUntil = etime.Time(num)
		case "AllowWeekdays":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.AllowWeekdays = etime.Weekdays(num)
		case "AllowDayhours":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.AllowDayhours = etime.Dayhours(num)
		case "Description":
			decoder.SetFounded()
			decoder.Offset(1)
			req.Description = decoder.DecodeString()
		case "Type":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
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

func (req *registerCustomProductAuctionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","SuggestPrice":`)
	encoder.EncodeUInt64(uint64(req.SuggestPrice))

	encoder.EncodeString(`,"DistributionCenterCommission":`)
	encoder.EncodeUInt64(uint64(req.DistributionCenterCommission))

	encoder.EncodeString(`,"SellerCommission":`)
	encoder.EncodeUInt64(uint64(req.SellerCommission))

	encoder.EncodeString(`,"Discount":`)
	encoder.EncodeUInt64(uint64(req.Discount))

	encoder.EncodeString(`,"DistributionCenterID":"`)
	encoder.EncodeByteSliceAsBase64(req.DistributionCenterID[:])

	encoder.EncodeString(`","GroupID":"`)
	encoder.EncodeByteSliceAsBase64(req.GroupID[:])

	encoder.EncodeString(`","MinNumBuy":`)
	encoder.EncodeUInt64(req.MinNumBuy)

	encoder.EncodeString(`,"StockNumber":`)
	encoder.EncodeUInt64(req.StockNumber)

	encoder.EncodeString(`,"LiveUntil":`)
	encoder.EncodeInt64(int64(req.LiveUntil))

	encoder.EncodeString(`,"AllowWeekdays":`)
	encoder.EncodeUInt8(uint8(req.AllowWeekdays))

	encoder.EncodeString(`,"AllowDayhours":`)
	encoder.EncodeUInt64(uint64(req.AllowDayhours))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(req.Description)

	encoder.EncodeString(`","Type":`)
	encoder.EncodeUInt8(uint8(req.Type))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerCustomProductAuctionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += 563
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
	var keyName string
	for len(decoder.Buf) > 2 {
		keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			decoder.SetFounded()
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
