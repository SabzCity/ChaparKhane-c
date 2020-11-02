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
	"../libgo/srpc"
	"../libgo/syllab"
)

var getPersonNumberService = achaemenid.Service{
	ID:                500496613,
	URI:               "", // API services can set like "/apis?500496613" but it is not efficient, find services by ID.
	CRUD:              authorization.CRUDRead,
	IssueDate:         1603724046,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get Person Number",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"PersonNumber",
	},

	SRPCHandler: GetPersonNumberSRPC,
	HTTPHandler: GetPersonNumberHTTP,
}

// GetPersonNumberSRPC is sRPC handler of GetPersonNumber service.
func GetPersonNumberSRPC(st *achaemenid.Stream) {
	var res *getPersonNumberRes
	res, st.Err = getPersonNumber(st)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetPersonNumberHTTP is HTTP handler of GetPersonNumber service.
func GetPersonNumberHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var res *getPersonNumberRes
	res, st.Err = getPersonNumber(st)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getPersonNumberRes struct {
	WriteTime        int64
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	Number           uint64
	Status           datastore.PersonNumberStatus
}

func getPersonNumber(st *achaemenid.Stream) (res *getPersonNumberRes, err *er.Error) {
	if st.Connection.UserType != achaemenid.UserTypePerson {
		err = authorization.ErrAuthorizationUserNotAllow
		return
	}

	err = st.Authorize()
	if err != nil {
		return
	}

	var pn = datastore.PersonNumber{
		PersonID: st.Connection.UserID,
	}
	err = pn.GetLastByPersonID()
	if err != nil {
		if err != ganjine.ErrGanjineRecordNotFound {
			err = ErrPlatformBadSituation
		}
		return
	}

	res = &getPersonNumberRes{
		WriteTime:        pn.WriteTime,
		AppInstanceID:    pn.AppInstanceID,
		UserConnectionID: pn.UserConnectionID,
		Number:           pn.Number,
		Status:           pn.Status,
	}
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getPersonNumberRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = syllab.GetInt64(buf, 0)
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	res.Number = syllab.GetUInt64(buf, 72)
	res.Status = datastore.PersonNumberStatus(syllab.GetUInt8(buf, 80))
	return
}

func (res *getPersonNumberRes) syllabEncoder(buf []byte) {
	syllab.SetInt64(buf, 0, res.WriteTime)
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	syllab.SetUInt64(buf, 72, res.Number)
	syllab.SetUInt8(buf, 80, uint8(res.Status))
	return
}

func (res *getPersonNumberRes) syllabStackLen() (ln uint32) {
	return 81
}

func (res *getPersonNumberRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getPersonNumberRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getPersonNumberRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'W':
			decoder.SetFounded()
			decoder.Offset(11)
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			res.WriteTime = int64(num)
		case 'A':
			decoder.SetFounded()
			decoder.Offset(16)
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
			if err != nil {
				return
			}
		case 'U':
			decoder.SetFounded()
			decoder.Offset(19)
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
			if err != nil {
				return
			}
		case 'N':
			decoder.SetFounded()
			decoder.Offset(8)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.Number = uint64(num)
		case 'S':
			decoder.SetFounded()
			decoder.Offset(8)
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			res.Status = datastore.PersonNumberStatus(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getPersonNumberRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"WriteTime":`)
	encoder.EncodeInt64(int64(res.WriteTime))

	encoder.EncodeString(`,"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","Number":`)
	encoder.EncodeUInt64(uint64(res.Number))

	encoder.EncodeString(`,"Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getPersonNumberRes) jsonLen() (ln int) {
	ln = 20 + 43 + 43 + 20 + 3 + 75
	return
}
