/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getOrganizationService = achaemenid.Service{
	ID:                2889005042,
	IssueDate:         1604475115,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Organization",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: GetOrganizationSRPC,
	HTTPHandler: GetOrganizationHTTP,
}

// GetOrganizationSRPC is sRPC handler of GetOrganization service.
func GetOrganizationSRPC(st *achaemenid.Stream) {
	var req = &getOrganizationReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getOrganizationRes
	res, st.Err = getOrganization(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetOrganizationHTTP is HTTP handler of GetOrganization service.
func GetOrganizationHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getOrganizationReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getOrganizationRes
	res, st.Err = getOrganization(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getOrganizationReq struct {
	ID [32]byte `json:",string"`
}

type getOrganizationRes struct {
	WriteTime        etime.Time
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`

	SocietyID    [32]byte `json:",string"`
	QuiddityID   [32]byte `json:",string"`
	ServicesType datastore.OrganizationAuthenticationType
	Status       datastore.OrganizationAuthenticationStatus
}

func getOrganization(st *achaemenid.Stream, req *getOrganizationReq) (res *getOrganizationRes, err *er.Error) {
	var oa = datastore.OrganizationAuthentication{
		ID: req.ID,
	}
	err = oa.GetLastByID()
	if err != nil {
		return
	}

	res = &getOrganizationRes{
		WriteTime:     oa.WriteTime,
		AppInstanceID: oa.AppInstanceID,
		// UserConnectionID: oa.UserConnectionID, TODO::: Due to HTTP use ConnectionID to authenticate connections can't enable it now!!

		SocietyID:    oa.SocietyID,
		QuiddityID:   oa.QuiddityID,
		ServicesType: oa.ServicesType,
		Status:       oa.Status,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getOrganizationReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getOrganizationReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	return
}

func (req *getOrganizationReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getOrganizationReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getOrganizationReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getOrganizationReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *getOrganizationReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getOrganizationReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getOrganizationRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	copy(res.SocietyID[:], buf[72:])
	copy(res.QuiddityID[:], buf[104:])
	res.ServicesType = datastore.OrganizationAuthenticationType(syllab.GetUInt8(buf, 136))
	res.Status = datastore.OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 137))
	return
}

func (res *getOrganizationRes) syllabEncoder(buf []byte) {
	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	copy(buf[72:], res.SocietyID[:])
	copy(buf[104:], res.QuiddityID[:])
	syllab.SetUInt8(buf, 136, uint8(res.ServicesType))
	syllab.SetUInt8(buf, 137, uint8(res.Status))
	return
}

func (res *getOrganizationRes) syllabStackLen() (ln uint32) {
	return 138
}

func (res *getOrganizationRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getOrganizationRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getOrganizationRes) jsonDecoder(buf []byte) (err *er.Error) {
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
		case "SocietyID":
			err = decoder.DecodeByteArrayAsBase64(res.SocietyID[:])
		case "QuiddityID":
			err = decoder.DecodeByteArrayAsBase64(res.QuiddityID[:])
		case "ServicesType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.ServicesType = datastore.OrganizationAuthenticationType(num)
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Status = datastore.OrganizationAuthenticationStatus(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getOrganizationRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"WriteTime":`)
	encoder.EncodeInt64(int64(res.WriteTime))

	encoder.EncodeString(`,"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","SocietyID":"`)
	encoder.EncodeByteSliceAsBase64(res.SocietyID[:])

	encoder.EncodeString(`","QuiddityID":"`)
	encoder.EncodeByteSliceAsBase64(res.QuiddityID[:])

	encoder.EncodeString(`","ServicesType":`)
	encoder.EncodeUInt8(uint8(res.ServicesType))

	encoder.EncodeString(`,"Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getOrganizationRes) jsonLen() (ln int) {
	ln = 480
	return
}
