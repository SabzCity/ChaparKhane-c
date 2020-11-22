/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/math"
	"../libgo/price"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getProductAuctionService = achaemenid.Service{
	ID:                879426688,
	IssueDate:         1605202870,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get Product Auction",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"ProductAuction",
	},

	SRPCHandler: GetProductAuctionSRPC,
	HTTPHandler: GetProductAuctionHTTP,
}

// GetProductAuctionSRPC is sRPC handler of GetProductAuction service.
func GetProductAuctionSRPC(st *achaemenid.Stream) {
	var req = &getProductAuctionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getProductAuctionRes
	res, st.Err = getProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetProductAuctionHTTP is HTTP handler of GetProductAuction service.
func GetProductAuctionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getProductAuctionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getProductAuctionRes
	res, st.Err = getProductAuction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getProductAuctionReq struct {
	ID [32]byte `json:",string"`
}

type getProductAuctionRes struct {
	WriteTime        etime.Time
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	OrgID            [32]byte `json:",string"`
	WikiID           [32]byte `json:",string"`

	// Price
	Currency                     price.Currency
	SuggestPrice                 price.Amount
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad
	Discount                     math.PerMyriad
	PayablePrice                 price.Amount

	// Authorization
	DistributionCenterID [32]byte `json:",string"`
	GroupID              [32]byte `json:",string"`
	MinNumBuy            uint64
	StockNumber          uint64
	LiveUntil            etime.Time
	AllowWeekdays        etime.Weekdays
	AllowDayhours        etime.Dayhours

	Description string
	Type        datastore.ProductAuctionType
	Status      datastore.ProductAuctionStatus
}

func getProductAuction(st *achaemenid.Stream, req *getProductAuctionReq) (res *getProductAuctionRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		ID: req.ID,
	}
	err = pa.GetLastByIDByHashIndex()
	if err != nil {
		return
	}

	res = &getProductAuctionRes{
		WriteTime: pa.WriteTime,

		AppInstanceID:    pa.AppInstanceID,
		UserConnectionID: pa.UserConnectionID,
		OrgID:            pa.OrgID,
		WikiID:           pa.WikiID,

		// Price
		Currency:                     pa.Currency,
		SuggestPrice:                 pa.SuggestPrice,
		DistributionCenterCommission: pa.DistributionCenterCommission,
		SellerCommission:             pa.SellerCommission,
		Discount:                     pa.Discount,
		PayablePrice:                 pa.PayablePrice,

		// Authorization
		DistributionCenterID: pa.DistributionCenterID,
		GroupID:              pa.GroupID,
		MinNumBuy:            pa.MinNumBuy,
		StockNumber:          pa.StockNumber,
		LiveUntil:            pa.LiveUntil,
		AllowWeekdays:        pa.AllowWeekdays,
		AllowDayhours:        pa.AllowDayhours,

		Description: pa.Description,
		Type:        pa.Type,
		Status:      pa.Status,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getProductAuctionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getProductAuctionReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	return
}

func (req *getProductAuctionReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getProductAuctionReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getProductAuctionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getProductAuctionReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getProductAuctionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getProductAuctionReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getProductAuctionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	copy(res.OrgID[:], buf[72:])
	copy(res.WikiID[:], buf[104:])

	res.Currency = price.Currency(syllab.GetUInt16(buf, 136))
	res.SuggestPrice = price.Amount(syllab.GetUInt64(buf, 138))
	res.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 146))
	res.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 148))
	res.Discount = math.PerMyriad(syllab.GetUInt16(buf, 150))
	res.PayablePrice = price.Amount(syllab.GetUInt64(buf, 152))

	copy(res.DistributionCenterID[:], buf[160:])
	copy(res.GroupID[:], buf[192:])
	res.MinNumBuy = syllab.GetUInt64(buf, 224)
	res.StockNumber = syllab.GetUInt64(buf, 232)
	res.LiveUntil = etime.Time(syllab.GetInt64(buf, 240))
	res.AllowWeekdays = etime.Weekdays(syllab.GetUInt8(buf, 248))
	res.AllowDayhours = etime.Dayhours(syllab.GetUInt32(buf, 249))

	res.Description = syllab.UnsafeGetString(buf, 253)
	res.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 261))
	res.Status = datastore.ProductAuctionStatus(syllab.GetUInt8(buf, 262))
	return
}

func (res *getProductAuctionRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	copy(buf[72:], res.OrgID[:])
	copy(buf[104:], res.WikiID[:])

	syllab.SetUInt16(buf, 136, uint16(res.Currency))
	syllab.SetUInt64(buf, 138, uint64(res.SuggestPrice))
	syllab.SetUInt16(buf, 146, uint16(res.DistributionCenterCommission))
	syllab.SetUInt16(buf, 148, uint16(res.SellerCommission))
	syllab.SetUInt16(buf, 150, uint16(res.Discount))
	syllab.SetUInt64(buf, 152, uint64(res.PayablePrice))

	copy(buf[160:], res.DistributionCenterID[:])
	copy(buf[192:], res.GroupID[:])
	syllab.SetUInt64(buf, 224, res.MinNumBuy)
	syllab.SetUInt64(buf, 232, res.StockNumber)
	syllab.SetInt64(buf, 240, int64(res.LiveUntil))
	syllab.SetUInt8(buf, 248, uint8(res.AllowWeekdays))
	syllab.SetUInt32(buf, 249, uint32(res.AllowDayhours))

	hsi = syllab.SetString(buf, res.Description, 253, hsi)
	syllab.SetUInt8(buf, 261, uint8(res.Type))
	syllab.SetUInt8(buf, 262, uint8(res.Status))
	return
}

