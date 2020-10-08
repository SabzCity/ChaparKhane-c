/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"
	"unsafe"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/syllab"
)

const (
	userNameStructureID uint64 = 12744998016788909151
)

var userNameStructure = ganjine.DataStructure{
	ID:                12744998016788909151,
	IssueDate:         1599020151,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         UserName{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "UserName",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store user name that translate it to UserID for any purpose like login, send message, ...!",
	},
	TAGS: []string{
		"",
	},
}

// UserName ---Read locale description in userNameStructure---
type UserName struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	UserID           [16]byte `ganjine:"Immutable,Unique"`
	Username         string   `ganjine:"Immutable,Unique"` // It is not replace of user ID! It usually use to find user by their friends!
	Status           UserNameStatus
}

// UserNameStatus indicate UserName record status
type UserNameStatus uint8

// UserName status
const (
	UserNameRegister UserNameStatus = iota
	UserNameRemove
	UserNameBlockByJustice
)

// Set method set some data and write entire UserName record!
func (un *UserName) Set() (err error) {
	un.RecordStructureID = userNameStructureID
	un.RecordSize = un.syllabLen()
	un.WriteTime = etime.Now()
	un.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: un.syllabEncoder(),
	}
	un.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], un.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (un *UserName) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: un.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = un.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if un.RecordStructureID != userNameStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByUserID method find and read last version of record by given UserID
func (un *UserName) GetByUserID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: un.HashUserID(),
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
		un.RecordID = indexRes.RecordIDs[ln]
		err = un.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

// GetByUserName method find and read last version of record by given UserName
func (un *UserName) GetByUserName() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: un.HashUserName(),
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
		un.RecordID = indexRes.RecordIDs[ln]
		err = un.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexUserID index Unique-Field(UserID) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (un *UserName) IndexUserID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: un.HashUserID(),
		RecordID:  un.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashUserID hash userNameStructureID + un.UserID
func (un *UserName) HashUserID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, userNameStructureID)
	copy(buf[8:], un.UserID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexUserName index to retrieve all un.UserID owned by given Username later.
// Don't call in update to an exiting record!
func (un *UserName) IndexUserName() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: un.HashUserName(),
	}
	copy(indexRequest.RecordID[:], un.UserID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashUserName hash userNameStructureID + un.Username
func (un *UserName) HashUserName() (hash [32]byte) {
	var buf = make([]byte, 8+len(un.Username))
	syllab.SetUInt64(buf, 0, userNameStructureID)
	copy(buf[8:], un.Username)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (un *UserName) syllabDecoder(buf []byte) (err error) {
	var add, ln uint32
	var tempSlice []byte

	if uint32(len(buf)) < un.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(un.RecordID[:], buf[0:])
	un.RecordStructureID = syllab.GetUInt64(buf, 32)
	un.RecordSize = syllab.GetUInt64(buf, 40)
	un.WriteTime = syllab.GetInt64(buf, 48)
	copy(un.OwnerAppID[:], buf[56:])

	copy(un.AppInstanceID[:], buf[72:])
	copy(un.UserConnectionID[:], buf[88:])
	copy(un.UserID[:], buf[104:])
	add = syllab.GetUInt32(buf, 120)
	ln = syllab.GetUInt32(buf, 124)
	// It must check len of every heap access but due to encode of data is safe proccess by us, skip it here!
	tempSlice = buf[add : add+ln]
	un.Username = *(*string)(unsafe.Pointer(&tempSlice))
	un.Status = UserNameStatus(syllab.GetUInt8(buf, 128))
	return
}

func (un *UserName) syllabEncoder() (buf []byte) {
	buf = make([]byte, un.syllabLen())
	var hsi uint32 = un.syllabStackLen() // Heap start index || Stack size!
	var ln uint32                        // len of strings, slices, maps, ...

	// copy(buf[0:], un.RecordID[:])
	syllab.SetUInt64(buf, 32, un.RecordStructureID)
	syllab.SetUInt64(buf, 40, un.RecordSize)
	syllab.SetInt64(buf, 48, un.WriteTime)
	copy(buf[56:], un.OwnerAppID[:])

	copy(buf[72:], un.AppInstanceID[:])
	copy(buf[88:], un.UserConnectionID[:])
	copy(buf[104:], un.UserID[:])
	ln = uint32(len(un.Username))
	syllab.SetUInt32(buf, 120, hsi)
	syllab.SetUInt32(buf, 124, ln)
	copy(buf[hsi:], un.Username)
	syllab.SetUInt8(buf, 128, uint8(un.Status))
	return
}

func (un *UserName) syllabStackLen() (ln uint32) {
	return 129 // fixed size data + variables data add&&len
}

func (un *UserName) syllabHeapLen() (ln uint32) {
	ln += uint32(len(un.Username))
	return
}

func (un *UserName) syllabLen() (ln uint64) {
	return uint64(un.syllabStackLen() + un.syllabHeapLen())
}
