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
	"../libgo/uuid"
)

var registerNewOrganizationService = achaemenid.Service{
	ID:                3017822306,
	IssueDate:         1604472315,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypePerson,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Register New Organization",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: RegisterNewOrganizationSRPC,
	HTTPHandler: RegisterNewOrganizationHTTP,
}

// RegisterNewOrganizationSRPC is sRPC handler of RegisterNewOrganization service.
func RegisterNewOrganizationSRPC(st *achaemenid.Stream) {
	var req = &registerNewOrganizationReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerNewOrganizationRes
	res, st.Err = registerNewOrganization(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterNewOrganizationHTTP is HTTP handler of RegisterNewOrganization service.
func RegisterNewOrganizationHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerNewOrganizationReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerNewOrganizationRes
	res, st.Err = registerNewOrganization(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerNewOrganizationReq struct {
	LeaderPersonID        [32]byte `json:",string"`
	Name                  string   `valid:"OrgName"`
	Domain                string   `valid:"Domain"`
	FinancialCreditAmount int64
	ServicesType          datastore.OrganizationAuthenticationType
}

type registerNewOrganizationRes struct {
	ID [32]byte `json:",string"`
}

func registerNewOrganization(st *achaemenid.Stream, req *registerNewOrganizationReq) (res *registerNewOrganizationRes, err *er.Error) {
	if st.Connection.UserID != adminUserID {
		err = authorization.ErrUserNotAllow
		return
	}

	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	// Check Founder status
	var getPersonStatusReq = getPersonStatusReq{
		PersonID: req.LeaderPersonID,
	}
	var getPersonStatusRes *getPersonStatusRes
	getPersonStatusRes, err = getPersonStatus(st, &getPersonStatusReq, true)
	if err != nil {
		return
	}
	if getPersonStatusRes.Status == datastore.PersonAuthenticationBlocked {
		err = ErrBlockedPerson
		return
	}

	err = checkOrgName(st, req.Name)
	if err != nil {
		return
	}
	err = checkOrgDomain(st, req.Domain)
	if err != nil {
		return
	}

	var oa = datastore.OrganizationAuthentication{
		AppInstanceID:         server.Nodes.LocalNode.InstanceID,
		UserConnectionID:      st.Connection.ID,
		ID:                    uuid.Random32Byte(),
		SocietyID:             server.Manifest.SocietyID,
		Name:                  req.Name,
		Domain:                req.Domain,
		FinancialCreditAmount: req.FinancialCreditAmount,
		ServicesType:          req.ServicesType,
		Status:                datastore.OrganizationStatusRegister,
	}
	err = oa.SaveNew()
	if err != nil {
		return
	}

	// make new connection for leader
	var uac = datastore.UserAppsConnection{
		Status:      datastore.UserAppsConnectionIssued,
		Description: "Leader connection created in create organization",

		ID: uuid.Random32Byte(),

		// ThingID:          st.Connection.ThingID,
		UserID:           oa.ID,
		UserType:         authorization.UserTypeOrg,
		DelegateUserID:   req.LeaderPersonID,
		DelegateUserType: st.Connection.UserType,
	}
	uac.AccessControl.GiveFullAccess()
	err = uac.SaveNew()
	if err != nil {
		// TODO::: Can't return easily!!
		return
	}

	res = &registerNewOrganizationRes{
		ID: oa.ID,
	}
	return
}

func (req *registerNewOrganizationReq) validator() (err *er.Error) {
	return
}

func checkOrgName(st *achaemenid.Stream, name string) (err *er.Error) {
	var goByNameReq = getOrganizationByNameReq{
		Name: name,
	}
	var goByNameRes *getOrganizationByNameRes
	goByNameRes, err = getOrganizationByName(st, &goByNameReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return
	}

	var goByIDReq = getOrganizationByIDReq{
		ID: goByNameRes.ID,
	}
	var goByIDRes *getOrganizationByIDRes
	goByIDRes, err = getOrganizationByID(st, &goByIDReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		// TODO::: how it is possible???
		return nil
	}
	if err != nil {
		return
	}
	if goByIDRes.Name != name {
		return nil
	}
	if goByIDRes.Status != datastore.OrganizationStatusClosed {
		err = ErrOrgNameRegistered
		return
	}
	return
}

func checkOrgDomain(st *achaemenid.Stream, domain string) (err *er.Error) {
	var goByDomainReq = getOrganizationByDomainReq{
		Domain: domain,
	}
	var goByDomainRes *getOrganizationByDomainRes
	goByDomainRes, err = getOrganizationByDomain(st, &goByDomainReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return
	}

	var goByIDReq = getOrganizationByIDReq{
		ID: goByDomainRes.ID,
	}
	var goByIDRes *getOrganizationByIDRes
	goByIDRes, err = getOrganizationByID(st, &goByIDReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		// TODO::: how it is possible???
		return nil
	}
	if err != nil {
		return
	}
	if goByIDRes.Domain != domain {
		return nil
	}
	if goByIDRes.Status != datastore.OrganizationStatusClosed {
		err = ErrOrgDomainRegistered
		return
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerNewOrganizationReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.LeaderPersonID[:], buf[0:])
	req.Name = syllab.UnsafeGetString(buf, 32)
	req.Domain = syllab.UnsafeGetString(buf, 40)
	req.FinancialCreditAmount = syllab.GetInt64(buf, 48)
	req.ServicesType = datastore.OrganizationAuthenticationType(syllab.GetUInt8(buf, 56))
	return
}

func (req *registerNewOrganizationReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.LeaderPersonID[:])
	hsi = syllab.SetString(buf, req.Name, 32, hsi)
	hsi = syllab.SetString(buf, req.Domain, 40, hsi)
	syllab.SetInt64(buf, 48, req.FinancialCreditAmount)
	syllab.SetUInt8(buf, 56, uint8(req.ServicesType))
	return
}

func (req *registerNewOrganizationReq) syllabStackLen() (ln uint32) {
	return 57
}

func (req *registerNewOrganizationReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Name))
	ln += uint32(len(req.Domain))
	return
}

