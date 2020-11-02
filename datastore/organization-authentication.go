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
Organization doesn't have any authenticate token and users have access by UserAppConnection.`,
	},
	TAGS: []string{
		"Authentication",
	},
}

// OrganizationAuthentication ---Read locale description in organizationAuthenticationStructure---
type OrganizationAuthentication struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID         [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID      [32]byte // Store to remember which user connection set||chanaged this record!
	ID                    [32]byte `ganjine:"Immutable,Unique"` // Organization UUID
	Name                  string   `ganjine:"Unique"`
	Domain                string   `ganjine:"Unique"`
	FinancialCreditAmount int64
	Status                OrganizationAuthenticationStatus
}

// OrganizationAuthenticationStatus indicate OrganizationAuthentication record status
type OrganizationAuthenticationStatus uint8

// OrganizationAuthentication status
const (
	OrganizationStatusUnset OrganizationAuthenticationStatus = iota
	OrganizationStatusRegister
	OrganizationStatusInactive // organization had been inactive and can't be use now!
	OrganizationStatusBlocked
	OrganizationStatusIdea
	OrganizationStatusStart
	OrganizationStatusClosed // can't be set to OrganizationProfitWithPhysicalPlace type!

)

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
		RecordID: oa.RecordID,
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
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastByID find and read last version of record by given ID
func (oa *OrganizationAuthentication) GetLastByID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: oa.hashIDforRecord(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	oa.RecordID = indexRes.IndexValues[0]
	err = oa.GetByRecordID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", organizationAuthenticationStructureID)
	}
	return
}

// GetLastByName find and read last version of record by given org name
func (oa *OrganizationAuthentication) GetLastByName() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: oa.hashNameforID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	oa.ID = indexRes.IndexValues[0]
	err = oa.GetLastByID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", organizationAuthenticationStructureID)
	}
	return
}

// GetLastByDomain find and read last version of record by given Domain
func (oa *OrganizationAuthentication) GetLastByDomain() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: oa.hashDomainforID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	oa.ID = indexRes.IndexValues[0]
	err = oa.GetLastByID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", organizationAuthenticationStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index oa.ID to retrieve record fast later.
func (oa *OrganizationAuthentication) IndexID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashIDforRecord(),
		IndexValue: oa.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (oa *OrganizationAuthentication) hashIDforRecord() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexName index oa.Name to retrieve record fast later.
func (oa *OrganizationAuthentication) IndexName() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashNameforID(),
		IndexValue: oa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (oa *OrganizationAuthentication) hashNameforID() (hash [32]byte) {
	var buf = make([]byte, 8+len(oa.Name))
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.Name)
	return sha512.Sum512_256(buf)
}

// IndexDomain index oa.Domain to retrieve record fast later.
func (oa *OrganizationAuthentication) IndexDomain() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   oa.hashDomainforID(),
		IndexValue: oa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (oa *OrganizationAuthentication) hashDomainforID() (hash [32]byte) {
	var buf = make([]byte, 8+len(oa.Domain))
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.Domain)
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
	oa.WriteTime = syllab.GetInt64(buf, 48)
	copy(oa.OwnerAppID[:], buf[56:])

	copy(oa.AppInstanceID[:], buf[88:])
	copy(oa.UserConnectionID[:], buf[120:])
	copy(oa.ID[:], buf[152:])
	oa.Name = syllab.UnsafeGetString(buf, 184)
	oa.Domain = syllab.UnsafeGetString(buf, 192)
	oa.FinancialCreditAmount = syllab.GetInt64(buf, 200)
	oa.Status = OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 208))
	return
}

func (oa *OrganizationAuthentication) syllabEncoder() (buf []byte) {
	buf = make([]byte, oa.syllabLen())
	var hsi uint32 = oa.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], oa.RecordID[:])
	syllab.SetUInt64(buf, 32, oa.RecordStructureID)
	syllab.SetUInt64(buf, 40, oa.RecordSize)
	syllab.SetInt64(buf, 48, oa.WriteTime)
	copy(buf[56:], oa.OwnerAppID[:])

	copy(buf[88:], oa.AppInstanceID[:])
	copy(buf[120:], oa.UserConnectionID[:])
	copy(buf[152:], oa.ID[:])
	hsi = syllab.SetString(buf, oa.Name, 184, hsi)
	syllab.SetString(buf, oa.Domain, 192, hsi)
	syllab.SetInt64(buf, 200, oa.FinancialCreditAmount)
	syllab.SetUInt8(buf, 208, uint8(oa.Status))
	return
}

func (oa *OrganizationAuthentication) syllabStackLen() (ln uint32) {
	return 209
}

func (oa *OrganizationAuthentication) syllabHeapLen() (ln uint32) {
	ln += uint32(len(oa.Name))
	ln += uint32(len(oa.Domain))
	return
}

func (oa *OrganizationAuthentication) syllabLen() (ln uint64) {
	return uint64(oa.syllabStackLen() + oa.syllabHeapLen())
}
