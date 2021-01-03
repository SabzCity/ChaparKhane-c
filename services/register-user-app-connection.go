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
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/uuid"
	"../libgo/validators"
)

var registerUserAppConnectionService = achaemenid.Service{
	ID:                2264014142,
	IssueDate:         1603876567,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register User App Connection",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"UserAppConnection",
	},

	SRPCHandler: RegisterUserAppConnectionSRPC,
	HTTPHandler: RegisterUserAppConnectionHTTP,
}

// RegisterUserAppConnectionSRPC is sRPC handler of RegisterUserAppConnection service.
func RegisterUserAppConnectionSRPC(st *achaemenid.Stream) {
	var req = &registerUserAppConnectionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerUserAppConnectionRes
	res, st.Err = registerUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterUserAppConnectionHTTP is HTTP handler of RegisterUserAppConnection service.
func RegisterUserAppConnectionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerUserAppConnectionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerUserAppConnectionRes
	res, st.Err = registerUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerUserAppConnectionReq struct {
	Description string `valid:"text[0:50]"`
	Weight      achaemenid.Weight

	ThingID          [32]byte `json:",string"`
	DelegateUserID   [32]byte `json:",string"`
	DelegateUserType authorization.UserType

	PublicKey     [32]byte `json:",string"`
	AccessControl authorization.AccessControl
}

type registerUserAppConnectionRes struct {
	ID [32]byte `json:",string"`
}

func registerUserAppConnection(st *achaemenid.Stream, req *registerUserAppConnectionReq) (res *registerUserAppConnectionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	// User can't delegate to yourself!
	if req.DelegateUserID == st.Connection.UserID {
		err = authorization.ErrNotAllowToDelegate
		return
	}
	// Person user type can't delegate to register new connection
	if st.Connection.UserType == authorization.UserTypePerson && st.Connection.DelegateUserType != authorization.UserTypeUnset {
		err = authorization.ErrNotAllowToDelegate
		return
	}
	// Org Connection can't be empty UserDelegateID
	if st.Connection.UserType == authorization.UserTypeOrg && req.DelegateUserID == [32]byte{} {
		err = authorization.ErrNotAllowToNotDelegate
		return
	}
	// Org Connection can't delegate to other org
	if st.Connection.UserType == authorization.UserTypeOrg && req.DelegateUserType == authorization.UserTypeOrg {
		err = authorization.ErrNotAllowToDelegate
		return
	}

	var uac = datastore.UserAppConnection{
		Status:      datastore.UserAppConnectionIssued,
		Description: req.Description,

		ID:     uuid.Random32Byte(),
		Weight: req.Weight,

		ThingID:          req.ThingID,
		UserID:           st.Connection.UserID,
		UserType:         st.Connection.UserType,
		DelegateUserID:   req.DelegateUserID,
		DelegateUserType: req.DelegateUserType,

		PeerPublicKey: req.PublicKey,
		AccessControl: req.AccessControl,
	}
	err = uac.SaveNew()
	if err != nil {
		return
	}

	res = &registerUserAppConnectionRes{
		ID: uac.ID,
	}

	return
}

func (req *registerUserAppConnectionReq) validator() (err *er.Error) {
	err = validators.ValidateText(req.Description, 0, 50)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerUserAppConnectionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Description = syllab.UnsafeGetString(buf, 0)
	req.Weight = achaemenid.Weight(syllab.GetUInt8(buf, 8))
	copy(req.ThingID[:], buf[9:])
	copy(req.DelegateUserID[:], buf[41:])
	req.DelegateUserType = authorization.UserType(syllab.GetUInt8(buf, 73))
	copy(req.PublicKey[:], buf[74:])
	req.AccessControl.SyllabDecoder(buf, 106)
	return
}

func (req *registerUserAppConnectionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.Description, 0, hsi)
	syllab.SetUInt8(buf, 8, uint8(req.Weight))
	copy(buf[9:], req.ThingID[:])
	copy(buf[41:], req.DelegateUserID[:])
	syllab.SetUInt8(buf, 73, uint8(req.DelegateUserType))
	copy(buf[74:], req.PublicKey[:])
	req.AccessControl.SyllabEncoder(buf, 106, hsi)
	return
}

func (req *registerUserAppConnectionReq) syllabStackLen() (ln uint32) {
	return 106 + req.AccessControl.SyllabStackLen()
}

func (req *registerUserAppConnectionReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Description))
	ln += req.AccessControl.SyllabHeapLen()
	return
}

func (req *registerUserAppConnectionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerUserAppConnectionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "Description":
			req.Description, err = decoder.DecodeString()
		case "Weight":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.Weight = achaemenid.Weight(num)
		case "ThingID":
			err = decoder.DecodeByteArrayAsBase64(req.ThingID[:])
		case "DelegateUserID":
			err = decoder.DecodeByteArrayAsBase64(req.DelegateUserID[:])
		case "DelegateUserType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.DelegateUserType = authorization.UserType(num)
		case "PublicKey":
			err = decoder.DecodeByteArrayAsBase64(req.PublicKey[:])

		case "AccessControl":
			err = req.AccessControl.JSONDecoder(decoder)

		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *registerUserAppConnectionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Description":"`)
	encoder.EncodeString(req.Description)

	encoder.EncodeString(`","Weight":`)
	encoder.EncodeUInt8(uint8(req.Weight))

	encoder.EncodeString(`,"ThingID":"`)
	encoder.EncodeByteSliceAsBase64(req.ThingID[:])

	encoder.EncodeString(`","DelegateUserID":"`)
	encoder.EncodeByteSliceAsBase64(req.DelegateUserID[:])

	encoder.EncodeString(`","DelegateUserType":`)
	encoder.EncodeUInt8(uint8(req.DelegateUserType))

	encoder.EncodeString(`,"PublicKey":"`)
	encoder.EncodeByteSliceAsBase64(req.PublicKey[:])

	encoder.EncodeString(`","AccessControl":`)
	req.AccessControl.JSONEncoder(encoder)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerUserAppConnectionReq) jsonLen() (ln int) {
	ln = len(req.Description)
	ln += req.AccessControl.JSONLen()
	ln += 251
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerUserAppConnectionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerUserAppConnectionRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerUserAppConnectionRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerUserAppConnectionRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerUserAppConnectionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerUserAppConnectionRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *registerUserAppConnectionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerUserAppConnectionRes) jsonLen() (ln int) {
	ln = 52
	return
}
