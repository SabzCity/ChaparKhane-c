/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	"../libgo/achaemenid"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/log"
	"../libgo/pehrest"
	psdk "../libgo/pehrest-sdk"
	"../libgo/syllab"
)

const (
	organizationAuthenticationStructureID uint64 = 9250029817569263954
)

var organizationAuthenticationStructure = ganjine.DataStructure{
	ID:                9250029817569263954,
	IssueDate:         1600109379,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         OrganizationAuthentication{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Organization Authentication",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: `store organization information!
Organization doesn't have any authenticate token and users have access by UserAppConnection.
Org can service distribution center that not just warehouses but any DC type like stores that do many more things like package multi product to send!`,
	},
	TAGS: []string{
		"Organization", "Authentication", "DistributionCenter",
	},
}

// OrganizationAuthentication ---Read locale description in organizationAuthenticationStructure---
type OrganizationAuthentication struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time `index-hash:"ID[daily]"`
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	ID               [32]byte `index-hash:"RecordID"` // Organization UUID
	SocietyID        [32]byte
	QuiddityID       [32]byte // To get more data like map of DC, ...
	ServicesType     OrganizationAuthenticationType
	Status           OrganizationAuthenticationStatus
}

// SaveNew method set some data and write entire OrganizationAuthentication record with all indexes!
func (oa *OrganizationAuthentication) SaveNew() (err *er.Error) {
	err = oa.Set()
	if err != nil {
		return
	}
	oa.IndexRecordIDForID()
	oa.IndexIDForRegisterTimeDaily()
	return
}

// Set method set some data and write entire OrganizationAuthentication record!
func (oa *OrganizationAuthentication) Set() (err *er.Error) {
	oa.RecordStructureID = organizationAuthenticationStructureID
	oa.RecordSize = oa.syllabLen()
	oa.WriteTime = etime.Now()
	oa.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: oa.syllabEncoder(),
	}
	oa.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], oa.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		// TODO::: Handle error situation
	}
	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (oa *OrganizationAuthentication) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          oa.RecordID,
		RecordStructureID: organizationAuthenticationStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = oa.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if oa.RecordStructureID != organizationAuthenticationStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", organizationAuthenticationStructureID)
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByID method read all existing record data by given RecordID!
func (oa *OrganizationAuthentication) GetLastByID() (err *er.Error) {
	var IDs [][32]byte
	IDs, err = oa.FindRecordsIDByID(18446744073709551615, 1)
	if err != nil || IDs == nil {
		return
	}

	oa.RecordID = IDs[0]
	err = oa.GetByRecordID()
	return
}

/*
	-- Search Methods --
*/

// FindRecordsIDByID find RecordsID by given ID
func (oa *OrganizationAuthentication) FindRecordsIDByID(offset, limit uint64) (RecordsID [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: oa.hashIDForRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	RecordsID = indexRes.IndexValues
	return
}

// FindLastIDs return org IDs register in platform in given dayNum before given WriteTime.
func (oa *OrganizationAuthentication) FindLastIDs(offset, limit uint64, dayNum int) (RecordsID [][32]byte, err *er.Error) {
	RecordsID = make([][32]byte, 0, limit)

	for i := 0; i < dayNum; i++ {
		var indexReq = &pehrest.HashGetValuesReq{
			IndexKey: oa.hashWriteTimeForIDDaily(),
			Offset:   offset,
			Limit:    limit,
		}
		var indexRes *pehrest.HashGetValuesRes
		indexRes, err = psdk.HashGetValues(indexReq)
		RecordsID = append(RecordsID, indexRes.IndexValues...)

		if len(RecordsID) >= int(limit) {
			break
		}

		limit -= uint64(len(indexRes.IndexValues))
		oa.WriteTime -= (24 * 60 * 60)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexRecordIDForID save RecordID chain for oa.ID
func (oa *OrganizationAuthentication) IndexRecordIDForID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashIDForRecordID(),
		IndexValue: oa.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (oa *OrganizationAuthentication) hashIDForRecordID() (hash [32]byte) {
	const field = "ID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.ID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexIDForRegisterTimeDaily index oa.WriteTime to retrieve all register Organizations on specific time in daily rate.
// Each year is 365 day that indicate we have 365 index record each year!
func (oa *OrganizationAuthentication) IndexIDForRegisterTimeDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashWriteTimeForIDDaily(),
		IndexValue: oa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index:", err)
		}
		// TODO::: we must retry more due to record wrote successfully before this func!
	}
}

func (oa *OrganizationAuthentication) hashWriteTimeForIDDaily() (hash [32]byte) {
	const field = "WriteTime"
	var buf = make([]byte, 16+len(field)) // 8+8
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	syllab.SetInt64(buf, 8, oa.WriteTime.RoundToDay())
	copy(buf[16:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (oa *OrganizationAuthentication) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < oa.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(oa.RecordID[:], buf[0:])
	oa.RecordStructureID = syllab.GetUInt64(buf, 32)
	oa.RecordSize = syllab.GetUInt64(buf, 40)
	oa.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(oa.OwnerAppID[:], buf[56:])

	copy(oa.AppInstanceID[:], buf[88:])
	copy(oa.UserConnectionID[:], buf[120:])
	copy(oa.ID[:], buf[152:])
	copy(oa.SocietyID[:], buf[184:])
	copy(oa.QuiddityID[:], buf[216:])
	oa.ServicesType = OrganizationAuthenticationType(syllab.GetUInt8(buf, 248))
	oa.Status = OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 249))
	return
}

func (oa *OrganizationAuthentication) syllabEncoder() (buf []byte) {
	buf = make([]byte, oa.syllabLen())

	// copy(buf[0:], oa.RecordID[:])
	syllab.SetUInt64(buf, 32, oa.RecordStructureID)
	syllab.SetUInt64(buf, 40, oa.RecordSize)
	syllab.SetInt64(buf, 48, int64(oa.WriteTime))
	copy(buf[56:], oa.OwnerAppID[:])

	copy(buf[88:], oa.AppInstanceID[:])
	copy(buf[120:], oa.UserConnectionID[:])
	copy(buf[152:], oa.ID[:])
	copy(buf[184:], oa.SocietyID[:])
	copy(buf[216:], oa.QuiddityID[:])
	syllab.SetUInt8(buf, 248, uint8(oa.ServicesType))
	syllab.SetUInt8(buf, 249, uint8(oa.Status))
	return
}

func (oa *OrganizationAuthentication) syllabStackLen() (ln uint32) {
	return 250
}

func (oa *OrganizationAuthentication) syllabHeapLen() (ln uint32) {
	return
}

func (oa *OrganizationAuthentication) syllabLen() (ln uint64) {
	return uint64(oa.syllabStackLen() + oa.syllabHeapLen())
}

/*
	-- Record types --
*/

// OrganizationAuthenticationType indicate OrganizationAuthentication type
type OrganizationAuthenticationType uint8

// Organization Authentication type
const (
	OrganizationTypeUnset OrganizationAuthenticationType = iota
	OrganizationTypePrivate
	OrganizationTypePublic
)

// GetShortDetailByID returns localize short and long details for given type code
func (oas OrganizationAuthenticationType) GetShortDetailByID() (short string) {
	switch lang.AppLanguage {
	case lang.LanguageEnglish:
		switch oas {
		case OrganizationTypeUnset:
			return "Unset"
		case OrganizationTypePrivate:
			return "Private"
		case OrganizationTypePublic:
			return "Public"
		default:
			return "Invalid Type"
		}
	}
	return
}

// GetLongDetailByID returns localize long details for given type code
func (oas OrganizationAuthenticationType) GetLongDetailByID() (long string) {
	switch lang.AppLanguage {
	case lang.LanguageEnglish:
		switch oas {
		case OrganizationTypeUnset:
			return "Organization type didn't set yet!"
		case OrganizationTypePrivate:
			return ""
		case OrganizationTypePublic:
			return ""
		default:
			return "Given organization type is not valid"
		}
	}
	return
}

// OrganizationAuthenticationStatus indicate OrganizationAuthentication record status
type OrganizationAuthenticationStatus uint8

// OrganizationAuthentication status
const (
	OrganizationStatusUnset OrganizationAuthenticationStatus = iota
	OrganizationStatusRegister
	OrganizationStatusRepresentative
	OrganizationStatusTransferred
	OrganizationStatusClosed
	OrganizationStatusBlocked
)

// GetShortDetailByID returns localize short and long details for given status code
func (oas OrganizationAuthenticationStatus) GetShortDetailByID() (short string) {
	switch lang.AppLanguage {
	case lang.LanguageEnglish:
		switch oas {
		case OrganizationStatusUnset:
			return "Unset"
		case OrganizationStatusRegister:
			return "Register"
		case OrganizationStatusRepresentative:
			return "Representative"
		case OrganizationStatusTransferred:
			return "Transferred"
		case OrganizationStatusClosed:
			return "Closed"
		case OrganizationStatusBlocked:
			return "Blocked"
		default:
			return "Invalid Status"
		}
	}
	return
}

// GetLongDetailByID returns localize long details for given status code
func (oas OrganizationAuthenticationStatus) GetLongDetailByID() (long string) {
	switch lang.AppLanguage {
	case lang.LanguageEnglish:
		switch oas {
		case OrganizationStatusUnset:
			return "Record status didn't set yet!"
		case OrganizationStatusRegister:
			return "Organization register and start in this society"
		case OrganizationStatusRepresentative:
			return "Organization is just representative of the org that register on other society"
		case OrganizationStatusTransferred:
			return "Organization transfred from other society."
		case OrganizationStatusClosed:
			return "Organization closed and can't & don't have any further activity"
		case OrganizationStatusBlocked:
			return "Organization had been inactive and can't be use now!"
		default:
			return "Given status code is not valid for this type of record"
		}
	}
	return
}
