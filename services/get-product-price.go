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

var getProductPriceService = achaemenid.Service{
	ID:                1514978150,
	IssueDate:         1608092738,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Product Price",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductPrice",
	},

	SRPCHandler: GetProductPriceSRPC,
	HTTPHandler: GetProductPriceHTTP,
}

// GetProductPriceSRPC is sRPC handler of GetProductPrice service.
func GetProductPriceSRPC(st *achaemenid.Stream) {
	var req = &getProductPriceReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getProductPriceRes
	res, st.Err = getProductPrice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetProductPriceHTTP is HTTP handler of GetProductPrice service.
func GetProductPriceHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getProductPriceReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getProductPriceRes
	res, st.Err = getProductPrice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getProductPriceReq struct {
	QuiddityID [32]byte `json:",string"`
}

type getProductPriceRes struct {
	WriteTime etime.Time

	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	OrgID            [32]byte `json:",string"`

	MaterialsPercent   math.PerMyriad
	MaterialsCost      price.Amount
	LaborPercent       math.PerMyriad
	LaborCost          price.Amount
	InvestmentsPercent math.PerMyriad
	InvestmentsCost    price.Amount
	TotalCost          price.Amount

	Markup          math.PerMyriad
	WholesaleProfit price.Amount
	Margin          math.PerMyriad
	RetailProfit    price.Amount

	TaxPercent math.PerMyriad
	Tax        price.Amount

	Price price.Amount
}

func getProductPrice(st *achaemenid.Stream, req *getProductPriceReq) (res *getProductPriceRes, err *er.Error) {
	var pp = datastore.ProductPrice{
		QuiddityID: req.QuiddityID,
	}
	err = pp.GetLastByQuiddityID()
	if err != nil {
		return
	}

	res = &getProductPriceRes{
		WriteTime: pp.WriteTime,

		AppInstanceID:    pp.AppInstanceID,
		UserConnectionID: pp.UserConnectionID,
		OrgID:            pp.OrgID,

		TaxPercent: pp.TaxPercent,
		Tax:        pp.Tax,

		Price: pp.Price,
	}

	if st.Connection.UserID == pp.OrgID {
		res.MaterialsPercent = pp.MaterialsPercent
		res.MaterialsCost = pp.MaterialsCost
		res.LaborPercent = pp.LaborPercent
		res.LaborCost = pp.LaborCost
		res.InvestmentsPercent = pp.InvestmentsPercent
		res.InvestmentsCost = pp.InvestmentsCost
		res.TotalCost = pp.TotalCost

		res.Markup = pp.Markup
		res.WholesaleProfit = pp.WholesaleProfit
		res.Margin = pp.Margin
		res.RetailProfit = pp.RetailProfit
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getProductPriceReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])
	return
}

func (req *getProductPriceReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])
	return
}

func (req *getProductPriceReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getProductPriceReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getProductPriceReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getProductPriceReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *getProductPriceReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getProductPriceReq) jsonLen() (ln int) {
	ln = 56
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getProductPriceRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	copy(res.OrgID[:], buf[72:])

	res.MaterialsPercent = math.PerMyriad(syllab.GetUInt16(buf, 104))
	res.MaterialsCost = price.Amount(syllab.GetInt64(buf, 106))
	res.LaborPercent = math.PerMyriad(syllab.GetUInt16(buf, 114))
	res.LaborCost = price.Amount(syllab.GetInt64(buf, 116))
	res.InvestmentsPercent = math.PerMyriad(syllab.GetUInt16(buf, 124))
	res.InvestmentsCost = price.Amount(syllab.GetInt64(buf, 126))
	res.TotalCost = price.Amount(syllab.GetInt64(buf, 134))
	res.Markup = math.PerMyriad(syllab.GetUInt16(buf, 142))
	res.WholesaleProfit = price.Amount(syllab.GetInt64(buf, 144))
	res.Margin = math.PerMyriad(syllab.GetUInt16(buf, 152))
	res.RetailProfit = price.Amount(syllab.GetInt64(buf, 154))

	res.TaxPercent = math.PerMyriad(syllab.GetUInt16(buf, 162))
	res.Tax = price.Amount(syllab.GetInt64(buf, 164))

	res.Price = price.Amount(syllab.GetInt64(buf, 172))
	return
}

