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
	"../libgo/validators"
)

var updateUserAppConnectionService = achaemenid.Service{
	ID:                1678413553,
	IssueDate:         1604152374,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Update User App Connection",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Use to update or expire or revoke the connection",
	},
	TAGS: []string{
		"",
	},

	SRPCHandler: UpdateUserAppConnectionSRPC,
	HTTPHandler: UpdateUserAppConnectionHTTP,
}

// UpdateUserAppConnectionSRPC is sRPC handler of UpdateUserAppConnection service.
func UpdateUserAppConnectionSRPC(st *achaemenid.Stream) {
	var req = &updateUserAppConnectionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateUserAppConnectionHTTP is HTTP handler of UpdateUserAppConnection service.
func UpdateUserAppConnectionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateUserAppConnectionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateUserAppConnection(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}
	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateUserAppConnectionReq struct {
	ID [32]byte `json:",string"`

	Status      datastore.UserAppConnectionStatus
	Description string `valid:"text[0:50]"`
	Weight      achaemenid.Weight

	PublicKey     [32]byte `json:",string"`
	AccessControl authorization.AccessControl
}

func updateUserAppConnection(st *achaemenid.Stream, req *updateUserAppConnectionReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	// Person user type can't delegate to update the connection
	if st.Connection.UserType == authorization.UserTypePerson && st.Connection.DelegateUserType != authorization.UserTypeUnset {
		err = authorization.ErrUserNotAllow
		return
	}

	var uac = datastore.UserAppConnection{
		ID: req.ID,
	}
	err = uac.GetLastByID()
	if err != nil {
		return
	}

	if uac.Status == datastore.UserAppConnectionExpired || uac.Status == datastore.UserAppConnectionRevoked {
		// err =
		return
	}

	// TODO::: tel all platform servers about changes

	uac.Status = req.Status
	uac.Description = req.Description
	uac.Weight = req.Weight
	uac.PeerPublicKey = req.PublicKey
	uac.AccessControl = req.AccessControl

	return
}

func (req *updateUserAppConnectionReq) validator() (err *er.Error) {
	err = validators.ValidateText(req.Description, 0, 50)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateUserAppConnectionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.Status = datastore.UserAppConnectionStatus(syllab.GetUInt8(buf, 32))
	req.Description = syllab.UnsafeGetString(buf, 33)
	req.Weight = achaemenid.Weight(syllab.GetUInt8(buf, 41))
	copy(req.PublicKey[:], buf[42:])
	req.AccessControl.SyllabDecoder(buf, 74)
	return
}

func (req *updateUserAppConnectionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])
	syllab.SetUInt8(buf, 32, uint8(req.Status))
	hsi = syllab.SetString(buf, req.Description, 33, hsi)
	syllab.SetUInt8(buf, 41, uint8(req.Weight))
	copy(buf[42:], req.PublicKey[:])
	req.AccessControl.SyllabEncoder(buf, 74, hsi)
	return
}

func (req *updateUserAppConnectionReq) syllabStackLen() (ln uint32) {
	return 74 + req.AccessControl.SyllabStackLen()
}

func (req *updateUserAppConnectionReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Description))
	ln += req.AccessControl.SyllabHeapLen()
	return
}

func (req *updateUserAppConnectionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateUserAppConnectionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.Status = datastore.UserAppConnectionStatus(num)
		case "Description":
			req.Description, err = decoder.DecodeString()
		case "Weight":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.Weight = achaemenid.Weight(num)
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

func (req *updateUserAppConnectionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Status":`)
	encoder.EncodeUInt8(uint8(req.Status))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(req.Description)

	encoder.EncodeString(`","Weight":`)
	encoder.EncodeUInt8(uint8(req.Weight))

	encoder.EncodeString(`,"PublicKey":"`)
	encoder.EncodeByteSliceAsBase64(req.PublicKey[:])

	encoder.EncodeString(`","AccessControl":`)
	req.AccessControl.JSONEncoder(encoder)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *updateUserAppConnectionReq) jsonLen() (ln int) {
	ln = len(req.Description) + req.AccessControl.JSONLen()
	ln += 173
	return
}
