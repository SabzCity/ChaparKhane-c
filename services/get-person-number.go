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
	"../libgo/srpc"
	"../libgo/syllab"
)

var getPersonNumberService = achaemenid.Service{
	ID:                500496613,
	IssueDate:         1603724046,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypePerson,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Person Number",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
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
	WriteTime        etime.Time
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	Number           uint64
	Status           datastore.PersonNumberStatus
}

func getPersonNumber(st *achaemenid.Stream) (res *getPersonNumberRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	// By use st.Connection.UserID just active user can retrieve number that own and don't need to check it anymore!
	var pn = datastore.PersonNumber{
		PersonID: st.Connection.UserID,
	}
	err = pn.GetLastByPersonID()
	if err != nil {
		if !err.Equal(ganjine.ErrRecordNotFound) {
			err = ErrBadSituation
		}
		return
	}

	res = &getPersonNumberRes{
		WriteTime:     pn.WriteTime,
		AppInstanceID: pn.AppInstanceID,
		// UserConnectionID: pn.UserConnectionID,  TODO::: Due to HTTP use ConnectionID to authenticate connections can't enable it now!!

		Number: pn.Number,
		Status: pn.Status,
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

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	res.Number = syllab.GetUInt64(buf, 72)
	res.Status = datastore.PersonNumberStatus(syllab.GetUInt8(buf, 80))
	return
}

func (res *getPersonNumberRes) syllabEncoder(buf []byte) {
	syllab.SetInt64(buf, 0, int64(res.WriteTime))
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
		case "Number":
			res.Number, err = decoder.DecodeUInt64()
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Status = datastore.PersonNumberStatus(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
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
	encoder.EncodeUInt64(res.Number)

	encoder.EncodeString(`,"Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getPersonNumberRes) jsonLen() (ln int) {
	ln = 187
	return
}
