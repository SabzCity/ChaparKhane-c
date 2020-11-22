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
	personPublicKeyStructureID uint64 = 13183953152561975962
)

var personPublicKeyStructure = ganjine.DataStructure{
	ID:                13183953152561975962,
	IssueDate:         1599027351,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         PersonPublicKey{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "PersonPublicKey",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store person Public-key (asymmetric) cryptography!",
	},
	TAGS: []string{
		"",
	},
}

// PersonPublicKey ---Read locale description in personPublicKeyStructure---
type PersonPublicKey struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which person connection set||chanaged this record!
	PersonID         [32]byte `ganjine:"Immutable"` // UUID of Person
	ThingID          [32]byte `ganjine:"Immutable" ganjine-list:"PersonID"`
	PublicKey        [32]byte `ganjine:"Immutable,Unique" ganjine-index:"PersonID,ThingID"` // Use new algorithm like ECC(256bit) instead of RSA(4096bit)
	Status           PersonPublicKeyStatus
}

// PersonPublicKeyStatus use to indicate PersonPublicKey record status
type PersonPublicKeyStatus uint8

// PersonPublicKey status
const (
	PersonPublicKeyIssueByPassword PersonPublicKeyStatus = iota
	PersonPublicKeyIssueByPasswordAndOTP
	PersonPublicKeyIssueByPasswordAndIdentification
	PersonPublicKeyIssueByPasswordAndOTPAndIdentification
	PersonPublicKeyExpired
	PersonPublicKeyRevoked
)

// Set method set some data and write entire PersonPublicKey record!
func (ppk *PersonPublicKey) Set() (err *er.Error) {
	ppk.RecordStructureID = personPublicKeyStructureID
	ppk.RecordSize = ppk.syllabLen()
	ppk.WriteTime = etime.Now()
	ppk.OwnerAppID = server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: ppk.syllabEncoder(),
	}
	ppk.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], ppk.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (ppk *PersonPublicKey) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          ppk.RecordID,
		RecordStructureID: personPublicKeyStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = ppk.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if ppk.RecordStructureID != personPublicKeyStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByPublicKey method find and read last version of record by given PublicKey
func (ppk *PersonPublicKey) GetLastByPublicKey() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: ppk.hashPublicKeyforRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	ppk.RecordID = indexRes.IndexValues[0]
	err = ppk.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", personPublicKeyStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexPublicKey index Unique-Field(PublicKey) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (ppk *PersonPublicKey) IndexPublicKey() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   ppk.hashPublicKeyforRecordID(),
		IndexValue: ppk.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ppk *PersonPublicKey) hashPublicKeyforRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.PublicKey[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexPerson index to retrieve Unique-Field(PublicKey) owned by given PersonID
// Don't call in update to an exiting record!
func (ppk *PersonPublicKey) IndexPerson() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   ppk.hashPersonIDforPublicKey(),
		IndexValue: ppk.PublicKey,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ppk *PersonPublicKey) hashPersonIDforPublicKey() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.PersonID[:])
	return sha512.Sum512_256(buf)
}

// IndexThing index to retrieve Unique-Field(PublicKey) owned by given ThingID
// Don't call in update to an exiting record!
func (ppk *PersonPublicKey) IndexThing() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   ppk.hashThingIDforPublicKey(),
		IndexValue: ppk.PublicKey,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ppk *PersonPublicKey) hashThingIDforPublicKey() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.ThingID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListThingIDforPersonID list all ThingID own by specific PersonID.
// Don't call in update to an exiting record!
func (ppk *PersonPublicKey) ListThingIDforPersonID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   ppk.hashPersonIDforThingID(),
		IndexValue: ppk.ThingID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ppk *PersonPublicKey) hashPersonIDforThingID() (hash [32]byte) {
	const field = "ListThingID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.PersonID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (ppk *PersonPublicKey) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < ppk.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(ppk.RecordID[:], buf[0:])
	ppk.RecordStructureID = syllab.GetUInt64(buf, 32)
	ppk.RecordSize = syllab.GetUInt64(buf, 40)
	ppk.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(ppk.OwnerAppID[:], buf[56:])

	copy(ppk.AppInstanceID[:], buf[88:])
	copy(ppk.UserConnectionID[:], buf[120:])
	copy(ppk.PersonID[:], buf[152:])
	copy(ppk.ThingID[:], buf[184:])
	copy(ppk.PublicKey[:], buf[216:])
	ppk.Status = PersonPublicKeyStatus(syllab.GetUInt8(buf, 248))
	return
}

func (ppk *PersonPublicKey) syllabEncoder() (buf []byte) {
	buf = make([]byte, ppk.syllabLen())

	// copy(buf[0:], ppk.RecordID[:])
	syllab.SetUInt64(buf, 32, ppk.RecordStructureID)
	syllab.SetUInt64(buf, 40, ppk.RecordSize)
	syllab.SetInt64(buf, 48, int64(ppk.WriteTime))
	copy(buf[56:], ppk.OwnerAppID[:])

	copy(buf[88:], ppk.AppInstanceID[:])
	copy(buf[120:], ppk.UserConnectionID[:])
	copy(buf[152:], ppk.PersonID[:])
	copy(buf[184:], ppk.ThingID[:])
	copy(buf[216:], ppk.PublicKey[:])
	syllab.SetUInt8(buf, 248, uint8(ppk.Status))
	return
}

func (ppk *PersonPublicKey) syllabStackLen() (ln uint32) {
	return 249
}

func (ppk *PersonPublicKey) syllabHeapLen() (ln uint32) {
	return
}

func (ppk *PersonPublicKey) syllabLen() (ln uint64) {
	return uint64(ppk.syllabStackLen() + ppk.syllabHeapLen())
}
