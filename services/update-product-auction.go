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
	"../libgo/validators"
)

var updateProductAuctionService = achaemenid.Service{
	ID:                3439443464,
	IssueDate:         1605250043,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Update Product Auction",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: UpdateProductAuctionSRPC,
	HTTPHandler: UpdateProductAuctionHTTP,
}

// UpdateProductAuctionSRPC is sRPC handler of UpdateProductAuction service.
func UpdateProductAuctionSRPC(st *achaemenid.Stream) {
	var req = &updateProductAuctionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateProductAuction(st, req)
	st.OutcomePayload = make([]byte, 4)
}

// UpdateProductAuctionHTTP is HTTP handler of UpdateProductAuction service.
func UpdateProductAuctionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateProductAuctionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateProductAuctionReq struct {
	ID [32]byte `json:",string"`

	// Price
	SuggestPrice                 price.Amount
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad
	Discount                     math.PerMyriad

	// Authorization
	MinNumBuy     uint64
	StockNumber   uint64
	LiveUntil     etime.Time
	AllowWeekdays etime.Weekdays
	AllowDayhours etime.Dayhours

	Description string `valid:"text[0:50]"`
	Type        datastore.ProductAuctionType
}

func updateProductAuction(st *achaemenid.Stream, req *updateProductAuctionReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var pa = datastore.ProductAuction{
		ID: req.ID,
	}
	err = pa.GetLastByIDByHashIndex()
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = ErrProductAuctionNotRegistered
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
		OrgID:            pa.OrgID,
		ID:               pa.ID,
		WikiID:           pa.WikiID,

		Currency:                     pa.Currency,
		SuggestPrice:                 req.SuggestPrice,
		DistributionCenterCommission: req.DistributionCenterCommission,
		SellerCommission:             req.SellerCommission,
		Discount:                     req.Discount,

		// Authorization
		DistributionCenterID: pa.DistributionCenterID,
		GroupID:              pa.GroupID,
		MinNumBuy:            req.MinNumBuy,
		StockNumber:          req.StockNumber,
		LiveUntil:            req.LiveUntil,
		AllowWeekdays:        req.AllowWeekdays,
		AllowDayhours:        req.AllowDayhours,

		Description: req.Description,
		Type:        req.Type,
		Status:      datastore.ProductAuctionUpdated,
	}

	if req.LiveUntil != 0 && req.LiveUntil.Pass(etime.Now()) {
		pa.Status = datastore.ProductAuctionExpired
	}

	err = pa.Set()
	if err != nil {
		return
	}
	pa.HashIndexRecordIDForID()
	return
}

func (req *updateProductAuctionReq) validator() (err *er.Error) {
	err = validators.ValidateText(req.Description, 0, 50)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateProductAuctionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.SuggestPrice = price.Amount(syllab.GetUInt64(buf, 32))
	req.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 40))
	req.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 42))
	req.Discount = math.PerMyriad(syllab.GetUInt16(buf, 44))
	req.MinNumBuy = syllab.GetUInt64(buf, 46)
	req.StockNumber = syllab.GetUInt64(buf, 54)
	req.LiveUntil = etime.Time(syllab.GetInt64(buf, 62))
	req.AllowWeekdays = etime.Weekdays(syllab.GetUInt8(buf, 70))
	req.AllowDayhours = etime.Dayhours(syllab.GetUInt32(buf, 71))
	req.Description = syllab.UnsafeGetString(buf, 75)
	req.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 83))
	return
}

func (req *updateProductAuctionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])
	syllab.SetUInt64(buf, 32, uint64(req.SuggestPrice))
	syllab.SetUInt16(buf, 40, uint16(req.DistributionCenterCommission))
	syllab.SetUInt16(buf, 42, uint16(req.SellerCommission))
	syllab.SetUInt16(buf, 44, uint16(req.Discount))
	syllab.SetUInt64(buf, 46, req.MinNumBuy)
	syllab.SetUInt64(buf, 54, req.StockNumber)
	syllab.SetInt64(buf, 62, int64(req.LiveUntil))
	syllab.SetUInt8(buf, 70, uint8(req.AllowWeekdays))
	syllab.SetUInt32(buf, 71, uint32(req.AllowDayhours))
	syllab.SetString(buf, req.Description, 75, hsi)
	syllab.SetUInt8(buf, 83, uint8(req.Type))
	return
}

func (req *updateProductAuctionReq) syllabStackLen() (ln uint32) {
	return 84
}

func (req *updateProductAuctionReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Description))
	return
}

func (req *updateProductAuctionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateProductAuctionReq) jsonDecoder(buf []byte) (err *er.Error) {
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

func (req *updateProductAuctionReq) jsonEncoder() (buf []byte) {
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

	encoder.EncodeString(`,"MinNumBuy":`)
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
	encoder.EncodeUInt64(uint64(req.Type))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *updateProductAuctionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += 438
	return
}