func (req *registerNewOrganizationReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerNewOrganizationReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'L':
			decoder.SetFounded()
			decoder.Offset(17)
			err = decoder.DecodeByteArrayAsBase64(req.LeaderPersonID[:])
			if err != nil {
				return
			}
		case 'F':
			decoder.SetFounded()
			decoder.Offset(23)
			req.FinancialCreditAmount, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
		case 'N':
			decoder.SetFounded()
			decoder.Offset(7)
			req.Name = decoder.DecodeString()
		case 'D':
			decoder.SetFounded()
			decoder.Offset(9)
			req.Domain = decoder.DecodeString()
		case 'S':
			decoder.SetFounded()
			decoder.Offset(14)
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			req.ServicesType = datastore.OrganizationAuthenticationType(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *registerNewOrganizationReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"LeaderPersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.LeaderPersonID[:])

	encoder.EncodeString(`","Name":"`)
	encoder.EncodeString(req.Name)

	encoder.EncodeString(`","Domain":"`)
	encoder.EncodeString(req.Domain)

	encoder.EncodeString(`","FinancialCreditAmount":`)
	encoder.EncodeInt64(req.FinancialCreditAmount)

	encoder.EncodeString(`,"ServicesType":`)
	encoder.EncodeUInt8(uint8(req.ServicesType))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerNewOrganizationReq) jsonLen() (ln int) {
	ln = len(req.Name) + len(req.Domain)
	ln += 185
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerNewOrganizationRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerNewOrganizationRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerNewOrganizationRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerNewOrganizationRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerNewOrganizationRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerNewOrganizationRes) jsonDecoder(buf []byte) (err *er.Error) {
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

func (res *registerNewOrganizationRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerNewOrganizationRes) jsonLen() (ln int) {
	ln = 52
	return
}
