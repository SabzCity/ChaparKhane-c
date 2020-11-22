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

var getOrganizationByDomainService = achaemenid.Service{
	ID:                1317623156,
	IssueDate:         1604475058,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get Organization By Domain",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "Return last OrgID associate with given domain. Maybe last org change its domain and not same with given domain!",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: GetOrganizationByDomainSRPC,
	HTTPHandler: GetOrganizationByDomainHTTP,
}

// GetOrganizationByDomainSRPC is sRPC handler of GetOrganizationByDomain service.
func GetOrganizationByDomainSRPC(st *achaemenid.Stream) {
	var req = &getOrganizationByDomainReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getOrganizationByDomainRes
	res, st.Err = getOrganizationByDomain(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetOrganizationByDomainHTTP is HTTP handler of GetOrganizationByDomain service.
func GetOrganizationByDomainHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getOrganizationByDomainReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getOrganizationByDomainRes
	res, st.Err = getOrganizationByDomain(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getOrganizationByDomainReq struct {
	Domain string `valid:"Domain"`
}

type getOrganizationByDomainRes struct {
	ID [32]byte `json:",string"`
}

func getOrganizationByDomain(st *achaemenid.Stream, req *getOrganizationByDomainReq) (res *getOrganizationByDomainRes, err *er.Error) {
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	var oa = datastore.OrganizationAuthentication{
		Domain: req.Domain,
	}

	var IDs [][32]byte
	IDs, err = oa.GetIDsByDomainByHashIndex(18446744073709551615, 1)
	if err != nil || IDs == nil {
		return
	}

	res = &getOrganizationByDomainRes{
		ID: IDs[0],
	}

	return
}

func (req *getOrganizationByDomainReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getOrganizationByDomainReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Domain = syllab.UnsafeGetString(buf, 0)
	return
}

func (req *getOrganizationByDomainReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.Domain, 0, hsi)
	return
}

func (req *getOrganizationByDomainReq) syllabStackLen() (ln uint32) {
	return 8
}

func (req *getOrganizationByDomainReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Domain))
	return
}

func (req *getOrganizationByDomainReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getOrganizationByDomainReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'D':
			decoder.SetFounded()
			decoder.Offset(9)
			req.Domain = decoder.DecodeString()
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getOrganizationByDomainReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Domain":"`)
	encoder.EncodeString(req.Domain)

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getOrganizationByDomainReq) jsonLen() (ln int) {
	ln += 0 + len(req.Domain)
	ln += 13
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getOrganizationByDomainRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *getOrganizationByDomainRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *getOrganizationByDomainRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *getOrganizationByDomainRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getOrganizationByDomainRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getOrganizationByDomainRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *getOrganizationByDomainRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *getOrganizationByDomainRes) jsonLen() (ln int) {
	ln = 52
	return
}
