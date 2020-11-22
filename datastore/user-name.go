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
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	UserID           [32]byte `ganjine:"Unique" hash-index:"RecordID"`
	Username         string   `hash-index:"UserID"` // It is not replace of user ID! It usually use to find user by their friends!
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
func (un *UserName) Set() (err *er.Error) {
	un.RecordStructureID = userNameStructureID
	un.RecordSize = un.syllabLen()
	un.WriteTime = etime.Now()
	un.OwnerAppID = server.AppID

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
func (un *UserName) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          un.RecordID,
		RecordStructureID: userNameStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = un.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if un.RecordStructureID != userNameStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByUserID method find and read last version of record by given UserID
func (un *UserName) GetLastByUserID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: un.hashUserIDfoRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	un.RecordID = indexRes.IndexValues[0]
	err = un.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", userNameStructureID)
	}
	return
}

// GetLastByUserName method find and read last version of record by given UserName
func (un *UserName) GetLastByUserName() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: un.hashUserNameforUserID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	un.UserID = indexRes.IndexValues[0]
	err = un.GetLastByUserID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", userNameStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexUserID index Unique-Field(UserID) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (un *UserName) IndexUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   un.hashUserIDfoRecordID(),
		IndexValue: un.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (un *UserName) hashUserIDfoRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
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
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   un.hashUserNameforUserID(),
		IndexValue: un.UserID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (un *UserName) hashUserNameforUserID() (hash [32]byte) {
	var buf = make([]byte, 8+len(un.Username))
	syllab.SetUInt64(buf, 0, userNameStructureID)
	copy(buf[8:], un.Username)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (un *UserName) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < un.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(un.RecordID[:], buf[0:])
	un.RecordStructureID = syllab.GetUInt64(buf, 32)
	un.RecordSize = syllab.GetUInt64(buf, 40)
	un.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(un.OwnerAppID[:], buf[56:])

	copy(un.AppInstanceID[:], buf[88:])
	copy(un.UserConnectionID[:], buf[120:])
	copy(un.UserID[:], buf[152:])
	un.Username = syllab.UnsafeGetString(buf, 184)
	un.Status = UserNameStatus(syllab.GetUInt8(buf, 192))
	return
}

func (un *UserName) syllabEncoder() (buf []byte) {
	buf = make([]byte, un.syllabLen())
	var hsi uint32 = un.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], un.RecordID[:])
	syllab.SetUInt64(buf, 32, un.RecordStructureID)
	syllab.SetUInt64(buf, 40, un.RecordSize)
	syllab.SetInt64(buf, 48, int64(un.WriteTime))
	copy(buf[56:], un.OwnerAppID[:])

	copy(buf[88:], un.AppInstanceID[:])
	copy(buf[120:], un.UserConnectionID[:])
	copy(buf[152:], un.UserID[:])
	hsi = syllab.SetString(buf, un.Username, 184, hsi)
	syllab.SetUInt8(buf, 192, uint8(un.Status))
	return
}

func (un *UserName) syllabStackLen() (ln uint32) {
	return 193
}

func (un *UserName) syllabHeapLen() (ln uint32) {
	ln += uint32(len(un.Username))
	return
}

func (un *UserName) syllabLen() (ln uint64) {
	return uint64(un.syllabStackLen() + un.syllabHeapLen())
}
