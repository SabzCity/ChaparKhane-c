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
	"../libgo/math"
	"../libgo/price"
	"../libgo/srpc"
	"../libgo/syllab"
)

var registerProductPriceService = achaemenid.Service{
	ID:                2291828869,
	URI:               "", // API services can set like "/apis?2291828869" but it is not efficient, find services by ID.
	IssueDate:         1608092756,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Product Price",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"ProductPrice",
	},

	SRPCHandler: RegisterProductPriceSRPC,
	HTTPHandler: RegisterProductPriceHTTP,
}

// RegisterProductPriceSRPC is sRPC handler of RegisterProductPrice service.
func RegisterProductPriceSRPC(st *achaemenid.Stream) {
	var req = &registerProductPriceReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}
	st.Err = registerProductPrice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// RegisterProductPriceHTTP is HTTP handler of RegisterProductPrice service.
func RegisterProductPriceHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerProductPriceReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = registerProductPrice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type registerProductPriceReq struct {
	QuiddityID [32]byte      `json:",string"`
	Language   lang.Language // Just use to check quiddity exist and belong to requested org

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

func registerProductPrice(st *achaemenid.Stream, req *registerProductPriceReq) (err *er.Error) {
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

	var pp = datastore.ProductPrice{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            st.Connection.UserID,
		QuiddityID:       req.QuiddityID,

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

func (req *registerProductPriceReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.QuiddityID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))

	req.MaterialsPercent = math.PerMyriad(syllab.GetUInt16(buf, 36))
	req.MaterialsCost = price.Amount(syllab.GetInt64(buf, 38))
	req.LaborPercent = math.PerMyriad(syllab.GetUInt16(buf, 46))
	req.LaborCost = price.Amount(syllab.GetInt64(buf, 48))
	req.InvestmentsPercent = math.PerMyriad(syllab.GetUInt16(buf, 56))
	req.InvestmentsCost = price.Amount(syllab.GetInt64(buf, 58))
	req.TotalCost = price.Amount(syllab.GetInt64(buf, 66))

	req.Markup = math.PerMyriad(syllab.GetUInt16(buf, 74))
	req.WholesaleProfit = price.Amount(syllab.GetInt64(buf, 76))
	req.Margin = math.PerMyriad(syllab.GetUInt16(buf, 84))
	req.RetailProfit = price.Amount(syllab.GetInt64(buf, 86))

	req.TaxPercent = math.PerMyriad(syllab.GetUInt16(buf, 94))
	req.Tax = price.Amount(syllab.GetInt64(buf, 96))

	req.Price = price.Amount(syllab.GetInt64(buf, 104))
	return
}

func (req *registerProductPriceReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.QuiddityID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))

	syllab.SetUInt16(buf, 36, uint16(req.MaterialsPercent))
	syllab.SetInt64(buf, 38, int64(req.MaterialsCost))
	syllab.SetUInt16(buf, 46, uint16(req.LaborPercent))
	syllab.SetInt64(buf, 48, int64(req.LaborCost))
	syllab.SetUInt16(buf, 56, uint16(req.InvestmentsPercent))
	syllab.SetInt64(buf, 58, int64(req.InvestmentsCost))
	syllab.SetInt64(buf, 66, int64(req.TotalCost))

	syllab.SetUInt16(buf, 74, uint16(req.Markup))
	syllab.SetInt64(buf, 76, int64(req.WholesaleProfit))
	syllab.SetUInt16(buf, 84, uint16(req.Margin))
	syllab.SetInt64(buf, 86, int64(req.RetailProfit))

	syllab.SetUInt16(buf, 94, uint16(req.TaxPercent))
	syllab.SetInt64(buf, 96, int64(req.Tax))

	syllab.SetInt64(buf, 104, int64(req.Price))
	return
}

func (req *registerProductPriceReq) syllabStackLen() (ln uint32) {
	return 112
}

func (req *registerProductPriceReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *registerProductPriceReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerProductPriceReq) jsonDecoder(buf []byte) (err *er.Error) {
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

func (req *registerProductPriceReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])
	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt32(uint32(req.Language))

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

func (req *registerProductPriceReq) jsonLen() (ln int) {
	ln = 558
	return
}