func (res *getProductAuctionRes) syllabStackLen() (ln uint32) {
	return 263
}

func (res *getProductAuctionRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Description))
	return
}

func (res *getProductAuctionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getProductAuctionRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	var keyName string
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		keyName = decoder.DecodeKey()
		switch keyName {
		case "WriteTime":
			decoder.SetFounded()
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			res.WriteTime = etime.Time(num)
		case "AppInstanceID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
			if err != nil {
				return
			}
		case "UserConnectionID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
			if err != nil {
				return
			}
		case "OrgID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(res.OrgID[:])
			if err != nil {
				return
			}
		case "WikiID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(res.WikiID[:])
			if err != nil {
				return
			}
		case "Currency":
			decoder.SetFounded()
			var num uint16
			num, err = decoder.DecodeUInt16()
			if err != nil {
				return
			}
			res.Currency = price.Currency(num)
		case "SuggestPrice":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			res.SuggestPrice = price.Amount(num)
			if err != nil {
				return
			}
		case "DistributionCenterCommission":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.DistributionCenterCommission = math.PerMyriad(num)
		case "SellerCommission":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.SellerCommission = math.PerMyriad(num)
		case "Discount":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.Discount = math.PerMyriad(num)
		case "PayablePrice":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			res.PayablePrice = price.Amount(num)
			if err != nil {
				return
			}
		case "DistributionCenterID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(res.DistributionCenterID[:])
			if err != nil {
				return
			}
		case "GroupID":
			decoder.SetFounded()
			decoder.Offset(1)
			err = decoder.DecodeByteArrayAsBase64(res.GroupID[:])
			if err != nil {
				return
			}
		case "MinNumBuy":
			decoder.SetFounded()
			res.MinNumBuy, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
		case "StockNumber":
			decoder.SetFounded()
			res.StockNumber, err = decoder.DecodeUInt64()
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
			res.LiveUntil = etime.Time(num)
		case "AllowWeekdays":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.AllowWeekdays = etime.Weekdays(num)
		case "AllowDayhours":
			decoder.SetFounded()
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.AllowDayhours = etime.Dayhours(num)
		case "Description":
			decoder.SetFounded()
			decoder.Offset(1)
			res.Description = decoder.DecodeString()
		case "Type":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.Type = datastore.ProductAuctionType(num)
		case "Status":
			decoder.SetFounded()
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.Status = datastore.ProductAuctionStatus(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getProductAuctionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"WriteTime":`)
	encoder.EncodeInt64(int64(res.WriteTime))

	encoder.EncodeString(`,"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","OrgID":"`)
	encoder.EncodeByteSliceAsBase64(res.OrgID[:])

	encoder.EncodeString(`","WikiID":"`)
	encoder.EncodeByteSliceAsBase64(res.WikiID[:])

	encoder.EncodeString(`","Currency":`)
	encoder.EncodeUInt64(uint64(res.Currency))

	encoder.EncodeString(`,"SuggestPrice":`)
	encoder.EncodeUInt64(uint64(res.SuggestPrice))

	encoder.EncodeString(`,"DistributionCenterCommission":`)
	encoder.EncodeUInt64(uint64(res.DistributionCenterCommission))

	encoder.EncodeString(`,"SellerCommission":`)
	encoder.EncodeUInt64(uint64(res.SellerCommission))

	encoder.EncodeString(`,"Discount":`)
	encoder.EncodeUInt64(uint64(res.Discount))

	encoder.EncodeString(`,"PayablePrice":`)
	encoder.EncodeUInt64(uint64(res.PayablePrice))

	encoder.EncodeString(`,"DistributionCenterID":"`)
	encoder.EncodeByteSliceAsBase64(res.DistributionCenterID[:])

	encoder.EncodeString(`","GroupID":"`)
	encoder.EncodeByteSliceAsBase64(res.GroupID[:])

	encoder.EncodeString(`","MinNumBuy":`)
	encoder.EncodeUInt64(res.MinNumBuy)

	encoder.EncodeString(`,"StockNumber":`)
	encoder.EncodeUInt64(res.StockNumber)

	encoder.EncodeString(`,"LiveUntil":`)
	encoder.EncodeInt64(int64(res.LiveUntil))

	encoder.EncodeString(`,"AllowWeekdays":`)
	encoder.EncodeUInt8(uint8(res.AllowWeekdays))

	encoder.EncodeString(`,"AllowDayhours":`)
	encoder.EncodeUInt64(uint64(res.AllowDayhours))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(res.Description)

	encoder.EncodeString(`","Type":`)
	encoder.EncodeUInt8(uint8(res.Type))

	encoder.EncodeString(`,"Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getProductAuctionRes) jsonLen() (ln int) {
	ln = len(res.Description)
	ln += 879
	return
}