func (res *getProductPriceRes) syllabEncoder(buf []byte) {
	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	copy(buf[72:], res.OrgID[:])

	syllab.SetUInt16(buf, 104, uint16(res.MaterialsPercent))
	syllab.SetInt64(buf, 106, int64(res.MaterialsCost))
	syllab.SetUInt16(buf, 114, uint16(res.LaborPercent))
	syllab.SetInt64(buf, 116, int64(res.LaborCost))
	syllab.SetUInt16(buf, 124, uint16(res.InvestmentsPercent))
	syllab.SetInt64(buf, 126, int64(res.InvestmentsCost))
	syllab.SetInt64(buf, 134, int64(res.TotalCost))

	syllab.SetUInt16(buf, 142, uint16(res.Markup))
	syllab.SetInt64(buf, 144, int64(res.WholesaleProfit))
	syllab.SetUInt16(buf, 152, uint16(res.Margin))
	syllab.SetInt64(buf, 154, int64(res.RetailProfit))

	syllab.SetUInt16(buf, 162, uint16(res.TaxPercent))
	syllab.SetInt64(buf, 164, int64(res.Tax))

	syllab.SetInt64(buf, 172, int64(res.Price))
	return
}

func (res *getProductPriceRes) syllabStackLen() (ln uint32) {
	return 180
}

func (res *getProductPriceRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getProductPriceRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getProductPriceRes) jsonDecoder(buf []byte) (err *er.Error) {
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

		case "MaterialsPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.MaterialsPercent = math.PerMyriad(num)
		case "MaterialsCost":
			var num int64
			num, err = decoder.DecodeInt64()
			res.MaterialsCost = price.Amount(num)
		case "LaborPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.LaborPercent = math.PerMyriad(num)
		case "LaborCost":
			var num int64
			num, err = decoder.DecodeInt64()
			res.LaborCost = price.Amount(num)
		case "InvestmentsPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.InvestmentsPercent = math.PerMyriad(num)
		case "InvestmentsCost":
			var num int64
			num, err = decoder.DecodeInt64()
			res.InvestmentsCost = price.Amount(num)
		case "TotalCost":
			var num int64
			num, err = decoder.DecodeInt64()
			res.TotalCost = price.Amount(num)

		case "Markup":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.Markup = math.PerMyriad(num)
		case "WholesaleProfit":
			var num int64
			num, err = decoder.DecodeInt64()
			res.WholesaleProfit = price.Amount(num)
		case "Margin":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.Margin = math.PerMyriad(num)
		case "RetailProfit":
			var num int64
			num, err = decoder.DecodeInt64()
			res.RetailProfit = price.Amount(num)

		case "TaxPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			res.TaxPercent = math.PerMyriad(num)
		case "Tax":
			var num int64
			num, err = decoder.DecodeInt64()
			res.Tax = price.Amount(num)

		case "Price":
			var num int64
			num, err = decoder.DecodeInt64()
			res.Price = price.Amount(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getProductPriceRes) jsonEncoder() (buf []byte) {
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

	encoder.EncodeString(`","MaterialsPercent":`)
	encoder.EncodeUInt16(uint16(res.MaterialsPercent))
	encoder.EncodeString(`,"MaterialsCost":`)
	encoder.EncodeInt64(int64(res.MaterialsCost))
	encoder.EncodeString(`,"LaborPercent":`)
	encoder.EncodeUInt16(uint16(res.LaborPercent))
	encoder.EncodeString(`,"LaborCost":`)
	encoder.EncodeInt64(int64(res.LaborCost))
	encoder.EncodeString(`,"InvestmentsPercent":`)
	encoder.EncodeUInt16(uint16(res.InvestmentsPercent))
	encoder.EncodeString(`,"InvestmentsCost":`)
	encoder.EncodeInt64(int64(res.InvestmentsCost))
	encoder.EncodeString(`,"TotalCost":`)
	encoder.EncodeInt64(int64(res.TotalCost))

	encoder.EncodeString(`,"Markup":`)
	encoder.EncodeUInt16(uint16(res.Markup))
	encoder.EncodeString(`,"WholesaleProfit":`)
	encoder.EncodeInt64(int64(res.WholesaleProfit))
	encoder.EncodeString(`,"Margin":`)
	encoder.EncodeUInt16(uint16(res.Margin))
	encoder.EncodeString(`,"RetailProfit":`)
	encoder.EncodeInt64(int64(res.RetailProfit))

	encoder.EncodeString(`,"TaxPercent":`)
	encoder.EncodeUInt16(uint16(res.TaxPercent))
	encoder.EncodeString(`,"Tax":`)
	encoder.EncodeInt64(int64(res.Tax))

	encoder.EncodeString(`,"Price":`)
	encoder.EncodeInt64(int64(res.Price))
	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getProductPriceRes) jsonLen() (ln int) {
	ln = 610
	return
}
