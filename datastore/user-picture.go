/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/picture"
	"../libgo/syllab"
)

const (
	userPictureStructureID uint64 = 9588981481850124477
)

var userPictureStructure = ganjine.DataStructure{
	ID:                9588981481850124477,
	IssueDate:         1599023751,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         UserPicture{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "UserPicture",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store any user type e.g. person,org,... official pictures.",
	},
	TAGS: []string{
		"",
	},
}

// UserPicture ---Read locale description in userPictureStructure---
type UserPicture struct {
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
	ObjectID         [32]byte // UUID of picture object.
	Rating           picture.Rating
	Status           userPictureStatus
}

type userPictureStatus uint8

// UserPicture status
const (
	UserPictureRegister userPictureStatus = iota
	UserPictureRemove
	UserPictureBlockByJustice
)

// Set method set some data and write entire UserPicture record!
func (up *UserPicture) Set() (err error) {
	up.RecordStructureID = userPictureStructureID
	up.RecordSize = up.syllabLen()
	up.WriteTime = etime.Now()
	up.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: up.syllabEncoder(),
	}
	up.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], up.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (up *UserPicture) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: up.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = up.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if up.RecordStructureID != userPictureStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByUserID method find and read last version of record by given UserID
func (up *UserPicture) GetByUserID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: up.HashUserID(),
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
		up.RecordID = indexRes.RecordIDs[ln]
		err = up.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexUserID index up.UserID to retrieve record fast later.
func (up *UserPicture) IndexUserID() {
	var userIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: up.HashUserID(),
		RecordID:  up.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &userIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashUserID hash userPictureStructureID + up.UserID
func (up *UserPicture) HashUserID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, userPictureStructureID)
	copy(buf[8:], up.UserID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (up *UserPicture) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < up.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(up.RecordID[:], buf[0:])
	up.RecordStructureID = syllab.GetUInt64(buf, 32)
	up.RecordSize = syllab.GetUInt64(buf, 40)
	up.WriteTime = syllab.GetInt64(buf, 48)
	copy(up.OwnerAppID[:], buf[56:])

	copy(up.AppInstanceID[:], buf[72:])
	copy(up.UserConnectionID[:], buf[88:])
	copy(up.UserID[:], buf[104:])
	copy(up.ObjectID[:], buf[120:])
	up.Rating = picture.Rating(syllab.GetInt8(buf, 152))
	up.Status = userPictureStatus(syllab.GetInt8(buf, 153))
	return
}

func (up *UserPicture) syllabEncoder() (buf []byte) {
	buf = make([]byte, up.syllabLen())

	// copy(buf[0:], up.RecordID[:])
	syllab.SetUInt64(buf, 32, up.RecordStructureID)
	syllab.SetUInt64(buf, 40, up.RecordSize)
	syllab.SetInt64(buf, 48, up.WriteTime)
	copy(buf[56:], up.OwnerAppID[:])

	copy(buf[72:], up.AppInstanceID[:])
	copy(buf[88:], up.UserConnectionID[:])
	copy(buf[104:], up.UserID[:])
	copy(buf[120:], up.ObjectID[:])
	syllab.SetUInt8(buf, 152, uint8(up.Rating))
	syllab.SetUInt8(buf, 153, uint8(up.Status))
	return
}

func (up *UserPicture) syllabStackLen() (ln uint32) {
	return 154 // fixed size data + variables data add&&len
}

func (up *UserPicture) syllabHeapLen() (ln uint32) {
	return
}

func (up *UserPicture) syllabLen() (ln uint64) {
	return uint64(up.syllabStackLen() + up.syllabHeapLen())
}
