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
	"../libgo/price"
	"../libgo/srpc"
	"../libgo/syllab"
)

var updateProductPriceService = achaemenid.Service{
	ID:                3392068160,
	IssueDate:         1608092793,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Update Product Price",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductPrice",
	},

	SRPCHandler: UpdateProductPriceSRPC,
	HTTPHandler: UpdateProductPriceHTTP,
}

// UpdateProductPriceSRPC is sRPC handler of UpdateProductPrice service.
func UpdateProductPriceSRPC(st *achaemenid.Stream) {
	var req = &updateProductPriceReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateProductPrice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateProductPriceHTTP is HTTP handler of UpdateProductPrice service.
func UpdateProductPriceHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateProductPriceReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateProductPrice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateProductPriceReq struct {
	QuiddityID [32]byte `json:",string"`

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

func updateProductPrice(st *achaemenid.Stream, req *updateProductPriceReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var pp = datastore.ProductPrice{
		QuiddityID: req.QuiddityID,
	}
	err = pp.GetLastByQuiddityID()
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = ErrProductPriceNotRegistered
		return
	}
	if err != nil {
		return
	}
	if pp.OrgID != st.Connection.UserID {
		err = authorization.ErrUserNotAllow
		return
	}

	pp = datastore.ProductPrice{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            st.Connection.UserID,
		QuiddityID:           req.QuiddityID,

		MaterialsPercent:   req.MaterialsPercent,
		MaterialsCost:      req.MaterialsCost,
		LaborPercent:       req.LaborPercent,
		LaborCost:          req.LaborCost,
		InvestmentsPercent: req.InvestmentsPercent,
		InvestmentsCost:    req.InvestmentsCost,
		TotalCost:          req.TotalCost,

		Markup:          req.Markup,
		WholesaleProfit: req.WholesaleProfit,
		Margin:          req.Margin,
		RetailProfit:    req.RetailProfit,

		TaxPercent: req.TaxPercent,
		Tax:        req.Tax,

		Price: req.Price,
	}
	err = pp.SaveNew()
	if err != nil {
		return
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateProductPriceReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])

	req.MaterialsPercent = math.PerMyriad(syllab.GetUInt16(buf, 32))
	req.MaterialsCost = price.Amount(syllab.GetInt64(buf, 34))
	req.LaborPercent = math.PerMyriad(syllab.GetUInt16(buf, 42))
	req.LaborCost = price.Amount(syllab.GetInt64(buf, 44))
	req.InvestmentsPercent = math.PerMyriad(syllab.GetUInt16(buf, 52))
	req.InvestmentsCost = price.Amount(syllab.GetInt64(buf, 54))
	req.TotalCost = price.Amount(syllab.GetInt64(buf, 62))

	req.Markup = math.PerMyriad(syllab.GetUInt16(buf, 70))
	req.WholesaleProfit = price.Amount(syllab.GetInt64(buf, 72))
	req.Margin = math.PerMyriad(syllab.GetUInt16(buf, 80))
	req.RetailProfit = price.Amount(syllab.GetInt64(buf, 82))

	req.TaxPercent = math.PerMyriad(syllab.GetUInt16(buf, 90))
	req.Tax = price.Amount(syllab.GetInt64(buf, 92))

	req.Price = price.Amount(syllab.GetInt64(buf, 100))
	return
}

func (req *updateProductPriceReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])

	syllab.SetUInt16(buf, 32, uint16(req.MaterialsPercent))
	syllab.SetInt64(buf, 34, int64(req.MaterialsCost))
	syllab.SetUInt16(buf, 42, uint16(req.LaborPercent))
	syllab.SetInt64(buf, 44, int64(req.LaborCost))
	syllab.SetUInt16(buf, 52, uint16(req.InvestmentsPercent))
	syllab.SetInt64(buf, 54, int64(req.InvestmentsCost))
	syllab.SetInt64(buf, 62, int64(req.TotalCost))

	syllab.SetUInt16(buf, 70, uint16(req.Markup))
	syllab.SetInt64(buf, 72, int64(req.WholesaleProfit))
	syllab.SetUInt16(buf, 80, uint16(req.Margin))
	syllab.SetInt64(buf, 82, int64(req.RetailProfit))

	syllab.SetUInt16(buf, 90, uint16(req.TaxPercent))
	syllab.SetInt64(buf, 92, int64(req.Tax))

	syllab.SetInt64(buf, 100, int64(req.Price))
	return
}

