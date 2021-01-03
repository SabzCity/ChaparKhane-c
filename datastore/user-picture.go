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
	"../libgo/picture"
	"../libgo/syllab"
)

const (
	userPictureStructureID uint64 = 14810657980163930313
)

var userPictureStructure = ganjine.DataStructure{
	ID:                14810657980163930313,
	IssueDate:         1599023751,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         UserPicture{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "User Picture",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "store any user type e.g. person,org,... official pictures.",
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
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	UserID           [32]byte `index-hash:"RecordID"`
	ObjectID         [32]byte // UUID of picture object.
	Rating           picture.Rating
	Status           UserPictureStatus
}

// SaveNew method set some data and write entire Quiddity record with all indexes!
func (up *UserPicture) SaveNew() (err *er.Error) {
	err = up.Set()
	if err != nil {
		return
	}
	up.IndexRecordIDForUserID()
	return
}

// Set method set some data and write entire UserPicture record!
func (up *UserPicture) Set() (err *er.Error) {
	up.RecordStructureID = userPictureStructureID
	up.RecordSize = up.syllabLen()
	up.WriteTime = etime.Now()
	up.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: up.syllabEncoder(),
	}
	up.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], up.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (up *UserPicture) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          up.RecordID,
		RecordStructureID: userPictureStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = up.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if up.RecordStructureID != userPictureStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByUserID find and read last version of record by given UserID
func (up *UserPicture) GetLastByUserID() (err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: up.hashUserIDForRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	if err != nil {
		return
	}

	up.RecordID = indexRes.IndexValues[0]
	err = up.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", userPictureStructureID)
	}
	return
}

/*
	-- Search Methods --
*/

/*
	-- PRIMARY INDEXES --
*/

// IndexRecordIDForUserID save RecordID chain for UserID
func (up *UserPicture) IndexRecordIDForUserID() {
	var userIDIndex = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   up.hashUserIDForRecordID(),
		IndexValue: up.RecordID,
	}
	var err = psdk.HashSetValue(&userIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (up *UserPicture) hashUserIDForRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, userPictureStructureID)
	copy(buf[8:], up.UserID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (up *UserPicture) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < up.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(up.RecordID[:], buf[0:])
	up.RecordStructureID = syllab.GetUInt64(buf, 32)
	up.RecordSize = syllab.GetUInt64(buf, 40)
	up.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(up.OwnerAppID[:], buf[56:])

	copy(up.AppInstanceID[:], buf[88:])
	copy(up.UserConnectionID[:], buf[120:])
	copy(up.UserID[:], buf[152:])
	copy(up.ObjectID[:], buf[184:])
	up.Rating = picture.Rating(syllab.GetInt8(buf, 216))
	up.Status = UserPictureStatus(syllab.GetInt8(buf, 217))
	return
}

func (up *UserPicture) syllabEncoder() (buf []byte) {
	buf = make([]byte, up.syllabLen())

	// copy(buf[0:], up.RecordID[:])
	syllab.SetUInt64(buf, 32, up.RecordStructureID)
	syllab.SetUInt64(buf, 40, up.RecordSize)
	syllab.SetInt64(buf, 48, int64(up.WriteTime))
	copy(buf[56:], up.OwnerAppID[:])

	copy(buf[88:], up.AppInstanceID[:])
	copy(buf[120:], up.UserConnectionID[:])
	copy(buf[152:], up.UserID[:])
	copy(buf[184:], up.ObjectID[:])
	syllab.SetUInt8(buf, 216, uint8(up.Rating))
	syllab.SetUInt8(buf, 217, uint8(up.Status))
	return
}

func (up *UserPicture) syllabStackLen() (ln uint32) {
	return 218
}

func (up *UserPicture) syllabHeapLen() (ln uint32) {
	return
}

func (up *UserPicture) syllabLen() (ln uint64) {
	return uint64(up.syllabStackLen() + up.syllabHeapLen())
}

/*
	-- Record types --
*/

// UserPictureStatus indicate UserPicture status
type UserPictureStatus uint8

// UserPicture status
const (
	UserPictureUnset UserPictureStatus = iota
	UserPictureRegister
	UserPictureRemove
	UserPictureBlockByJustice
)
