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
		lang.LanguageEnglish: "Update Product Auction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
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
	Discount                     math.PerMyriad
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad

	// Authorization
	Authorization authorization.Product

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
	err = pa.GetLastByID()
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
	if pa.Status == datastore.QuiddityStatusBlocked {
		err = ErrBlockedByJustice
		return
	}

	pa = datastore.ProductAuction{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            pa.OrgID,
		ID:               pa.ID,
		QuiddityID:       pa.QuiddityID,

		Discount:         req.Discount,
		DCCommission:     req.DistributionCenterCommission,
		SellerCommission: req.SellerCommission,

		// Authorization
		Authorization: req.Authorization,

		Description: req.Description,
		Type:        req.Type,
		Status:      datastore.ProductAuctionUpdated,
	}

	if req.Authorization.LiveUntil != 0 && req.Authorization.LiveUntil.Pass(etime.Now()) {
		pa.Status = datastore.ProductAuctionExpired
	}

	err = pa.Set()
	if err != nil {
		return
	}
	pa.IndexRecordIDForID()
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

	req.Discount = math.PerMyriad(syllab.GetUInt16(buf, 32))
	req.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 34))
	req.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 36))

	req.Authorization.SyllabDecoder(buf, 38)

	req.Description = syllab.UnsafeGetString(buf, 38+req.Authorization.SyllabStackLen())
	req.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 46+req.Authorization.SyllabStackLen()))
	return
}

func (req *updateProductAuctionReq) syllabEncoder(buf []byte) {
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

func (req *updateProductAuctionReq) syllabStackLen() (ln uint32) {
	return 47 + req.Authorization.SyllabStackLen()
}

func (req *updateProductAuctionReq) syllabHeapLen() (ln uint32) {
	ln += req.Authorization.SyllabHeapLen()
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

func (req *updateProductAuctionReq) jsonEncoder() (buf []byte) {
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

func (req *updateProductAuctionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += req.Authorization.JSONLen()
	ln += 179
	return
}
