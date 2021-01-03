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
	"../libgo/uuid"
	"../libgo/validators"
)

var registerNewOrganizationService = achaemenid.Service{
	ID:                2005182932,
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
		lang.LanguageEnglish: "Register Organization",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
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
	SocietyID    [32]byte
	ServicesType datastore.OrganizationAuthenticationType

	Language lang.Language
	Title    string `valid:"OrgName"`
	Domain   string `valid:"Domain"`

	LeaderPersonID [32]byte `json:",string"`
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

	var oa = datastore.OrganizationAuthentication{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               uuid.Random32Byte(),
		SocietyID:        req.SocietyID,
		ServicesType:     req.ServicesType,
		Status:           datastore.OrganizationStatusRegister,
	}
	if req.SocietyID == [32]byte{} {
		oa.SocietyID = achaemenid.Server.Manifest.SocietyUUID
	} else if req.SocietyID != achaemenid.Server.Manifest.SocietyUUID {
		oa.Status = datastore.OrganizationStatusRepresentative
	}

	err = checkQuiddityURI(st, req.Domain)
	if err != nil {
		return
	}
	var q = datastore.Quiddity{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               uuid.Random32Byte(),
		OrgID:            oa.ID,

		Language: req.Language,
		URI:      req.Domain,
		Title:    req.Title,
		Status:   datastore.QuiddityStatusRegister,
	}
	err = q.SaveNew()
	if err != nil {
		return
	}

	oa.QuiddityID = q.ID
	err = oa.SaveNew()
	if err != nil {
		return
	}

	// make new connection for leader
	var uac = datastore.UserAppConnection{
		Status:      datastore.UserAppConnectionIssued,
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
	err = validators.ValidateText(req.Title, 0, 100)
	err = validators.ValidateText(req.Domain, 0, 100)
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

	copy(req.SocietyID[:], buf[0:])
	req.ServicesType = datastore.OrganizationAuthenticationType(syllab.GetUInt8(buf, 32))

	req.Language = lang.Language(syllab.GetUInt32(buf, 33))
	req.Title = syllab.UnsafeGetString(buf, 37)
	req.Domain = syllab.UnsafeGetString(buf, 45)

	copy(req.LeaderPersonID[:], buf[53:])
	return
}

func (req *registerNewOrganizationReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.SocietyID[:])
	syllab.SetUInt8(buf, 32, uint8(req.ServicesType))

	syllab.SetUInt32(buf, 33, uint32(req.Language))
	hsi = syllab.SetString(buf, req.Title, 37, hsi)
	hsi = syllab.SetString(buf, req.Domain, 45, hsi)

	copy(buf[53:], req.LeaderPersonID[:])
	return
}

func (req *registerNewOrganizationReq) syllabStackLen() (ln uint32) {
	return 85
}

func (req *registerNewOrganizationReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Title))
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "SocietyID":
			err = decoder.DecodeByteArrayAsNumber(req.SocietyID[:])
		case "ServicesType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.ServicesType = datastore.OrganizationAuthenticationType(num)

		case "Language":
			var num uint32
			num, err = decoder.DecodeUInt32()
			req.Language = lang.Language(num)
		case "Title":
			req.Title, err = decoder.DecodeString()
		case "Domain":
			req.Domain, err = decoder.DecodeString()

		case "LeaderPersonID":
			err = decoder.DecodeByteArrayAsBase64(req.LeaderPersonID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *registerNewOrganizationReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"SocietyID":[`)
	encoder.EncodeByteSliceAsNumber(req.SocietyID[:])
	encoder.EncodeString(`],"ServicesType":`)
	encoder.EncodeUInt8(uint8(req.ServicesType))

	encoder.EncodeString(`,"Language":`)
	encoder.EncodeUInt32(uint32(req.Language))
	encoder.EncodeString(`,"Title":"`)
	encoder.EncodeString(req.Title)
	encoder.EncodeString(`","Domain":"`)
	encoder.EncodeString(req.Domain)

	encoder.EncodeString(`","LeaderPersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.LeaderPersonID[:])

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *registerNewOrganizationReq) jsonLen() (ln int) {
	ln = len(req.Title) + len(req.Domain)
	ln += 271
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
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
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
