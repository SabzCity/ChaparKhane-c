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
		lang.LanguageEnglish: "Get Product Auction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
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
	QuiddityID       [32]byte `json:",string"`

	// Price
	Discount         math.PerMyriad
	DCCommission     math.PerMyriad
	SellerCommission math.PerMyriad

	// Authorization
	Authorization authorization.Product

	Description string
	Type        datastore.ProductAuctionType
	Status      datastore.ProductAuctionStatus
}

func getProductAuction(st *achaemenid.Stream, req *getProductAuctionReq) (res *getProductAuctionRes, err *er.Error) {
	var pa = datastore.ProductAuction{
		ID: req.ID,
	}
	err = pa.GetLastByID()
	if err != nil {
		return
	}

	res = &getProductAuctionRes{
		WriteTime: pa.WriteTime,

		AppInstanceID:    pa.AppInstanceID,
		UserConnectionID: pa.UserConnectionID,
		OrgID:            pa.OrgID,
		QuiddityID:       pa.QuiddityID,

		// Price
		Discount:         pa.Discount,
		DCCommission:     pa.DCCommission,
		SellerCommission: pa.SellerCommission,

		// Authorization
		Authorization: pa.Authorization,

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
	// var add, ln uint32
	// var tempSlice []byte

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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
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
	copy(res.QuiddityID[:], buf[104:])

	res.Discount = math.PerMyriad(syllab.GetUInt16(buf, 136))
	res.DCCommission = math.PerMyriad(syllab.GetUInt16(buf, 138))
	res.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 140))

	res.Authorization.SyllabDecoder(buf, 142)

	res.Description = syllab.UnsafeGetString(buf, 142+res.Authorization.SyllabStackLen())
	res.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 150+res.Authorization.SyllabStackLen()))
	res.Status = datastore.ProductAuctionStatus(syllab.GetUInt8(buf, 151+res.Authorization.SyllabStackLen()))
	return
}

func (res *getProductAuctionRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	copy(buf[72:], res.OrgID[:])
	copy(buf[104:], res.QuiddityID[:])

	syllab.SetUInt16(buf, 136, uint16(res.Discount))
	syllab.SetUInt16(buf, 138, uint16(res.DCCommission))
	syllab.SetUInt16(buf, 140, uint16(res.SellerCommission))

	hsi = res.Authorization.SyllabEncoder(buf, 142, hsi)

	hsi = syllab.SetString(buf, res.Description, 142+res.Authorization.SyllabStackLen(), hsi)
	syllab.SetUInt8(buf, 150+res.Authorization.SyllabStackLen(), uint8(res.Type))
	syllab.SetUInt8(buf, 151+res.Authorization.SyllabStackLen(), uint8(res.Status))
	return
}

func (res *getProductAuctionRes) syllabStackLen() (ln uint32) {
	return 152 + res.Authorization.SyllabStackLen()
}

func (res *getProductAuctionRes) syllabHeapLen() (ln uint32) {
	ln += res.Authorization.SyllabHeapLen()
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "WriteTime":
			var num int64
			num, err = decoder.DecodeInt64()
			res.WriteTime = etime.Time(num)
		case "AppInstanceID":
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
		case "UserConnectionID":
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
		case "OrgID":
			err = decoder.DecodeByteArrayAsBase64(res.OrgID[:])
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(res.QuiddityID[:])

		case "Discount":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.Discount = math.PerMyriad(num)
		case "DCCommission":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.DCCommission = math.PerMyriad(num)
		case "SellerCommission":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.SellerCommission = math.PerMyriad(num)

		case "Authorization":
			err = res.Authorization.JSONDecoder(decoder)

		case "Description":
			res.Description, err = decoder.DecodeString()
		case "Type":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Type = datastore.ProductAuctionType(num)
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Status = datastore.ProductAuctionStatus(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
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

	encoder.EncodeString(`","QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(res.QuiddityID[:])

	encoder.EncodeString(`","Discount":`)
	encoder.EncodeUInt16(uint16(res.Discount))

	encoder.EncodeString(`,"DCCommission":`)
	encoder.EncodeUInt16(uint16(res.DCCommission))

	encoder.EncodeString(`,"SellerCommission":`)
	encoder.EncodeUInt16(uint16(res.SellerCommission))

	encoder.EncodeString(`,"Authorization":`)
	res.Authorization.JSONEncoder(encoder)

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
	ln += res.Authorization.JSONLen()
	ln += 394
	return
}
