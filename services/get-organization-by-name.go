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

var getOrganizationByNameService = achaemenid.Service{
	ID:                3649049198,
	IssueDate:         1604475046,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get Organization By Name",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "Return last OrgID associate with given name. Maybe last org change its name and not same with given name!",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: GetOrganizationByNameSRPC,
	HTTPHandler: GetOrganizationByNameHTTP,
}

// GetOrganizationByNameSRPC is sRPC handler of GetOrganizationByName service.
func GetOrganizationByNameSRPC(st *achaemenid.Stream) {
	var req = &getOrganizationByNameReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getOrganizationByNameRes
	res, st.Err = getOrganizationByName(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetOrganizationByNameHTTP is HTTP handler of GetOrganizationByName service.
func GetOrganizationByNameHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getOrganizationByNameReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getOrganizationByNameRes
	res, st.Err = getOrganizationByName(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getOrganizationByNameReq struct {
	Name string `valid:"OrgName"`
}

type getOrganizationByNameRes struct {
	ID [32]byte `json:",string"`
}

func getOrganizationByName(st *achaemenid.Stream, req *getOrganizationByNameReq) (res *getOrganizationByNameRes, err *er.Error) {
	var oa = datastore.OrganizationAuthentication{
		Name: req.Name,
	}

	var IDs [][32]byte
	IDs, err = oa.GetIDsByNameByHashIndex(18446744073709551615, 1)
	if err != nil || IDs == nil {
		return
	}

	res = &getOrganizationByNameRes{
		ID: IDs[0],
	}

	return
}

/*
	Request Encoders & Decoders
*/

func (req *getOrganizationByNameReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Name = syllab.UnsafeGetString(buf, 0)
	return
}

func (req *getOrganizationByNameReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.Name, 0, hsi)
	return
}

func (req *getOrganizationByNameReq) syllabStackLen() (ln uint32) {
	return 8
}

func (req *getOrganizationByNameReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Name))
	return
}

func (req *getOrganizationByNameReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getOrganizationByNameReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'N':
			decoder.SetFounded()
			decoder.Offset(7)
			req.Name = decoder.DecodeString()
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getOrganizationByNameReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Name":"`)
	encoder.EncodeString(req.Name)

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getOrganizationByNameReq) jsonLen() (ln int) {
	ln = len(req.Name)
	ln += 11
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getOrganizationByNameRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *getOrganizationByNameRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *getOrganizationByNameRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *getOrganizationByNameRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getOrganizationByNameRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getOrganizationByNameRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
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

func (res *getOrganizationByNameRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *getOrganizationByNameRes) jsonLen() (ln int) {
	ln = 52
	return
}
