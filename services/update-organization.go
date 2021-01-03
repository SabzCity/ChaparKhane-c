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
		lang.LanguageEnglish: "Update Organization",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
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
	ID           [32]byte `json:",string"`
	SocietyID    [32]byte `json:",string"`
	ServicesType datastore.OrganizationAuthenticationType

	LeaderPersonID [32]byte `json:",string"` // not empty just if need to change leader and make new connection for this person as leader!s
}

func updateOrganization(st *achaemenid.Stream, req *updateOrganizationReq) (err *er.Error) {
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
	err = oa.GetLastByID()
	if err != nil {
		return
	}

	if oa.Status == datastore.OrganizationStatusBlocked {
		err = ErrBlockedByJustice
		return
	}
	if req.SocietyID != oa.SocietyID {
		oa.Status = datastore.OrganizationStatusTransferred
	}

	oa = datastore.OrganizationAuthentication{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               req.ID,
		SocietyID:        req.SocietyID,
		QuiddityID:       oa.QuiddityID,
		ServicesType:     req.ServicesType,
		Status:           oa.Status,
	}
	err = oa.Set()
	if err != nil {
		return
	}
	oa.IndexRecordIDForID()

	if req.LeaderPersonID != [32]byte{} {
		// make new connection for leader
		var uac = datastore.UserAppConnection{
			Status:      datastore.UserAppConnectionIssued,
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

	copy(req.ID[:], buf[0:])
	copy(req.SocietyID[:], buf[32:])
	req.ServicesType = datastore.OrganizationAuthenticationType(syllab.GetUInt8(buf, 64))
	copy(req.LeaderPersonID[:], buf[65:])
	return
}

func (req *updateOrganizationReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	copy(buf[32:], req.SocietyID[:])
	syllab.SetUInt8(buf, 64, uint8(req.ServicesType))
	copy(buf[65:], req.LeaderPersonID[:])
	return
}

func (req *updateOrganizationReq) syllabStackLen() (ln uint32) {
	return 97
}

func (req *updateOrganizationReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *updateOrganizationReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateOrganizationReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		case "SocietyID":
			err = decoder.DecodeByteArrayAsBase64(req.SocietyID[:])
		case "ServicesType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.ServicesType = datastore.OrganizationAuthenticationType(num)
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

func (req *updateOrganizationReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","SocietyID":"`)
	encoder.EncodeByteSliceAsBase64(req.SocietyID[:])

	encoder.EncodeString(`","ServicesType":`)
	encoder.EncodeUInt8(uint8(req.ServicesType))

	encoder.EncodeString(`,"LeaderPersonID":"`)
	encoder.EncodeByteSliceAsBase64(req.LeaderPersonID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *updateOrganizationReq) jsonLen() (ln int) {
	ln = 192
	return
}
