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
		lang.EnglishLanguage: "Distribution Center",
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
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	ID               [32]byte `ganjine:"Unique" hash-index:"RecordID"`
	Name             string   `hash-index:"ID"`
	OrgID            [32]byte `hash-index:"ID"`
	Type             uint8    // Private, Public,
	ThingID          [32]byte // To get more data like map of DC, ...
	CoordinateID     [32]byte `hash-index:"ID"`
}

// Set method set some data and write entire DistributionCenter record!
func (dc *DistributionCenter) Set() (err *er.Error) {
	dc.RecordStructureID = distributionCenterStructureID
	dc.RecordSize = dc.syllabLen()
	dc.WriteTime = etime.Now()
	dc.OwnerAppID = server.AppID

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
func (dc *DistributionCenter) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID: dc.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = dc.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if dc.RecordStructureID != distributionCenterStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastByID method find and read last version of record by given ID
func (dc *DistributionCenter) GetLastByID() (err *er.Error) {
	var indexRequest = &gs.HashIndexGetValuesReq{
		IndexKey: dc.hashIDforRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexRequest)
	if err != nil {
		return
	}

	dc.RecordID = indexRes.IndexValues[0]
	err = dc.GetByRecordID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", distributionCenterStructureID)
	}
	return
}

// GetLastByName method find and read last version of record by given Name
func (dc *DistributionCenter) GetLastByName() (err *er.Error) {
	var indexRequest = &gs.HashIndexGetValuesReq{
		IndexKey: dc.hashNameforID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexRequest)
	if err != nil {
		return
	}

	dc.ID = indexRes.IndexValues[0]
	err = dc.GetLastByID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", distributionCenterStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index Unique-Field(ID) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (dc *DistributionCenter) IndexID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   dc.hashIDforRecordID(),
		IndexValue: dc.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (dc *DistributionCenter) hashIDforRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
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
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   dc.hashNameforID(),
		IndexValue: dc.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (dc *DistributionCenter) hashNameforID() (hash [32]byte) {
	var buf = make([]byte, 8+len(dc.Name))
	syllab.SetUInt64(buf, 0, distributionCenterStructureID)
	copy(buf[8:], dc.Name)
	return sha512.Sum512_256(buf)
}

// IndexOrgID index to retrieve all Unique-Field(ID) owned by given OrgID
// Don't call in update to an exiting record!
func (dc *DistributionCenter) IndexOrgID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   dc.hashOrgIDforID(),
		IndexValue: dc.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (dc *DistributionCenter) hashOrgIDforID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, distributionCenterStructureID)
	copy(buf[8:], dc.OrgID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (dc *DistributionCenter) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < dc.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(dc.RecordID[:], buf[0:])
	dc.RecordStructureID = syllab.GetUInt64(buf, 32)
	dc.RecordSize = syllab.GetUInt64(buf, 40)
	dc.WriteTime = syllab.GetInt64(buf, 48)
	copy(dc.OwnerAppID[:], buf[56:])

	copy(dc.AppInstanceID[:], buf[88:])
	copy(dc.UserConnectionID[:], buf[120:])
	copy(dc.ID[:], buf[152:])
	dc.Name = syllab.UnsafeGetString(buf, 184)
	copy(dc.OrgID[:], buf[192:])
	dc.Type = syllab.GetUInt8(buf, 224)
	copy(dc.ThingID[:], buf[225:])
	copy(dc.CoordinateID[:], buf[257:])
	return
}

func (dc *DistributionCenter) syllabEncoder() (buf []byte) {
	buf = make([]byte, dc.syllabLen())
	var hsi uint32 = dc.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], dc.RecordID[:])
	syllab.SetUInt64(buf, 32, dc.RecordStructureID)
	syllab.SetUInt64(buf, 40, dc.RecordSize)
	syllab.SetInt64(buf, 48, dc.WriteTime)
	copy(buf[56:], dc.OwnerAppID[:])

	copy(buf[88:], dc.AppInstanceID[:])
	copy(buf[120:], dc.UserConnectionID[:])
	copy(buf[152:], dc.ID[:])
	syllab.SetString(buf, dc.Name, 184, hsi)
	copy(buf[192:], dc.OrgID[:])
	syllab.SetUInt8(buf, 224, dc.Type)
	copy(buf[225:], dc.ThingID[:])
	copy(buf[257:], dc.CoordinateID[:])
	return
}

func (dc *DistributionCenter) syllabStackLen() (ln uint32) {
	return 289
}

func (dc *DistributionCenter) syllabHeapLen() (ln uint32) {
	ln += uint32(len(dc.Name))
	return
}

func (dc *DistributionCenter) syllabLen() (ln uint64) {
	return uint64(dc.syllabStackLen() + dc.syllabHeapLen())
}
