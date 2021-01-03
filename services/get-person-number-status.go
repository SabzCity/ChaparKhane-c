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

var getPersonNumberStatusService = achaemenid.Service{
	ID:                365808761,
	IssueDate:         1602742728,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "GetPersonNumberStatus",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"PersonNumber",
	},

	SRPCHandler: GetPersonNumberStatusSRPC,
	HTTPHandler: GetPersonNumberStatusHTTP,
}

// GetPersonNumberStatusSRPC is sRPC handler of GetPersonNumberStatus service.
func GetPersonNumberStatusSRPC(st *achaemenid.Stream) {
	var req = &getPersonNumberStatusReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getPersonNumberStatusRes
	res, st.Err = getPersonNumberStatus(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetPersonNumberStatusHTTP is HTTP handler of GetPersonNumberStatus service.
func GetPersonNumberStatusHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getPersonNumberStatusReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getPersonNumberStatusRes
	res, st.Err = getPersonNumberStatus(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getPersonNumberStatusReq struct {
	CaptchaID   [16]byte `json:",string"`
	PhoneNumber uint64
}

type getPersonNumberStatusRes struct {
	PersonID [32]byte `json:",string"` // UUID of registered user
	Status   datastore.PersonNumberStatus
}

func getPersonNumberStatus(st *achaemenid.Stream, req *getPersonNumberStatusReq) (res *getPersonNumberStatusRes, err *er.Error) {
	if st.Connection.UserType == authorization.UserTypeGuest {
		// Prevent DDos attack by do some easy process for user e.g. captcha is not good way!
		err = phraseCaptchas.Check(req.CaptchaID)
		if err != nil {
			return
		}
	}

	var pn = datastore.PersonNumber{
		Number: req.PhoneNumber,
	}
	err = pn.GetLastByNumber()
	if err != nil {
		if err.Equal(ganjine.ErrRecordNotFound) {
			err = nil
		} else {
			err = ErrBadSituation
			return
		}
	}

	res = &getPersonNumberStatusRes{
		PersonID: pn.PersonID,
		Status:   pn.Status,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getPersonNumberStatusReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.CaptchaID[:], buf[0:])
	req.PhoneNumber = syllab.GetUInt64(buf, 0)
	return
}

func (req *getPersonNumberStatusReq) syllabStackLen() (ln uint32) {
	return 24
}

func (req *getPersonNumberStatusReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "CaptchaID":
		err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
		case "PhoneNumber":
		req.PhoneNumber, err = decoder.DecodeUInt64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getPersonNumberStatusRes) syllabEncoder(buf []byte) {
	syllab.SetUInt8(buf, 0, uint8(res.Status))
}

func (res *getPersonNumberStatusRes) syllabStackLen() (ln uint32) {
	return 1
}

func (res *getPersonNumberStatusRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getPersonNumberStatusRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getPersonNumberStatusRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"PersonID":"`)
	encoder.EncodeByteSliceAsBase64(res.PersonID[:])

	encoder.EncodeString(`","Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getPersonNumberStatusRes) jsonLen() (ln int) {
	ln = 71
	return
}
