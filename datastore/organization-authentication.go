/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
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
	ExpireInFavorOfID: 0,  // Other ServiceID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         OrganizationAuthentication{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "OrganizationAuthentication",
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
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	ID               [16]byte `ganjine:"Immutable,Unique"` // Organization UUID
	Name             string   `ganjine:"Unique"`
	Domain           string   `ganjine:"Unique"`
	Status           OrganizationAuthenticationStatus
}

// OrganizationAuthenticationStatus indicate OrganizationAuthentication record status
type OrganizationAuthenticationStatus uint8

// OrganizationAuthentication status
const (
	// OrganizationStatusInactive indicate organization had been inactive and can't be use now!
	OrganizationStatusInactive OrganizationAuthenticationStatus = iota
	// OrganizationStatusBlocked
	OrganizationStatusBlocked
	// OrganizationStatusIdea
	OrganizationStatusIdea
	// OrganizationStatusStart
	OrganizationStatusStart
	// OrganizationStatusClosed can't be set to OrganizationProfitWithPhysicalPlace type!
	OrganizationStatusClosed
	// OrganizationStatusRegister
	OrganizationStatusRegister
)

// Set method set some data and write entire OrganizationAuthentication record!
func (oa *OrganizationAuthentication) Set() (err error) {
	oa.RecordStructureID = organizationAuthenticationStructureID
	oa.RecordSize = oa.syllabLen()
	oa.WriteTime = etime.Now()
	oa.OwnerAppID = server.Manifest.AppID

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
func (oa *OrganizationAuthentication) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: oa.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = oa.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if oa.RecordStructureID != organizationAuthenticationStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByID method find and read last version of record by given ID
func (oa *OrganizationAuthentication) GetByID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: oa.HashID(),
		Offset:    18446744073709551615,
		Limit:     0,
	}
	var indexRes *gs.FindRecordsRes
	indexRes, err = gsdk.FindRecords(cluster, indexReq)
	if err != nil {
		return err
	}

	var ln = len(indexRes.RecordIDs)
	// TODO::: Need to handle this here?? if collision ocurred and last record ID is not our purpose??
	ln--
	for ; ln > 0; ln-- {
		oa.RecordID = indexRes.RecordIDs[ln]
		err = oa.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

// GetByDomain method find and read last version of record by given Domain
func (oa *OrganizationAuthentication) GetByDomain() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: oa.HashDomain(),
		Offset:    18446744073709551615,
		Limit:     0,
	}
	var indexRes *gs.FindRecordsRes
	indexRes, err = gsdk.FindRecords(cluster, indexReq)
	if err != nil {
		return err
	}

	var ln = len(indexRes.RecordIDs)
	// TODO::: Need to handle this here?? if collision ocurred and last record ID is not our purpose??
	ln--
	for ; ln > 0; ln-- {
		oa.RecordID = indexRes.RecordIDs[ln]
		err = oa.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index oa.ID to retrieve record fast later.
func (oa *OrganizationAuthentication) IndexID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: oa.HashID(),
		RecordID:  oa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashID hash organizationAuthenticationStructureID + oa.ID
func (oa *OrganizationAuthentication) HashID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexName index oa.Name to retrieve record fast later.
func (oa *OrganizationAuthentication) IndexName() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: oa.HashName(),
	}
	copy(indexRequest.RecordID[:], oa.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashName hash organizationAuthenticationStructureID + oa.Name
func (oa *OrganizationAuthentication) HashName() (hash [32]byte) {
	var buf = make([]byte, 8+len(oa.Name))
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.Name)
	return sha512.Sum512_256(buf)
}

// IndexDomain index oa.Domain to retrieve record fast later.
func (oa *OrganizationAuthentication) IndexDomain() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: oa.HashDomain(),
	}
	copy(indexRequest.RecordID[:], oa.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashDomain hash organizationAuthenticationStructureID + Domain
func (oa *OrganizationAuthentication) HashDomain() (hash [32]byte) {
	var buf = make([]byte, 8+len(oa.Domain))
	syllab.SetUInt64(buf, 0, organizationAuthenticationStructureID)
	copy(buf[8:], oa.Domain)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (oa *OrganizationAuthentication) syllabDecoder(buf []byte) (err error) {
	var add, ln uint32

	if uint32(len(buf)) < oa.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(oa.RecordID[:], buf[0:])
	oa.RecordStructureID = syllab.GetUInt64(buf, 32)
	oa.RecordSize = syllab.GetUInt64(buf, 40)
	oa.WriteTime = syllab.GetInt64(buf, 48)
	copy(oa.OwnerAppID[:], buf[56:])

	copy(oa.AppInstanceID[:], buf[72:])
	copy(oa.UserConnectionID[:], buf[88:])
	copy(oa.ID[:], buf[104:])
	add = syllab.GetUInt32(buf, 120)
	ln = syllab.GetUInt32(buf, 124)
	oa.Name = string(buf[add : add+ln])
	add = syllab.GetUInt32(buf, 128)
	ln = syllab.GetUInt32(buf, 132)
	oa.Domain = string(buf[add : add+ln])
	oa.Status = OrganizationAuthenticationStatus(syllab.GetUInt8(buf, 136))

	return
}

func (oa *OrganizationAuthentication) syllabEncoder() (buf []byte) {
	buf = make([]byte, oa.syllabLen())
	var hsi uint32 = oa.syllabStackLen() // Heap start index || Stack size!
	var ln uint32                        // len of strings, slices, maps, ...

	// copy(buf[0:], oa.RecordID[:])
	syllab.SetUInt64(buf, 32, oa.RecordStructureID)
	syllab.SetUInt64(buf, 40, oa.RecordSize)
	syllab.SetInt64(buf, 48, oa.WriteTime)
	copy(buf[56:], oa.OwnerAppID[:])

	copy(buf[72:], oa.AppInstanceID[:])
	copy(buf[88:], oa.UserConnectionID[:])
	copy(buf[104:], oa.ID[:])
	ln = uint32(len(oa.Name))
	syllab.SetUInt32(buf, 120, hsi)
	syllab.SetUInt32(buf, 124, ln)
	copy(buf[hsi:], oa.Name)
	hsi += ln
	ln = uint32(len(oa.Domain))
	syllab.SetUInt32(buf, 128, hsi)
	syllab.SetUInt32(buf, 132, ln)
	copy(buf[hsi:], oa.Domain)
	hsi += ln
	syllab.SetUInt8(buf, 136, uint8(oa.Status))
	return
}

func (oa *OrganizationAuthentication) syllabStackLen() (ln uint32) {
	return 137 // fixed size data + variables data add&&len
}

func (oa *OrganizationAuthentication) syllabHeapLen() (ln uint32) {
	ln += uint32(len(oa.Name))
	ln += uint32(len(oa.Domain))
	return
}

func (oa *OrganizationAuthentication) syllabLen() (ln uint64) {
	return uint64(oa.syllabStackLen() + oa.syllabHeapLen())
}
