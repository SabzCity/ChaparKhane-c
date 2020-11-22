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

var getOrganizationByIDService = achaemenid.Service{
	ID:                2713628543,
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
		lang.EnglishLanguage: "Get Organization By ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: GetOrganizationByIDSRPC,
	HTTPHandler: GetOrganizationByIDHTTP,
}

// GetOrganizationByIDSRPC is sRPC handler of GetOrganizationByID service.
func GetOrganizationByIDSRPC(st *achaemenid.Stream) {
	var req = &getOrganizationByIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getOrganizationByIDRes
	res, st.Err = getOrganizationByID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetOrganizationByIDHTTP is HTTP handler of GetOrganizationByID service.
func GetOrganizationByIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getOrganizationByIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getOrganizationByIDRes
	res, st.Err = getOrganizationByID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getOrganizationByIDReq struct {
	ID [32]byte `json:",string"`
}

type getOrganizationByIDRes struct {
	WriteTime        etime.Time
	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`

	SocietyID             uint32
	Name                  string
	Domain                string
	FinancialCreditAmount int64
	ThingID               [32]byte `json:",string"`
	ServicesType          datastore.OrganizationAuthenticationType
	Status                datastore.OrganizationAuthenticationStatus
}

func getOrganizationByID(st *achaemenid.Stream, req *getOrganizationByIDReq) (res *getOrganizationByIDRes, err *er.Error) {
	var oa = datastore.OrganizationAuthentication{
		ID: req.ID,
	}

	var IDs [][32]byte
	IDs, err = oa.GetRecordsIDByIDByHashIndex(18446744073709551615, 1)
	if err != nil || IDs == nil {
		return
	}

	oa.RecordID = IDs[0]
	oa.GetByRecordID()

	res = &getOrganizationByIDRes{
		WriteTime:     oa.WriteTime,
		AppInstanceID: oa.AppInstanceID,
		// UserConnectionID: oa.UserConnectionID, TODO::: Due to HTTP use ConnectionID to authenticate connections can't enable it now!!

		Name:                  oa.Name,
		Domain:                oa.Domain,
		FinancialCreditAmount: oa.FinancialCreditAmount,
		ThingID:               oa.ThingID,
		ServicesType:          oa.ServicesType,
		Status:                oa.Status,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getOrganizationByIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getOrganizationByIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	return
}

func (req *getOrganizationByIDReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getOrganizationByIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getOrganizationByIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getOrganizationByIDReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getOrganizationByIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getOrganizationByIDReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getOrganizationByIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])

	res.Name = syllab.UnsafeGetString(buf, 72)
	res.Domain = syllab.UnsafeGetString(buf, 80)
	res.FinancialCreditAmount = syllab.GetInt64(buf, 88)
	copy(res.ThingID[:], buf[96:])
	res.ServicesType = datastore.OrganizationAuthenticationType(syllab.GetUInt8(buf, 128))
	res.Status = datastore.OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 129))
	return
}

func (res *getOrganizationByIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])

	hsi = syllab.SetString(buf, res.Name, 72, hsi)
	hsi = syllab.SetString(buf, res.Domain, 80, hsi)
	syllab.SetInt64(buf, 88, res.FinancialCreditAmount)
	copy(buf[96:], res.ThingID[:])
	syllab.SetUInt8(buf, 128, uint8(res.ServicesType))
	syllab.SetUInt8(buf, 129, uint8(res.Status))
	return
}

func (res *getOrganizationByIDRes) syllabStackLen() (ln uint32) {
	return 130
}

func (res *getOrganizationByIDRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Name))
	ln += uint32(len(res.Domain))
	return
}

func (res *getOrganizationByIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getOrganizationByIDRes) jsonDecoder(buf []byte) (err *er.Error) {
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
			res.WriteTime = etime.Time(num)
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
			decoder.Offset(7)
			res.Name = decoder.DecodeString()
		case 'D':
			decoder.SetFounded()
			decoder.Offset(9)
			res.Domain = decoder.DecodeString()
		case 'F':
			decoder.SetFounded()
			decoder.Offset(23)
			res.FinancialCreditAmount, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
		case 'T':
			decoder.SetFounded()
			decoder.Offset(10)
			err = decoder.DecodeByteArrayAsBase64(res.ThingID[:])
			if err != nil {
				return
			}
		case 'S':
			switch decoder.Buf[1] {
			case 'e':
				decoder.SetFounded()
				decoder.Offset(14)
				var num uint8
				num, err = decoder.DecodeUInt8()
				if err != nil {
					return
				}
				res.ServicesType = datastore.OrganizationAuthenticationType(num)
			case 't':
				decoder.SetFounded()
				decoder.Offset(8)
				var num uint8
				num, err = decoder.DecodeUInt8()
				if err != nil {
					return
				}
				res.Status = datastore.OrganizationAuthenticationStatus(num)
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getOrganizationByIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"WriteTime":`)
	encoder.EncodeInt64(int64(res.WriteTime))

	encoder.EncodeString(`,"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","Name":"`)
	encoder.EncodeString(res.Name)

	encoder.EncodeString(`","Domain":"`)
	encoder.EncodeString(res.Domain)

	encoder.EncodeString(`","FinancialCreditAmount":`)
	encoder.EncodeInt64(res.FinancialCreditAmount)

	encoder.EncodeString(`,"ThingID":"`)
	encoder.EncodeByteSliceAsBase64(res.ThingID[:])

	encoder.EncodeString(`","ServicesType":`)
	encoder.EncodeUInt8(uint8(res.ServicesType))

	encoder.EncodeString(`,"Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getOrganizationByIDRes) jsonLen() (ln int) {
	ln = len(res.Name) + len(res.Domain)
	ln += 360
	return
}
