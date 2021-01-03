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

var getLastOrganizationsIDService = achaemenid.Service{
	ID:                278690539,
	IssueDate:         1604573930,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Last Organizations ID",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "return last 9 org ID register in platform in last 30 days",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: GetLastOrganizationsIDSRPC,
	HTTPHandler: GetLastOrganizationsIDHTTP,
}

// GetLastOrganizationsIDSRPC is sRPC handler of GetLastOrganizationsID service.
func GetLastOrganizationsIDSRPC(st *achaemenid.Stream) {
	var res *getLastOrganizationsIDRes
	res, st.Err = getLastOrganizationsID(st)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetLastOrganizationsIDHTTP is HTTP handler of GetLastOrganizationsID service.
func GetLastOrganizationsIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var res *getLastOrganizationsIDRes
	res, st.Err = getLastOrganizationsID(st)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getLastOrganizationsIDRes struct {
	IDs [][32]byte
}

func getLastOrganizationsID(st *achaemenid.Stream) (res *getLastOrganizationsIDRes, err *er.Error) {
	res = &getLastOrganizationsIDRes{}
	var oa = datastore.OrganizationAuthentication{
		WriteTime: etime.Now(),
	}
	res.IDs, err = oa.FindLastIDs(18446744073709551615, 9, 30)
	if err.Equal(ganjine.ErrRecordNotFound) {
		err = nil
	}
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getLastOrganizationsIDRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.IDs = syllab.UnsafeGet32ByteArraySlice(buf, 0)
	return
}

func (res *getLastOrganizationsIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!
	syllab.Set32ByteArrayArray(buf, res.IDs, 0, hsi)
	return
}

func (res *getLastOrganizationsIDRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getLastOrganizationsIDRes) syllabHeapLen() (ln uint32) {
	ln = uint32(len(res.IDs) * 32)
	return
}

func (res *getLastOrganizationsIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getLastOrganizationsIDRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "IDs":
			res.IDs, err = decoder.Decode32ByteArraySliceAsBase64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getLastOrganizationsIDRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"IDs":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.IDs)

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *getLastOrganizationsIDRes) jsonLen() (ln int) {
	ln += len(res.IDs) * 46
	ln += 10
	return
}
