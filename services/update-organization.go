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
)

var updateOrganizationService = achaemenid.Service{
	ID:                1900937469,
	IssueDate:         1604472326,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeOrg | authorization.UserTypePerson, // TODO::: remove person when register org free by everyone.
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Update Organization",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"OrganizationAuthentication",
	},

	SRPCHandler: UpdateOrganizationSRPC,
	HTTPHandler: UpdateOrganizationHTTP,
}

// UpdateOrganizationSRPC is sRPC handler of UpdateOrganization service.
func UpdateOrganizationSRPC(st *achaemenid.Stream) {
	var req = &updateOrganizationReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateOrganization(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateOrganizationHTTP is HTTP handler of UpdateOrganization service.
func UpdateOrganizationHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateOrganizationReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateOrganization(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}
	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateOrganizationReq struct {
	LeaderPersonID        [32]byte `json:",string"`
	ID                    [32]byte `json:",string"`
	Name                  string   `valid:"OrgName"`
	Domain                string   `valid:"Domain"`
	FinancialCreditAmount int64
	ThingID               [32]byte `json:",string"`
	ServicesType          datastore.OrganizationAuthenticationType
	Status                datastore.OrganizationAuthenticationStatus
}

func updateOrganization(st *achaemenid.Stream, req *updateOrganizationReq) (err *er.Error) {
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

	var oa = datastore.OrganizationAuthentication{
		ID: req.ID,
	}
	err = oa.GetLastByIDByHashIndex()
	if err != nil {
		return
	}

	if oa.Status == datastore.OrganizationStatusBlocked {
		err = ErrBlockedByJustice
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

	var oldName = oa.Name
	var oldDomain = oa.Domain

	oa = datastore.OrganizationAuthentication{
		AppInstanceID:         server.Nodes.LocalNode.InstanceID,
		UserConnectionID:      st.Connection.ID,
		ID:                    req.ID,
		SocietyID:             oa.SocietyID,
		Name:                  req.Name,
		Domain:                req.Domain,
		FinancialCreditAmount: req.FinancialCreditAmount,
		ServicesType:          req.ServicesType,
		Status:                req.Status,
	}
	err = oa.Set()
	if err != nil {
		return
	}
	oa.HashIndexRecordIDForID()
	if req.Name != oldName {
		oa.HashIndexIDForName()
	}
	if req.Domain != oldDomain && req.Domain != "" {
		oa.HashIndexIDForDomain()
	}
	if req.LeaderPersonID != [32]byte{} {
		// make new connection for leader
		var uac = datastore.UserAppsConnection{
			Status:      datastore.UserAppsConnectionIssued,
			Description: "Leader connection created in update organization",

			ID: uuid.Random32Byte(),

			// ThingID:          st.Connection.ThingID,
			UserID:           oa.ID,
			UserType:         authorization.UserTypeOrg,
			DelegateUserID:   req.LeaderPersonID,
			DelegateUserType: st.Connection.UserType,
		}
		err = uac.SaveNew()
		if err != nil {
			return
		}
	}

	return
}

func (req *updateOrganizationReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateOrganizationReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.LeaderPersonID[:], buf[0:])
	copy(req.ID[:], buf[32:])
	req.Name = syllab.UnsafeGetString(buf, 64)
	req.Domain = syllab.UnsafeGetString(buf, 72)
	req.FinancialCreditAmount = syllab.GetInt64(buf, 80)
	copy(req.ThingID[:], buf[88:])
	req.ServicesType = datastore.OrganizationAuthenticationType(syllab.GetUInt8(buf, 120))
	req.Status = datastore.OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 121))
	return
}

func (req *updateOrganizationReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.LeaderPersonID[:])
	copy(buf[32:], req.ID[:])
	hsi = syllab.SetString(buf, req.Name, 64, hsi)
	hsi = syllab.SetString(buf, req.Domain, 72, hsi)
	syllab.SetInt64(buf, 80, req.FinancialCreditAmount)
	copy(buf[88:], req.ThingID[:])
	syllab.SetUInt8(buf, 120, uint8(req.ServicesType))
	syllab.SetUInt8(buf, 121, uint8(req.Status))
	return
}

func (req *updateOrganizationReq) syllabStackLen() (ln uint32) {
	return 122
}

func (req *updateOrganizationReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.Name))
	ln += uint32(len(req.Domain))
	return
}

func (req *updateOrganizationReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateOrganizationReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
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
		case 'F':
			decoder.SetFounded()
			decoder.Offset(23)
			req.FinancialCreditAmount, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
		case 'T':
			decoder.SetFounded()
			decoder.Offset(10)
			err = decoder.DecodeByteArrayAsBase64(req.ThingID[:])
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
				req.ServicesType = datastore.OrganizationAuthenticationType(num)
			case 't':
				decoder.SetFounded()
				decoder.Offset(8)
				var num uint8
				num, err = decoder.DecodeUInt8()
				if err != nil {
					return
				}
				req.Status = datastore.OrganizationAuthenticationStatus(num)
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *updateOrganizationReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"LeaderPersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.LeaderPersonID[:])

	encoder.EncodeString(`","ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Name":"`)
	encoder.EncodeString(req.Name)

	encoder.EncodeString(`","Domain":"`)
	encoder.EncodeString(req.Domain)

	encoder.EncodeString(`","FinancialCreditAmount":`)
	encoder.EncodeInt64(req.FinancialCreditAmount)

	encoder.EncodeString(`,"ThingID":"`)
	encoder.EncodeByteSliceAsBase64(req.ThingID[:])

	encoder.EncodeString(`","ServicesType":`)
	encoder.EncodeUInt8(uint8(req.ServicesType))

	encoder.EncodeString(`,"Status":`)
	encoder.EncodeUInt8(uint8(req.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *updateOrganizationReq) jsonLen() (ln int) {
	ln = len(req.Name) + len(req.Domain)
	ln += 280
	return
}
