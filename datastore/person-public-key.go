/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	"../libgo/syllab"
)

const (
	personPublicKeyStructureID uint64 = 13183953152561975962
	personPublicKeyFixedSize   uint64 = 177 // 72 + 105 + (0 * 8) >> Common header + Unique data + vars add&&len
	personPublicKeyState       uint8  = ganjine.DataStructureStatePreAlpha
)

// PersonPublicKey store person Public-key (asymmetric) cryptography!
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
	PersonID         [16]byte // UUID of Person
	ThingID          [16]byte
	PublicKey        [32]byte // Use new algorithm like ECC(256bit) instead of RSA(4096bit)
	ExpireAt         int64
	Status           personPublicKeyStatus
}

type personPublicKeyStatus uint8

// PersonPublicKey status
const (
	// PersonPublicKeyRevoked indicate related public key was inactivated by person
	PersonPublicKeyRevoked personPublicKeyStatus = iota
	// 1: just password
	// 2: password + otp
	// 3: password + identification factors
	// 4: password + otp + identification factors
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
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByPublicKey method find and read last version of record by given PublicKey
func (ppk *PersonPublicKey) GetByPublicKey() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: ppk.hashPublicKey(),
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
		ppk.RecordID = indexRes.RecordIDs[ln]
		err = ppk.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexPersonID index ppk.PersonID to retrieve record fast later.
func (ppk *PersonPublicKey) IndexPersonID() {
	var personIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.hashPersonID(),
		RecordID:  ppk.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &personIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexThingID index ppk.ThingID to retrieve record fast later.
func (ppk *PersonPublicKey) IndexThingID() {
	var thingIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.hashThingID(),
		RecordID:  ppk.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &thingIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexPublicKey index ppk.PublicKey to retrieve record fast later.
func (ppk *PersonPublicKey) IndexPublicKey() {
	var publicKeyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ppk.hashPublicKey(),
		RecordID:  ppk.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &publicKeyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ppk *PersonPublicKey) hashPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(ppk.RecordStructureID)
	buf[1] = byte(ppk.RecordStructureID >> 8)
	buf[2] = byte(ppk.RecordStructureID >> 16)
	buf[3] = byte(ppk.RecordStructureID >> 24)
	buf[4] = byte(ppk.RecordStructureID >> 32)
	buf[5] = byte(ppk.RecordStructureID >> 40)
	buf[6] = byte(ppk.RecordStructureID >> 48)
	buf[7] = byte(ppk.RecordStructureID >> 56)

	copy(buf[8:], ppk.PersonID[:])

	return sha512.Sum512_256(buf)
}

func (ppk *PersonPublicKey) hashThingID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(ppk.RecordStructureID)
	buf[1] = byte(ppk.RecordStructureID >> 8)
	buf[2] = byte(ppk.RecordStructureID >> 16)
	buf[3] = byte(ppk.RecordStructureID >> 24)
	buf[4] = byte(ppk.RecordStructureID >> 32)
	buf[5] = byte(ppk.RecordStructureID >> 40)
	buf[6] = byte(ppk.RecordStructureID >> 48)
	buf[7] = byte(ppk.RecordStructureID >> 56)

	copy(buf[8:], ppk.ThingID[:])

	return sha512.Sum512_256(buf)
}

func (ppk *PersonPublicKey) hashPublicKey() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32

	buf[0] = byte(ppk.RecordStructureID)
	buf[1] = byte(ppk.RecordStructureID >> 8)
	buf[2] = byte(ppk.RecordStructureID >> 16)
	buf[3] = byte(ppk.RecordStructureID >> 24)
	buf[4] = byte(ppk.RecordStructureID >> 32)
	buf[5] = byte(ppk.RecordStructureID >> 40)
	buf[6] = byte(ppk.RecordStructureID >> 48)
	buf[7] = byte(ppk.RecordStructureID >> 56)

	copy(buf[8:], ppk.PublicKey[:])

	return sha512.Sum512_256(buf)
}

func (ppk *PersonPublicKey) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < personPublicKeyFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(ppk.RecordID[:], buf[:])
	ppk.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	ppk.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	ppk.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(ppk.OwnerAppID[:], buf[56:])

	copy(ppk.AppInstanceID[:], buf[72:])
	copy(ppk.UserConnectionID[:], buf[88:])
	copy(ppk.PersonID[:], buf[104:])
	copy(ppk.ThingID[:], buf[120:])
	copy(ppk.PublicKey[:], buf[136:])
	ppk.ExpireAt = int64(buf[168]) | int64(buf[169])<<8 | int64(buf[170])<<16 | int64(buf[171])<<24 | int64(buf[172])<<32 | int64(buf[173])<<40 | int64(buf[174])<<48 | int64(buf[175])<<56
	ppk.Status = personPublicKeyStatus(buf[176])

	return
}

func (ppk *PersonPublicKey) syllabEncoder() (buf []byte) {
	buf = make([]byte, ppk.syllabLen())

	// copy(buf[0:], ppk.RecordID[:])
	buf[32] = byte(ppk.RecordStructureID)
	buf[33] = byte(ppk.RecordStructureID >> 8)
	buf[34] = byte(ppk.RecordStructureID >> 16)
	buf[35] = byte(ppk.RecordStructureID >> 24)
	buf[36] = byte(ppk.RecordStructureID >> 32)
	buf[37] = byte(ppk.RecordStructureID >> 40)
	buf[38] = byte(ppk.RecordStructureID >> 48)
	buf[39] = byte(ppk.RecordStructureID >> 56)
	buf[40] = byte(ppk.RecordSize)
	buf[41] = byte(ppk.RecordSize >> 8)
	buf[42] = byte(ppk.RecordSize >> 16)
	buf[43] = byte(ppk.RecordSize >> 24)
	buf[44] = byte(ppk.RecordSize >> 32)
	buf[45] = byte(ppk.RecordSize >> 40)
	buf[46] = byte(ppk.RecordSize >> 48)
	buf[47] = byte(ppk.RecordSize >> 56)
	buf[48] = byte(ppk.WriteTime)
	buf[49] = byte(ppk.WriteTime >> 8)
	buf[50] = byte(ppk.WriteTime >> 16)
	buf[51] = byte(ppk.WriteTime >> 24)
	buf[52] = byte(ppk.WriteTime >> 32)
	buf[53] = byte(ppk.WriteTime >> 40)
	buf[54] = byte(ppk.WriteTime >> 48)
	buf[55] = byte(ppk.WriteTime >> 56)
	copy(buf[56:], ppk.OwnerAppID[:])

	copy(buf[72:], ppk.AppInstanceID[:])
	copy(buf[88:], ppk.UserConnectionID[:])
	copy(buf[104:], ppk.PersonID[:])
	copy(buf[120:], ppk.ThingID[:])
	copy(buf[136:], ppk.PublicKey[:])
	buf[168] = byte(ppk.ExpireAt)
	buf[169] = byte(ppk.ExpireAt >> 8)
	buf[170] = byte(ppk.ExpireAt >> 16)
	buf[171] = byte(ppk.ExpireAt >> 24)
	buf[172] = byte(ppk.ExpireAt >> 32)
	buf[173] = byte(ppk.ExpireAt >> 40)
	buf[174] = byte(ppk.ExpireAt >> 48)
	buf[175] = byte(ppk.ExpireAt >> 56)
	buf[176] = byte(ppk.Status)

	return
}

func (ppk *PersonPublicKey) syllabLen() uint64 {
	return personPublicKeyFixedSize
}
