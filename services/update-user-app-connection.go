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
)

var updateUserAppConnectionService = achaemenid.Service{
	ID:                1678413553,
	CRUD:              authorization.CRUDUpdate,
	IssueDate:         1604152374,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Update User App Connection",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "Use to update or expire or revoke the connection",
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

	Status      datastore.UserAppsConnectionStatus
	Description string
	Weight      achaemenid.Weight

	PublicKey     [32]byte `json:",string"`
	AccessControl authorization.AccessControl
}

func updateUserAppConnection(st *achaemenid.Stream, req *updateUserAppConnectionReq) (err *er.Error) {
	// Person user type can't delegate to update the connection
	if st.Connection.UserType == achaemenid.UserTypePerson && st.Connection.DelegateUserType != achaemenid.UserTypeUnset {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	err = st.Authorize()
	if err != nil {
		return
	}

	var uac = datastore.UserAppsConnection{
		ID: req.ID,
	}
	err = uac.GetLastByID()
	if err != nil {
		return
	}

	if uac.Status == datastore.UserAppsConnectionExpired || uac.Status == datastore.UserAppsConnectionRevoked {
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

/*
	Request Encoders & Decoders
*/

func (req *updateUserAppConnectionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.Status = datastore.UserAppsConnectionStatus(syllab.GetUInt8(buf, 32))
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
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
			if err != nil {
				return
			}
		case 'S':
			decoder.SetFounded()
			decoder.Offset(8)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Status = datastore.UserAppsConnectionStatus(num)
		case 'D':
			decoder.SetFounded()
			decoder.Offset(14)
			req.Description = decoder.DecodeString()
		case 'W':
			decoder.SetFounded()
			decoder.Offset(8)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Weight = achaemenid.Weight(num)
		case 'P':
			decoder.SetFounded()
			decoder.Offset(12)
			err = decoder.DecodeByteArrayAsBase64(req.PublicKey[:])
			if err != nil {
				return
			}
		case 'A':
			decoder.SetFounded()
			decoder.Offset(15)
			decoder.Buf, err = req.AccessControl.JSONDecoder(decoder.Buf)
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
	encoder.Buf = req.AccessControl.JSONEncoder(encoder.Buf)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *updateUserAppConnectionReq) jsonLen() (ln int) {
	ln += len(req.Description) + req.AccessControl.JSONLen()
	ln += 207
	return
}