func (req *updateProductPriceReq) syllabStackLen() (ln uint32) {
	return 108
}

func (req *updateProductPriceReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *updateProductPriceReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateProductPriceReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])

		case "MaterialsPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.MaterialsPercent = math.PerMyriad(num)
		case "MaterialsCost":
			var num int64
			num, err = decoder.DecodeInt64()
			req.MaterialsCost = price.Amount(num)
		case "LaborPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.LaborPercent = math.PerMyriad(num)
		case "LaborCost":
			var num int64
			num, err = decoder.DecodeInt64()
			req.LaborCost = price.Amount(num)
		case "InvestmentsPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.InvestmentsPercent = math.PerMyriad(num)
		case "InvestmentsCost":
			var num int64
			num, err = decoder.DecodeInt64()
			req.InvestmentsCost = price.Amount(num)
		case "TotalCost":
			var num int64
			num, err = decoder.DecodeInt64()
			req.TotalCost = price.Amount(num)

		case "Markup":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.Markup = math.PerMyriad(num)
		case "WholesaleProfit":
			var num int64
			num, err = decoder.DecodeInt64()
			req.WholesaleProfit = price.Amount(num)
		case "Margin":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.Margin = math.PerMyriad(num)
		case "RetailProfit":
			var num int64
			num, err = decoder.DecodeInt64()
			req.RetailProfit = price.Amount(num)

		case "TaxPercent":
			var num uint16
			num, err = decoder.DecodeUInt16()
			req.TaxPercent = math.PerMyriad(num)
		case "Tax":
			var num int64
			num, err = decoder.DecodeInt64()
			req.Tax = price.Amount(num)

		case "Price":
			var num int64
			num, err = decoder.DecodeInt64()
			req.Price = price.Amount(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *updateProductPriceReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])

	encoder.EncodeString(`","MaterialsPercent":`)
	encoder.EncodeUInt16(uint16(req.MaterialsPercent))
	encoder.EncodeString(`,"MaterialsCost":`)
	encoder.EncodeInt64(int64(req.MaterialsCost))
	encoder.EncodeString(`,"LaborPercent":`)
	encoder.EncodeUInt16(uint16(req.LaborPercent))
	encoder.EncodeString(`,"LaborCost":`)
	encoder.EncodeInt64(int64(req.LaborCost))
	encoder.EncodeString(`,"InvestmentsPercent":`)
	encoder.EncodeUInt16(uint16(req.InvestmentsPercent))
	encoder.EncodeString(`,"InvestmentsCost":`)
	encoder.EncodeInt64(int64(req.InvestmentsCost))
	encoder.EncodeString(`,"TotalCost":`)
	encoder.EncodeInt64(int64(req.TotalCost))

	encoder.EncodeString(`,"Markup":`)
	encoder.EncodeUInt16(uint16(req.Markup))
	encoder.EncodeString(`,"WholesaleProfit":`)
	encoder.EncodeInt64(int64(req.WholesaleProfit))
	encoder.EncodeString(`,"Margin":`)
	encoder.EncodeUInt16(uint16(req.Margin))
	encoder.EncodeString(`,"RetailProfit":`)
	encoder.EncodeInt64(int64(req.RetailProfit))

	encoder.EncodeString(`,"TaxPercent":`)
	encoder.EncodeUInt16(uint16(req.TaxPercent))
	encoder.EncodeString(`,"Tax":`)
	encoder.EncodeInt64(int64(req.Tax))

	encoder.EncodeString(`,"Price":`)
	encoder.EncodeInt64(int64(req.Price))
	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *updateProductPriceReq) jsonLen() (ln int) {
	ln = 451
	return
}
