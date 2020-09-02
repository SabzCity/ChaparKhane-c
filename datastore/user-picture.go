/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	"../libgo/picture"
	"../libgo/syllab"
)

const (
	userPictureStructureID uint64 = 9588981481850124477
	userPictureFixedSize   uint64 = 154 // 72 + 82 + (0 * 8) >> Common header + Unique data + vars add&&len
	userPictureState       uint8  = ganjine.DataStructureStatePreAlpha
)

// UserPicture store any user type e.g. person,org,... official pictures.
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
	UserID           [16]byte
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
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByUserID method find and read last version of record by given UserID
func (up *UserPicture) GetByUserID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: up.hashUserID(),
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
	for {
		ln--
		up.RecordID = indexRes.RecordIDs[ln]
		err = up.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexUserID index up.UserID to retrieve record fast later.
func (up *UserPicture) IndexUserID() {
	var userIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: up.hashUserID(),
		RecordID:  up.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &userIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (up *UserPicture) hashUserID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(up.RecordStructureID)
	buf[1] = byte(up.RecordStructureID >> 8)
	buf[2] = byte(up.RecordStructureID >> 16)
	buf[3] = byte(up.RecordStructureID >> 24)
	buf[4] = byte(up.RecordStructureID >> 32)
	buf[5] = byte(up.RecordStructureID >> 40)
	buf[6] = byte(up.RecordStructureID >> 48)
	buf[7] = byte(up.RecordStructureID >> 56)

	copy(buf[8:], up.UserID[:])

	return sha512.Sum512_256(buf)
}

func (up *UserPicture) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < userPictureFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(up.RecordID[:], buf[:])
	up.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	up.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	up.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(up.OwnerAppID[:], buf[56:])

	copy(up.AppInstanceID[:], buf[72:])
	copy(up.UserConnectionID[:], buf[88:])
	copy(up.UserID[:], buf[104:])
	copy(up.ObjectID[:], buf[120:])
	up.Rating = picture.Rating(buf[152])
	up.Status = userPictureStatus(buf[153])

	return
}

func (up *UserPicture) syllabEncoder() (buf []byte) {
	buf = make([]byte, up.syllabLen())

	// copy(buf[0:], up.RecordID[:])
	buf[32] = byte(up.RecordStructureID)
	buf[33] = byte(up.RecordStructureID >> 8)
	buf[34] = byte(up.RecordStructureID >> 16)
	buf[35] = byte(up.RecordStructureID >> 24)
	buf[36] = byte(up.RecordStructureID >> 32)
	buf[37] = byte(up.RecordStructureID >> 40)
	buf[38] = byte(up.RecordStructureID >> 48)
	buf[39] = byte(up.RecordStructureID >> 56)
	buf[40] = byte(up.RecordSize)
	buf[41] = byte(up.RecordSize >> 8)
	buf[42] = byte(up.RecordSize >> 16)
	buf[43] = byte(up.RecordSize >> 24)
	buf[44] = byte(up.RecordSize >> 32)
	buf[45] = byte(up.RecordSize >> 40)
	buf[46] = byte(up.RecordSize >> 48)
	buf[47] = byte(up.RecordSize >> 56)
	buf[48] = byte(up.WriteTime)
	buf[49] = byte(up.WriteTime >> 8)
	buf[50] = byte(up.WriteTime >> 16)
	buf[51] = byte(up.WriteTime >> 24)
	buf[52] = byte(up.WriteTime >> 32)
	buf[53] = byte(up.WriteTime >> 40)
	buf[54] = byte(up.WriteTime >> 48)
	buf[55] = byte(up.WriteTime >> 56)
	copy(buf[56:], up.OwnerAppID[:])

	copy(buf[72:], up.AppInstanceID[:])
	copy(buf[88:], up.UserConnectionID[:])
	copy(buf[104:], up.UserID[:])
	copy(buf[120:], up.ObjectID[:])
	buf[152] = byte(up.Rating)
	buf[153] = byte(up.Status)

	return
}

func (up *UserPicture) syllabLen() uint64 {
	return userPictureFixedSize
}
