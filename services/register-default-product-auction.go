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
		lang.LanguageEnglish: "Register Default Product Auction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
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
	QuiddityID [32]byte      `json:",string"`
	Language   lang.Language // Just use to check quiddity exist and belong to requested org

	Discount                     math.PerMyriad
	DistributionCenterCommission math.PerMyriad
	SellerCommission             math.PerMyriad

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

	var pa = datastore.ProductAuction{
		QuiddityID: req.QuiddityID,
	}
	var IDs [][32]byte
	IDs, err = pa.FindIDsByQuiddityID(0, 1)
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
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		OrgID:            st.Connection.UserID,
		ID:               uuid.Random32Byte(),
		QuiddityID:       req.QuiddityID,

		Discount:         req.Discount,
		DCCommission:     req.DistributionCenterCommission,
		SellerCommission: req.SellerCommission,

		Authorization: authorization.Product{
			AllowUserType: authorization.UserTypeAll,
			AllowWeekdays: etime.WeekdaysAll,
			AllowDayhours: etime.DayhoursAll,
		},

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

	copy(req.QuiddityID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))

	req.Discount = math.PerMyriad(syllab.GetUInt16(buf, 36))
	req.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 38))
	req.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 40))

	req.Description = syllab.UnsafeGetString(buf, 42)
	req.Type = datastore.ProductAuctionType(syllab.GetUInt8(buf, 50))
	return
}

func (req *registerDefaultProductAuctionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.QuiddityID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))

	syllab.SetUInt16(buf, 36, uint16(req.Discount))
	syllab.SetUInt16(buf, 38, uint16(req.DistributionCenterCommission))
	syllab.SetUInt16(buf, 40, uint16(req.SellerCommission))

	syllab.SetString(buf, req.Description, 42, hsi)
	syllab.SetUInt8(buf, 50, uint8(req.Type))
	return
}

func (req *registerDefaultProductAuctionReq) syllabStackLen() (ln uint32) {
	return 51
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(req.QuiddityID[:])
		case "Language":
			var num uint32
			num, err = decoder.DecodeUInt32()
			req.Language = lang.Language(num)

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

func (req *registerDefaultProductAuctionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(req.QuiddityID[:])
	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt32(uint32(req.Language))

	encoder.EncodeString(`,"Discount":`)
	encoder.EncodeUInt16(uint16(req.Discount))
	encoder.EncodeString(`,"DistributionCenterCommission":`)
	encoder.EncodeUInt16(uint16(req.DistributionCenterCommission))
	encoder.EncodeString(`,"SellerCommission":`)
	encoder.EncodeUInt16(uint16(req.SellerCommission))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(req.Description)
	encoder.EncodeString(`","Type":`)
	encoder.EncodeUInt8(uint8(req.Type))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerDefaultProductAuctionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += 185
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
