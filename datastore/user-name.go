/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"
	"unsafe"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	"../libgo/syllab"
)

const (
	userNameStructureID uint64 = 12744998016788909151
	userNameFixedSize   uint64 = 129 // 72 + 49 + (1 * 8) >> Common header + Unique data + vars add&&len
	userNameState       uint8  = ganjine.DataStructureStatePreAlpha
)

// UserName store user name that translate it to UserID for any purpose like login, send message, ...!
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
	UserID           [16]byte
	Username         string // It is not replace of user ID! It usually use to find user by their friends!
	Status           userNameStatus
}

type userNameStatus uint8

// UserName status
const (
	UserNameRegister userNameStatus = iota
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
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByUserID method find and read last version of record by given UserID
func (un *UserName) GetByUserID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: un.hashUserID(),
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
		un.RecordID = indexRes.RecordIDs[ln]
		err = un.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// GetByUserName method find and read last version of record by given UserName
func (un *UserName) GetByUserName() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: un.hashUserName(),
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
		un.RecordID = indexRes.RecordIDs[ln]
		err = un.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexUserName index un.UserName to retrieve record fast later.
func (un *UserName) IndexUserName() {
	var userNameIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: un.hashUserName(),
		RecordID:  un.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &userNameIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexUserID index un.UserID to retrieve record fast later.
func (un *UserName) IndexUserID() {
	var userIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: un.hashUserID(),
		RecordID:  un.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &userIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (un *UserName) hashUserID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(un.RecordStructureID)
	buf[1] = byte(un.RecordStructureID >> 8)
	buf[2] = byte(un.RecordStructureID >> 16)
	buf[3] = byte(un.RecordStructureID >> 24)
	buf[4] = byte(un.RecordStructureID >> 32)
	buf[5] = byte(un.RecordStructureID >> 40)
	buf[6] = byte(un.RecordStructureID >> 48)
	buf[7] = byte(un.RecordStructureID >> 56)

	copy(buf[8:], un.UserID[:])

	return sha512.Sum512_256(buf)
}

func (un *UserName) hashUserName() (hash [32]byte) {
	var buf = make([]byte, 8+len(un.Username))

	buf[0] = byte(un.RecordStructureID)
	buf[1] = byte(un.RecordStructureID >> 8)
	buf[2] = byte(un.RecordStructureID >> 16)
	buf[3] = byte(un.RecordStructureID >> 24)
	buf[4] = byte(un.RecordStructureID >> 32)
	buf[5] = byte(un.RecordStructureID >> 40)
	buf[6] = byte(un.RecordStructureID >> 48)
	buf[7] = byte(un.RecordStructureID >> 56)

	copy(buf[8:], un.Username)

	return sha512.Sum512_256(buf)
}

func (un *UserName) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < userNameFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(un.RecordID[:], buf[:])
	un.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	un.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	un.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(un.OwnerAppID[:], buf[56:])

	copy(un.AppInstanceID[:], buf[72:])
	copy(un.UserConnectionID[:], buf[88:])
	copy(un.UserID[:], buf[104:])
	var userNameAdd = uint32(buf[120]) | uint32(buf[121])<<8 | uint32(buf[122])<<16 | uint32(buf[123])<<24
	var userNameLen = uint32(buf[124]) | uint32(buf[125])<<8 | uint32(buf[126])<<16 | uint32(buf[127])<<24
	un.Status = userNameStatus(buf[128])

	// It must check len of every heap access but due to encode of data is safe proccess by us, skip it here!
	buf = buf[userNameAdd:userNameLen]
	un.Username = *(*string)(unsafe.Pointer(&buf))

	return
}

func (un *UserName) syllabEncoder() (buf []byte) {
	buf = make([]byte, un.syllabLen())

	// copy(buf[0:], un.RecordID[:])
	buf[32] = byte(un.RecordStructureID)
	buf[33] = byte(un.RecordStructureID >> 8)
	buf[34] = byte(un.RecordStructureID >> 16)
	buf[35] = byte(un.RecordStructureID >> 24)
	buf[36] = byte(un.RecordStructureID >> 32)
	buf[37] = byte(un.RecordStructureID >> 40)
	buf[38] = byte(un.RecordStructureID >> 48)
	buf[39] = byte(un.RecordStructureID >> 56)
	buf[40] = byte(un.RecordSize)
	buf[41] = byte(un.RecordSize >> 8)
	buf[42] = byte(un.RecordSize >> 16)
	buf[43] = byte(un.RecordSize >> 24)
	buf[44] = byte(un.RecordSize >> 32)
	buf[45] = byte(un.RecordSize >> 40)
	buf[46] = byte(un.RecordSize >> 48)
	buf[47] = byte(un.RecordSize >> 56)
	buf[48] = byte(un.WriteTime)
	buf[49] = byte(un.WriteTime >> 8)
	buf[50] = byte(un.WriteTime >> 16)
	buf[51] = byte(un.WriteTime >> 24)
	buf[52] = byte(un.WriteTime >> 32)
	buf[53] = byte(un.WriteTime >> 40)
	buf[54] = byte(un.WriteTime >> 48)
	buf[55] = byte(un.WriteTime >> 56)
	copy(buf[56:], un.OwnerAppID[:])

	copy(buf[72:], un.AppInstanceID[:])
	copy(buf[88:], un.UserConnectionID[:])
	copy(buf[104:], un.UserID[:])
	buf[120] = byte(userNameFixedSize) // Heap start index
	buf[121] = byte(userNameFixedSize >> 8)
	buf[122] = byte(userNameFixedSize >> 16)
	buf[123] = byte(userNameFixedSize >> 24)
	var ln = len(un.Username)
	buf[124] = byte(ln)
	buf[125] = byte(ln >> 8)
	buf[126] = byte(ln >> 16)
	buf[127] = byte(ln >> 24)
	copy(buf[userNameFixedSize:], un.Username)
	buf[128] = byte(un.Status)

	return
}

func (un *UserName) syllabLen() uint64 {
	return userNameFixedSize + uint64(len(un.Username))
}
