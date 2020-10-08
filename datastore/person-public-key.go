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
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which person connection set||chanaged this record!
	PersonID         [16]byte `ganjine:"Immutable"` // UUID of Person
	ThingID          [16]byte `ganjine:"Immutable"`
	PublicKey        [32]byte `ganjine:"Immutable,Unique"` // Use new algorithm like ECC(256bit) instead of RSA(4096bit)
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
func (ppk *PersonPublicKey) Set() (err error) {
	ppk.RecordStructureID = personPublicKeyStructureID
	ppk.RecordSize = ppk.syllabLen()
	ppk.WriteTime = etime.Now()
	ppk.OwnerAppID = server.Manifest.AppID

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
func (ppk *PersonPublicKey) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: ppk.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = ppk.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if ppk.RecordStructureID != personPublicKeyStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastByPublicKey method find and read last version of record by given PublicKey
func (ppk *PersonPublicKey) GetLastByPublicKey() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: ppk.HashPublicKey(),
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
		ppk.RecordID = indexRes.RecordIDs[ln]
		err = ppk.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexPublicKey index Unique-Field(PublicKey) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (ppk *PersonPublicKey) IndexPublicKey() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.HashPublicKey(),
		RecordID:  ppk.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashPublicKey hash personPublicKeyStructureID + ppk.PublicKey
func (ppk *PersonPublicKey) HashPublicKey() (hash [32]byte) {
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
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.HashPersonID(),
		RecordID:  ppk.PublicKey,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashPersonID hash personPublicKeyStructureID + PersonID
func (ppk *PersonPublicKey) HashPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.PersonID[:])
	return sha512.Sum512_256(buf)
}

// IndexThing index to retrieve Unique-Field(PublicKey) owned by given ThingID
// Don't call in update to an exiting record!
func (ppk *PersonPublicKey) IndexThing() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.HashThingID(),
		RecordID:  ppk.PublicKey,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashThingID hash personPublicKeyStructureID + ThingID
func (ppk *PersonPublicKey) HashThingID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.ThingID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListPersonThing store all ThingID own by specific Person.
// Don't call in update to an exiting record!
func (ppk *PersonPublicKey) ListPersonThing() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.HashPersonThingField(),
	}
	copy(indexRequest.RecordID[:], ppk.ThingID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashPersonThingField hash personPublicKeyStructureID + PersonID + "Thing" field
func (ppk *PersonPublicKey) HashPersonThingField() (hash [32]byte) {
	const field = "Thing"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, personPublicKeyStructureID)
	copy(buf[8:], ppk.PersonID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (ppk *PersonPublicKey) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < ppk.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(ppk.RecordID[:], buf[0:])
	ppk.RecordStructureID = syllab.GetUInt64(buf, 32)
	ppk.RecordSize = syllab.GetUInt64(buf, 40)
	ppk.WriteTime = syllab.GetInt64(buf, 48)
	copy(ppk.OwnerAppID[:], buf[56:])

	copy(ppk.AppInstanceID[:], buf[72:])
	copy(ppk.UserConnectionID[:], buf[88:])
	copy(ppk.PersonID[:], buf[104:])
	copy(ppk.ThingID[:], buf[120:])
	copy(ppk.PublicKey[:], buf[136:])
	ppk.Status = PersonPublicKeyStatus(syllab.GetUInt8(buf, 168))
	return
}

func (ppk *PersonPublicKey) syllabEncoder() (buf []byte) {
	buf = make([]byte, ppk.syllabLen())

	// copy(buf[0:], ppk.RecordID[:])
	syllab.SetUInt64(buf, 32, ppk.RecordStructureID)
	syllab.SetUInt64(buf, 40, ppk.RecordSize)
	syllab.SetInt64(buf, 48, ppk.WriteTime)
	copy(buf[56:], ppk.OwnerAppID[:])

	copy(buf[72:], ppk.AppInstanceID[:])
	copy(buf[88:], ppk.UserConnectionID[:])
	copy(buf[104:], ppk.PersonID[:])
	copy(buf[120:], ppk.ThingID[:])
	copy(buf[136:], ppk.PublicKey[:])
	syllab.SetUInt8(buf, 168, uint8(ppk.Status))
	return
}

func (ppk *PersonPublicKey) syllabStackLen() (ln uint32) {
	return 169 // fixed size data + variables data add&&len
}

func (ppk *PersonPublicKey) syllabHeapLen() (ln uint32) {
	return
}

func (ppk *PersonPublicKey) syllabLen() (ln uint64) {
	return uint64(ppk.syllabStackLen() + ppk.syllabHeapLen())
}
