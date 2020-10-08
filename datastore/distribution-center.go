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
	distributionCenterStructureID uint64 = 465051532317110086
)

var distributionCenterStructure = ganjine.DataStructure{
	ID:                465051532317110086,
	IssueDate:         1599292620,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         DistributionCenter{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "DistributionCenter",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store not just warehouses but any DC type like stores that do many more things like package multi product to send!",
	},
	TAGS: []string{
		"",
	},
}

// DistributionCenter ---Read locale description in userAppsConnectionStructure---
type DistributionCenter struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	ID               [16]byte `ganjine:"Immutable,Unique" ganjine-index:"Name,OrgID"`
	Name             string
	OrgID            [16]byte `ganjine:"Immutable"`
	Type             uint8    // Private, Public,
	ThingID          [16]byte `ganjine:"Immutable"` // To get more data like map of DC, ...
	CoordinateID     [16]byte `ganjine:"Immutable"`
}

// Set method set some data and write entire DistributionCenter record!
func (dc *DistributionCenter) Set() (err error) {
	dc.RecordStructureID = distributionCenterStructureID
	dc.RecordSize = dc.syllabLen()
	dc.WriteTime = etime.Now()
	dc.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: dc.syllabEncoder(),
	}
	dc.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], dc.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (dc *DistributionCenter) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: dc.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = dc.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if dc.RecordStructureID != distributionCenterStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByID method find and read last version of record by given ID
func (dc *DistributionCenter) GetByID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: dc.HashID(),
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
		dc.RecordID = indexRes.RecordIDs[ln]
		err = dc.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

// GetByName method find and read last version of record by given Name
func (dc *DistributionCenter) GetByName() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: dc.HashName(),
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
		dc.RecordID = indexRes.RecordIDs[ln]
		err = dc.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index Unique-Field(ID) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (dc *DistributionCenter) IndexID() {
	var idIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: dc.HashID(),
		RecordID:  dc.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &idIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashID hash distributionCenterStructureID + ID
func (dc *DistributionCenter) HashID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, distributionCenterStructureID)
	copy(buf[8:], dc.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexName index to retrieve all Unique-Field(ID) owned by given Name
// Don't call in update to an exiting record!
func (dc *DistributionCenter) IndexName() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: dc.HashName(),
	}
	copy(indexRequest.RecordID[:], dc.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashName hash distributionCenterStructureID + Name
func (dc *DistributionCenter) HashName() (hash [32]byte) {
	var buf = make([]byte, 8+len(dc.Name))
	syllab.SetUInt64(buf, 0, distributionCenterStructureID)
	copy(buf[8:], dc.Name)
	return sha512.Sum512_256(buf)
}

// IndexOrg index to retrieve all Unique-Field(ID) owned by given OrgID
// Don't call in update to an exiting record!
func (dc *DistributionCenter) IndexOrg() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: dc.HashName(),
	}
	copy(indexRequest.RecordID[:], dc.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOrgID hash distributionCenterStructureID + OrgID
func (dc *DistributionCenter) HashOrgID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, distributionCenterStructureID)
	copy(buf[8:], dc.OrgID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (dc *DistributionCenter) syllabDecoder(buf []byte) (err error) {
	var add, ln uint32

	if uint32(len(buf)) < dc.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(dc.RecordID[:], buf[0:])
	dc.RecordStructureID = syllab.GetUInt64(buf, 32)
	dc.RecordSize = syllab.GetUInt64(buf, 40)
	dc.WriteTime = syllab.GetInt64(buf, 48)
	copy(dc.OwnerAppID[:], buf[56:])

	copy(dc.AppInstanceID[:], buf[72:])
	copy(dc.UserConnectionID[:], buf[88:])
	copy(dc.ID[:], buf[104:])
	add = syllab.GetUInt32(buf, 120)
	ln = syllab.GetUInt32(buf, 124)
	dc.Name = string(buf[add : add+ln])
	copy(dc.OrgID[:], buf[128:])
	dc.Type = syllab.GetUInt8(buf, 144)
	copy(dc.ThingID[:], buf[145:])
	copy(dc.CoordinateID[:], buf[161:])
	return
}

func (dc *DistributionCenter) syllabEncoder() (buf []byte) {
	buf = make([]byte, dc.syllabLen())
	var hsi uint32 = dc.syllabStackLen() // Heap start index || Stack size!
	var ln uint32                        // len of strings, slices, maps, ...

	// copy(buf[0:], dc.RecordID[:])
	syllab.SetUInt64(buf, 32, dc.RecordStructureID)
	syllab.SetUInt64(buf, 40, dc.RecordSize)
	syllab.SetInt64(buf, 48, dc.WriteTime)
	copy(buf[56:], dc.OwnerAppID[:])

	copy(buf[72:], dc.AppInstanceID[:])
	copy(buf[88:], dc.UserConnectionID[:])
	copy(buf[104:], dc.ID[:])
	ln = uint32(len(dc.Name))
	syllab.SetUInt32(buf, 120, hsi)
	syllab.SetUInt32(buf, 124, ln)
	copy(buf[hsi:], dc.Name)
	hsi += ln
	copy(buf[128:], dc.OrgID[:])
	syllab.SetUInt8(buf, 144, dc.Type)
	copy(buf[145:], dc.ThingID[:])
	copy(buf[161:], dc.CoordinateID[:])
	return
}

func (dc *DistributionCenter) syllabStackLen() (ln uint32) {
	return 177 // fixed size data + variables data add&&len
}

func (dc *DistributionCenter) syllabHeapLen() (ln uint32) {
	ln += uint32(len(dc.Name))
	return
}

func (dc *DistributionCenter) syllabLen() (ln uint64) {
	return uint64(dc.syllabStackLen() + dc.syllabHeapLen())
}
