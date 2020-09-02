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
	personNumberStructureID uint64 = 1212190932488392076
	personNumberFixedSize   uint64 = 129 // 72 + 57 + (0 * 8) >> Common header + Unique data + vars add&&len
	personNumberState       uint8  = ganjine.DataStructureStatePreAlpha
)

// PersonNumber store user number that act for some process like exiting phone, mobile, ...
type PersonNumber struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	PersonID         [16]byte
	Number           uint64 // must start with country code e.g. (00)98-912-345-6789
	Status           personNumberStatus
}

type personNumberStatus uint8

// PersonNumber status
const (
	PersonNumberRegister personNumberStatus = iota
	PersonNumberRemove
	PersonNumberBlockByJustice
)

// Set method set some data and write entire PersonNumber record!
func (pn *PersonNumber) Set() (err error) {
	pn.RecordStructureID = personNumberStructureID
	pn.RecordSize = pn.syllabLen()
	pn.WriteTime = etime.Now()
	pn.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pn.syllabEncoder(),
	}
	pn.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pn.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pn *PersonNumber) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: pn.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = pn.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if pn.RecordStructureID != personNumberStructureID {
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByPersonID method find and read last version of record by given PersonID
func (pn *PersonNumber) GetByPersonID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pn.hashPersonID(),
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
		pn.RecordID = indexRes.RecordIDs[ln]
		err = pn.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// GetByNumber method find and read last version of record by given Number
func (pn *PersonNumber) GetByNumber() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pn.hashNumber(),
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
		pn.RecordID = indexRes.RecordIDs[ln]
		err = pn.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexPersonID index pn.PersonID to retrieve record fast later.
func (pn *PersonNumber) IndexPersonID() {
	var personIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pn.hashPersonID(),
		RecordID:  pn.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &personIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexNumber index pn.Number to retrieve record fast later.
func (pn *PersonNumber) IndexNumber() {
	var numberIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pn.hashNumber(),
		RecordID:  pn.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &numberIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pn *PersonNumber) hashPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(pn.RecordStructureID)
	buf[1] = byte(pn.RecordStructureID >> 8)
	buf[2] = byte(pn.RecordStructureID >> 16)
	buf[3] = byte(pn.RecordStructureID >> 24)
	buf[4] = byte(pn.RecordStructureID >> 32)
	buf[5] = byte(pn.RecordStructureID >> 40)
	buf[6] = byte(pn.RecordStructureID >> 48)
	buf[7] = byte(pn.RecordStructureID >> 56)

	copy(buf[8:], pn.PersonID[:])

	return sha512.Sum512_256(buf)
}

func (pn *PersonNumber) hashNumber() (hash [32]byte) {
	var buf = make([]byte, 16) // 8+8

	buf[0] = byte(pn.RecordStructureID)
	buf[1] = byte(pn.RecordStructureID >> 8)
	buf[2] = byte(pn.RecordStructureID >> 16)
	buf[3] = byte(pn.RecordStructureID >> 24)
	buf[4] = byte(pn.RecordStructureID >> 32)
	buf[5] = byte(pn.RecordStructureID >> 40)
	buf[6] = byte(pn.RecordStructureID >> 48)
	buf[7] = byte(pn.RecordStructureID >> 56)

	buf[8] = byte(pn.Number)
	buf[9] = byte(pn.Number >> 8)
	buf[10] = byte(pn.Number >> 16)
	buf[11] = byte(pn.Number >> 24)
	buf[12] = byte(pn.Number >> 32)
	buf[13] = byte(pn.Number >> 40)
	buf[14] = byte(pn.Number >> 48)
	buf[15] = byte(pn.Number >> 56)

	return sha512.Sum512_256(buf)
}

func (pn *PersonNumber) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < personNumberFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(pn.RecordID[:], buf[:])
	pn.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	pn.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	pn.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(pn.OwnerAppID[:], buf[56:])

	copy(pn.AppInstanceID[:], buf[72:])
	copy(pn.UserConnectionID[:], buf[88:])
	copy(pn.PersonID[:], buf[104:])
	pn.Number = uint64(buf[120]) | uint64(buf[121])<<8 | uint64(buf[122])<<16 | uint64(buf[123])<<24 | uint64(buf[124])<<32 | uint64(buf[125])<<40 | uint64(buf[126])<<48 | uint64(buf[127])<<56
	pn.Status = personNumberStatus(buf[128])

	return
}

func (pn *PersonNumber) syllabEncoder() (buf []byte) {
	buf = make([]byte, pn.syllabLen())

	// copy(buf[0:], pn.RecordID[:])
	buf[32] = byte(pn.RecordStructureID)
	buf[33] = byte(pn.RecordStructureID >> 8)
	buf[34] = byte(pn.RecordStructureID >> 16)
	buf[35] = byte(pn.RecordStructureID >> 24)
	buf[36] = byte(pn.RecordStructureID >> 32)
	buf[37] = byte(pn.RecordStructureID >> 40)
	buf[38] = byte(pn.RecordStructureID >> 48)
	buf[39] = byte(pn.RecordStructureID >> 56)
	buf[40] = byte(pn.RecordSize)
	buf[41] = byte(pn.RecordSize >> 8)
	buf[42] = byte(pn.RecordSize >> 16)
	buf[43] = byte(pn.RecordSize >> 24)
	buf[44] = byte(pn.RecordSize >> 32)
	buf[45] = byte(pn.RecordSize >> 40)
	buf[46] = byte(pn.RecordSize >> 48)
	buf[47] = byte(pn.RecordSize >> 56)
	buf[48] = byte(pn.WriteTime)
	buf[49] = byte(pn.WriteTime >> 8)
	buf[50] = byte(pn.WriteTime >> 16)
	buf[51] = byte(pn.WriteTime >> 24)
	buf[52] = byte(pn.WriteTime >> 32)
	buf[53] = byte(pn.WriteTime >> 40)
	buf[54] = byte(pn.WriteTime >> 48)
	buf[55] = byte(pn.WriteTime >> 56)
	copy(buf[56:], pn.OwnerAppID[:])

	copy(buf[72:], pn.AppInstanceID[:])
	copy(buf[88:], pn.UserConnectionID[:])
	copy(buf[104:], pn.PersonID[:])
	buf[120] = byte(pn.Number)
	buf[121] = byte(pn.Number >> 8)
	buf[122] = byte(pn.Number >> 16)
	buf[123] = byte(pn.Number >> 24)
	buf[124] = byte(pn.Number >> 32)
	buf[125] = byte(pn.Number >> 40)
	buf[126] = byte(pn.Number >> 48)
	buf[127] = byte(pn.Number >> 56)
	buf[128] = byte(pn.Status)

	return
}

func (pn *PersonNumber) syllabLen() uint64 {
	return personNumberFixedSize
}
