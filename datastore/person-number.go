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
	personNumberStructureID uint64 = 1212190932488392076
)

var personNumberStructure = ganjine.DataStructure{
	ID:                1212190932488392076,
	IssueDate:         1599048951,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         PersonNumber{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "PersonNumber",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store user number that act for some process like exiting phone, mobile, ...",
	},
	TAGS: []string{
		"",
	},
}

// PersonNumber ---Read locale description in personNumberStructure---
type PersonNumber struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	PersonID         [16]byte `ganjine:"Immutable,Unique"`
	Number           uint64   `ganjine:"Unique"` // must start with country code e.g. (00)98-912-345-6789
	Status           PersonNumberStatus
}

// PersonNumberStatus indicate PersonNumber record status
type PersonNumberStatus uint8

// PersonNumber status
const (
	PersonNumberRegister PersonNumberStatus = iota
	PersonNumberRemove
	PersonNumberBlockByJustice
)

// Set method set some data and write entire PersonNumber record!
func (pn *PersonNumber) Set() (err error) {
	pn.RecordStructureID = personNumberStructureID
	pn.RecordSize = pn.syllabLen()
	pn.WriteTime = etime.Now()
	pn.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pn.syllabEncoder(),
	}
	pn.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pn.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pn *PersonNumber) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: pn.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = pn.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if pn.RecordStructureID != personNumberStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByPersonID method find and read last version of record by given PersonID
func (pn *PersonNumber) GetByPersonID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pn.HashPersonID(),
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
		pn.RecordID = indexRes.RecordIDs[ln]
		err = pn.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

// GetByNumber method find and read last version of record by given Number
func (pn *PersonNumber) GetByNumber() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pn.HashNumber(),
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
		pn.RecordID = indexRes.RecordIDs[ln]
		err = pn.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexPersonID index pn.PersonID to retrieve record fast later.
func (pn *PersonNumber) IndexPersonID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pn.HashPersonID(),
		RecordID:  pn.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashPersonID hash personNumberStructureID + pn.PersonID
func (pn *PersonNumber) HashPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, personNumberStructureID)
	copy(buf[8:], pn.PersonID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexNumber index pn.Number to retrieve record fast later.
func (pn *PersonNumber) IndexNumber() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pn.HashNumber(),
	}
	copy(indexRequest.RecordID[:], pn.PersonID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashNumber hash personNumberStructureID + pn.Number
func (pn *PersonNumber) HashNumber() (hash [32]byte) {
	var buf = make([]byte, 16) // 8+8
	syllab.SetUInt64(buf, 0, personNumberStructureID)
	syllab.SetUInt64(buf, 8, pn.Number)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (pn *PersonNumber) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < pn.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(pn.RecordID[:], buf[0:])
	pn.RecordStructureID = syllab.GetUInt64(buf, 32)
	pn.RecordSize = syllab.GetUInt64(buf, 40)
	pn.WriteTime = syllab.GetInt64(buf, 48)
	copy(pn.OwnerAppID[:], buf[56:])

	copy(pn.AppInstanceID[:], buf[72:])
	copy(pn.UserConnectionID[:], buf[88:])
	copy(pn.PersonID[:], buf[104:])
	pn.Number = syllab.GetUInt64(buf, 120)
	pn.Status = PersonNumberStatus(syllab.GetUInt8(buf, 128))
	return
}

func (pn *PersonNumber) syllabEncoder() (buf []byte) {
	buf = make([]byte, pn.syllabLen())

	// copy(buf[0:], pn.RecordID[:])
	syllab.SetUInt64(buf, 32, pn.RecordStructureID)
	syllab.SetUInt64(buf, 40, pn.RecordSize)
	syllab.SetInt64(buf, 48, pn.WriteTime)
	copy(buf[56:], pn.OwnerAppID[:])

	copy(buf[72:], pn.AppInstanceID[:])
	copy(buf[88:], pn.UserConnectionID[:])
	copy(buf[104:], pn.PersonID[:])
	syllab.SetUInt64(buf, 120, pn.Number)
	syllab.SetUInt8(buf, 128, uint8(pn.Status))
	return
}

func (pn *PersonNumber) syllabStackLen() (ln uint32) {
	return 129 // fixed size data + variables data add&&len
}

func (pn *PersonNumber) syllabHeapLen() (ln uint32) {
	return
}

func (pn *PersonNumber) syllabLen() (ln uint64) {
	return uint64(pn.syllabStackLen() + pn.syllabHeapLen())
}
