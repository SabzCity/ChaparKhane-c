/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/log"
	"../libgo/syllab"
)

const (
	organizationAuthenticationStructureID uint64 = 17647865025269007914
)

var organizationAuthenticationStructure = ganjine.DataStructure{
	ID:                17647865025269007914,
	IssueDate:         1600109379,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         OrganizationAuthentication{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Organization Authentication",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: `store organization information!
Organization doesn't have any authenticate token and users have access by UserAppConnection.
Org can service distribution center that not just warehouses but any DC type like stores that do many more things like package multi product to send!`,
	},
	TAGS: []string{
		"Organization", "Authentication", "Distribution Center",
	},
}

// OrganizationAuthentication ---Read locale description in organizationAuthenticationStructure---
type OrganizationAuthentication struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time `hash-index:"ID[daily]"`
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID         [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID      [32]byte // Store to remember which user connection set||chanaged this record!
	ID                    [32]byte `ganjine:"Unique" hash-index:"RecordID"` // Organization UUID
	SocietyID             uint32
	Name                  string `hash-index:"ID"`
	Domain                string `hash-index:"ID"`
	FinancialCreditAmount int64
	ThingID               [32]byte // To get more data like map of DC, ...
	ServicesType          OrganizationAuthenticationType
	Status                OrganizationAuthenticationStatus
}

// SaveNew method set some data and write entire OrganizationAuthentication record with all indexes!
func (oa *OrganizationAuthentication) SaveNew() (err *er.Error) {
	err = oa.Set()
	if err != nil {
		return
	}
	oa.HashIndexRecordIDForID()
	oa.HashIndexIDForRegisterTimeDaily()
	oa.HashIndexIDForName()
	if oa.Domain != "" {
		oa.HashIndexIDForDomain()
	}
	return
}

// Set method set some data and write entire OrganizationAuthentication record!
func (oa *OrganizationAuthentication) Set() (err *er.Error) {
	oa.RecordStructureID = organizationAuthenticationStructureID
	oa.RecordSize = oa.syllabLen()
	oa.WriteTime = etime.Now()
	oa.OwnerAppID = server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: oa.syllabEncoder(),
	}
	oa.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], oa.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
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
	res, err = gsdk.GetRecord(cluster, &req)
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

// GetLastByIDByHashIndex method read all existing record data by given RecordID!
func (oa *OrganizationAuthentication) GetLastByIDByHashIndex() (err *er.Error) {
	var IDs [][32]byte
	IDs, err = oa.GetRecordsIDByIDByHashIndex(18446744073709551615, 1)
	if err != nil || IDs == nil {
		return
	}

	oa.RecordID = IDs[0]
	err = oa.GetByRecordID()
	return
}

// GetRecordsIDByIDByHashIndex find RecordsID by given ID
func (oa *OrganizationAuthentication) GetRecordsIDByIDByHashIndex(offset, limit uint64) (RecordsID [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: oa.hashIDForRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	RecordsID = indexRes.IndexValues
	return
}

// GetLastIDsByHashIndex return org IDs register in platform in given dayNum before given WriteTime.
func (oa *OrganizationAuthentication) GetLastIDsByHashIndex(offset, limit uint64, dayNum int) (RecordsID [][32]byte, err *er.Error) {
	RecordsID = make([][32]byte, 0, limit)

	for i := 0; i < dayNum; i++ {
		var indexReq = &gs.HashIndexGetValuesReq{
			IndexKey: oa.hashWriteTimeForIDDaily(),
			Offset:   offset,
			Limit:    limit,
		}
		var indexRes *gs.HashIndexGetValuesRes
		indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
		RecordsID = append(RecordsID, indexRes.IndexValues...)

		if len(RecordsID) >= int(limit) {
			break
		}

		limit -= uint64(len(indexRes.IndexValues))
		oa.WriteTime -= (24 * 60 * 60)
	}
	return
}

// GetIDsByNameByHashIndex find IDs by given org name
func (oa *OrganizationAuthentication) GetIDsByNameByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: oa.hashNameForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// GetIDsByDomainByHashIndex find IDs by given Domain
func (oa *OrganizationAuthentication) GetIDsByDomainByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: oa.hashDomainForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

/*
	-- PRIMARY INDEXES --
*/

// HashIndexRecordIDForID save RecordID chain for oa.ID
func (oa *OrganizationAuthentication) HashIndexRecordIDForID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashIDForRecordID(),
		IndexValue: oa.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
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

// HashIndexIDForRegisterTimeDaily index oa.WriteTime to retrieve all register Organizations on specific time in daily rate.
// Each year is 365 day that indicate we have 365 index record each year!
func (oa *OrganizationAuthentication) HashIndexIDForRegisterTimeDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashWriteTimeForIDDaily(),
		IndexValue: oa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
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

// HashIndexIDForName save ID chain for Name.
func (oa *OrganizationAuthentication) HashIndexIDForName() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashNameForID(),
		IndexValue: oa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (oa *OrganizationAuthentication) hashNameForID() (hash [32]byte) {
	const field = "Name"
	var buf = make([]byte, 8+len(field)+len(oa.Name))
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], field)
	copy(buf[8+len(field):], oa.Name)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForDomain index oa.Domain to retrieve record fast later.
func (oa *OrganizationAuthentication) HashIndexIDForDomain() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashDomainForID(),
		IndexValue: oa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (oa *OrganizationAuthentication) hashDomainForID() (hash [32]byte) {
	const field = "Domain"
	var buf = make([]byte, 8+len(field)+len(oa.Domain))
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], field)
	copy(buf[8+len(field):], oa.Domain)
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
	oa.SocietyID = syllab.GetUInt32(buf, 184)
	oa.Name = syllab.UnsafeGetString(buf, 188)
	oa.Domain = syllab.UnsafeGetString(buf, 196)
	oa.FinancialCreditAmount = syllab.GetInt64(buf, 204)
	copy(oa.ThingID[:], buf[212:])
	oa.ServicesType = OrganizationAuthenticationType(syllab.GetUInt8(buf, 244))
	oa.Status = OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 245))
	return
}

func (oa *OrganizationAuthentication) syllabEncoder() (buf []byte) {
	buf = make([]byte, oa.syllabLen())
	var hsi uint32 = oa.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], oa.RecordID[:])
	syllab.SetUInt64(buf, 32, oa.RecordStructureID)
	syllab.SetUInt64(buf, 40, oa.RecordSize)
	syllab.SetInt64(buf, 48, int64(oa.WriteTime))
	copy(buf[56:], oa.OwnerAppID[:])

	copy(buf[88:], oa.AppInstanceID[:])
	copy(buf[120:], oa.UserConnectionID[:])
	copy(buf[152:], oa.ID[:])
	syllab.SetUInt32(buf, 184, oa.SocietyID)
	hsi = syllab.SetString(buf, oa.Name, 188, hsi)
	hsi = syllab.SetString(buf, oa.Domain, 196, hsi)
	syllab.SetInt64(buf, 204, oa.FinancialCreditAmount)
	copy(buf[212:], oa.ThingID[:])
	syllab.SetUInt8(buf, 244, uint8(oa.ServicesType))
	syllab.SetUInt8(buf, 245, uint8(oa.Status))
	return
}

func (oa *OrganizationAuthentication) syllabStackLen() (ln uint32) {
	return 246
}

func (oa *OrganizationAuthentication) syllabHeapLen() (ln uint32) {
	ln += uint32(len(oa.Name))
	ln += uint32(len(oa.Domain))
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
func (oas OrganizationAuthenticationType) GetShortDetailByID(language lang.Language) (short string) {
	switch language {
	case lang.EnglishLanguage:
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
func (oas OrganizationAuthenticationType) GetLongDetailByID(language lang.Language) (long string) {
	switch language {
	case lang.EnglishLanguage:
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
	OrganizationStatusTransfred
	OrganizationStatusClosed
	OrganizationStatusBlocked
)

// GetShortDetailByID returns localize short and long details for given status code
func (oas OrganizationAuthenticationStatus) GetShortDetailByID(language lang.Language) (short string) {
	switch language {
	case lang.EnglishLanguage:
		switch oas {
		case OrganizationStatusUnset:
			return "Unset"
		case OrganizationStatusRegister:
			return "Register"
		case OrganizationStatusTransfred:
			return "Transfred"
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
func (oas OrganizationAuthenticationStatus) GetLongDetailByID(language lang.Language) (long string) {
	switch language {
	case lang.EnglishLanguage:
		switch oas {
		case OrganizationStatusUnset:
			return "Record status didn't set yet!"
		case OrganizationStatusRegister:
			return ""
		case OrganizationStatusTransfred:
			return "Org transfred from other society."
		case OrganizationStatusClosed:
			return "Org closed and don't have any activity"
		case OrganizationStatusBlocked:
			return "Organization had been inactive and can't be use now!"
		default:
			return "Given status code is not valid for this type of record"
		}
	}
	return
}
