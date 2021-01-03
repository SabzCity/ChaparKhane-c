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
	personNumberStructureID uint64 = 18356608896785246637
)

var personNumberStructure = ganjine.DataStructure{
	ID:                18356608896785246637,
	IssueDate:         1599048951,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         PersonNumber{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Person Number",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "store user number that act for some process like exiting phone, mobile, ...",
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
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	PersonID         [32]byte `index-hash:"RecordID"`
	Number           uint64   `index-hash:"Number"` // must start with country code e.g. (00)98-912-345-6789
	Status           PersonNumberStatus
}

// SaveNew method set some data and write entire Quiddity record with all indexes!
func (pn *PersonNumber) SaveNew() (err *er.Error) {
	err = pn.Set()
	if err != nil {
		return
	}
	pn.IndexRecordIDForPersonID()
	pn.IndexPersonIDForNumber()
	return
}

// Set method set some data and write entire PersonNumber record!
func (pn *PersonNumber) Set() (err *er.Error) {
	pn.RecordStructureID = personNumberStructureID
	pn.RecordSize = pn.syllabLen()
	pn.WriteTime = etime.Now()
	pn.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pn.syllabEncoder(),
	}
	pn.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pn.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Record:", err)
		}
		// TODO::: Handle error situation
	}
	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pn *PersonNumber) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          pn.RecordID,
		RecordStructureID: personNumberStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = pn.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if pn.RecordStructureID != personNumberStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByPersonID method find and read last version of record by given PersonID
func (pn *PersonNumber) GetLastByPersonID() (err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pn.hashPersonIDForRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	if err != nil {
		return
	}

	pn.RecordID = indexRes.IndexValues[0]
	err = pn.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", personNumberStructureID)
	}
	return
}

// GetLastByNumber method find and read last version of record by given Number
func (pn *PersonNumber) GetLastByNumber() (err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pn.hashNumberForPersonID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	if err != nil {
		return
	}

	pn.PersonID = indexRes.IndexValues[0]
	err = pn.GetLastByPersonID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", personNumberStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexRecordIDForPersonID save RecordID chain for PersonID
func (pn *PersonNumber) IndexRecordIDForPersonID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pn.hashPersonIDForRecordID(),
		IndexValue: pn.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index:", err)
		}
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pn *PersonNumber) hashPersonIDForRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, personNumberStructureID)
	copy(buf[8:], pn.PersonID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexPersonIDForNumber save PersonID chain for Number
func (pn *PersonNumber) IndexPersonIDForNumber() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pn.hashNumberForPersonID(),
		IndexValue: pn.PersonID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index:", err)
		}
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pn *PersonNumber) hashNumberForPersonID() (hash [32]byte) {
	var buf = make([]byte, 16) // 8+8
	syllab.SetUInt64(buf, 0, personNumberStructureID)
	syllab.SetUInt64(buf, 8, pn.Number)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (pn *PersonNumber) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < pn.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(pn.RecordID[:], buf[0:])
	pn.RecordStructureID = syllab.GetUInt64(buf, 32)
	pn.RecordSize = syllab.GetUInt64(buf, 40)
	pn.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(pn.OwnerAppID[:], buf[56:])

	copy(pn.AppInstanceID[:], buf[88:])
	copy(pn.UserConnectionID[:], buf[120:])
	copy(pn.PersonID[:], buf[152:])
	pn.Number = syllab.GetUInt64(buf, 184)
	pn.Status = PersonNumberStatus(syllab.GetUInt8(buf, 192))
	return
}

func (pn *PersonNumber) syllabEncoder() (buf []byte) {
	buf = make([]byte, pn.syllabLen())

	// copy(buf[0:], pn.RecordID[:])
	syllab.SetUInt64(buf, 32, pn.RecordStructureID)
	syllab.SetUInt64(buf, 40, pn.RecordSize)
	syllab.SetInt64(buf, 48, int64(pn.WriteTime))
	copy(buf[56:], pn.OwnerAppID[:])

	copy(buf[88:], pn.AppInstanceID[:])
	copy(buf[120:], pn.UserConnectionID[:])
	copy(buf[152:], pn.PersonID[:])
	syllab.SetUInt64(buf, 184, pn.Number)
	syllab.SetUInt8(buf, 192, uint8(pn.Status))
	return
}

func (pn *PersonNumber) syllabStackLen() (ln uint32) {
	return 193
}

func (pn *PersonNumber) syllabHeapLen() (ln uint32) {
	return
}

func (pn *PersonNumber) syllabLen() (ln uint64) {
	return uint64(pn.syllabStackLen() + pn.syllabHeapLen())
}

/*
	-- Record types --
*/

// PersonNumberStatus indicate PersonNumber record status
type PersonNumberStatus uint8

// PersonNumber status
const (
	PersonNumberUnset PersonNumberStatus = iota
	PersonNumberRegister
	PersonNumberRemove
	PersonNumberBlockedByJustice
)
