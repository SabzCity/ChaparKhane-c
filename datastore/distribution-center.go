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
	distributionCenterStructureID uint64 = 465051532317110086
	distributionCenterFixedSize   uint64 = 177 // 72 + 97 + (1 * 8) >> Common header + Unique data + vars add&&len
	distributionCenterState       uint8  = ganjine.DataStructureStatePreAlpha
)

// DistributionCenter store not just warehouses but any DC type like stores that do many more things like package multi product to send!
type DistributionCenter struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	ID               [16]byte
	Name             string
	OrgID            [16]byte
	Type             uint8    // Private, Public,
	ThingID          [16]byte // To get more data like map of DC, ...
	CoordinateID     [16]byte
}

// Set method set some data and write entire DistributionCenter record!
func (dc *DistributionCenter) Set() (err error) {
	dc.RecordStructureID = distributionCenterStructureID
	dc.RecordSize = dc.syllabLen()
	dc.WriteTime = etime.Now()
	dc.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: dc.syllabEncoder(),
	}
	dc.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], dc.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (dc *DistributionCenter) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: dc.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = dc.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if dc.RecordStructureID != distributionCenterStructureID {
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByID method find and read last version of record by given ID
func (dc *DistributionCenter) GetByID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: dc.HashID(),
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
		dc.RecordID = indexRes.RecordIDs[ln]
		err = dc.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// GetByName method find and read last version of record by given Name
func (dc *DistributionCenter) GetByName() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: dc.HashName(),
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
		dc.RecordID = indexRes.RecordIDs[ln]
		err = dc.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexID index dc.ID to retrieve record fast later.
func (dc *DistributionCenter) IndexID() {
	var idIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: dc.HashID(),
		RecordID:  dc.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &idIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexName index dc.Name to retrieve record fast later.
func (dc *DistributionCenter) IndexName() {
	var nameIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: dc.HashName(),
		RecordID:  dc.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &nameIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashName hash distributionCenterStructureID + Name
func (dc *DistributionCenter) HashName() (hash [32]byte) {
	var buf = make([]byte, 8+len(dc.Name)) // 8+16

	buf[0] = byte(dc.RecordStructureID)
	buf[1] = byte(dc.RecordStructureID >> 8)
	buf[2] = byte(dc.RecordStructureID >> 16)
	buf[3] = byte(dc.RecordStructureID >> 24)
	buf[4] = byte(dc.RecordStructureID >> 32)
	buf[5] = byte(dc.RecordStructureID >> 40)
	buf[6] = byte(dc.RecordStructureID >> 48)
	buf[7] = byte(dc.RecordStructureID >> 56)

	copy(buf[8:], dc.Name)

	return sha512.Sum512_256(buf)
}

// HashID hash distributionCenterStructureID + ID
func (dc *DistributionCenter) HashID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(dc.RecordStructureID)
	buf[1] = byte(dc.RecordStructureID >> 8)
	buf[2] = byte(dc.RecordStructureID >> 16)
	buf[3] = byte(dc.RecordStructureID >> 24)
	buf[4] = byte(dc.RecordStructureID >> 32)
	buf[5] = byte(dc.RecordStructureID >> 40)
	buf[6] = byte(dc.RecordStructureID >> 48)
	buf[7] = byte(dc.RecordStructureID >> 56)

	copy(buf[8:], dc.ID[:])

	return sha512.Sum512_256(buf)
}

func (dc *DistributionCenter) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < distributionCenterFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(dc.RecordID[:], buf[:])
	dc.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	dc.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	dc.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(dc.OwnerAppID[:], buf[56:])

	copy(dc.AppInstanceID[:], buf[72:])
	copy(dc.UserConnectionID[:], buf[88:])
	copy(dc.ID[:], buf[104:])
	var NameAdd = uint32(buf[120]) | uint32(buf[121])<<8 | uint32(buf[122])<<16 | uint32(buf[123])<<24
	var NameLen = uint32(buf[124]) | uint32(buf[125])<<8 | uint32(buf[126])<<16 | uint32(buf[127])<<24
	dc.Name = string(buf[NameAdd:NameLen])
	copy(dc.OrgID[:], buf[128:])
	dc.Type = uint8(buf[144])
	copy(dc.ThingID[:], buf[145:])
	copy(dc.CoordinateID[:], buf[161:])

	return
}

func (dc *DistributionCenter) syllabEncoder() (buf []byte) {
	buf = make([]byte, dc.syllabLen())

	// copy(buf[0:], dc.RecordID[:])
	buf[32] = byte(dc.RecordStructureID)
	buf[33] = byte(dc.RecordStructureID >> 8)
	buf[34] = byte(dc.RecordStructureID >> 16)
	buf[35] = byte(dc.RecordStructureID >> 24)
	buf[36] = byte(dc.RecordStructureID >> 32)
	buf[37] = byte(dc.RecordStructureID >> 40)
	buf[38] = byte(dc.RecordStructureID >> 48)
	buf[39] = byte(dc.RecordStructureID >> 56)
	buf[40] = byte(dc.RecordSize)
	buf[41] = byte(dc.RecordSize >> 8)
	buf[42] = byte(dc.RecordSize >> 16)
	buf[43] = byte(dc.RecordSize >> 24)
	buf[44] = byte(dc.RecordSize >> 32)
	buf[45] = byte(dc.RecordSize >> 40)
	buf[46] = byte(dc.RecordSize >> 48)
	buf[47] = byte(dc.RecordSize >> 56)
	buf[48] = byte(dc.WriteTime)
	buf[49] = byte(dc.WriteTime >> 8)
	buf[50] = byte(dc.WriteTime >> 16)
	buf[51] = byte(dc.WriteTime >> 24)
	buf[52] = byte(dc.WriteTime >> 32)
	buf[53] = byte(dc.WriteTime >> 40)
	buf[54] = byte(dc.WriteTime >> 48)
	buf[55] = byte(dc.WriteTime >> 56)
	copy(buf[56:], dc.OwnerAppID[:])

	copy(buf[72:], dc.AppInstanceID[:])
	copy(buf[88:], dc.UserConnectionID[:])
	copy(buf[104:], dc.ID[:])
	var ln = len(dc.Name)
	buf[120] = byte(distributionCenterFixedSize)
	buf[121] = byte(distributionCenterFixedSize >> 8)
	buf[122] = byte(distributionCenterFixedSize >> 16)
	buf[123] = byte(distributionCenterFixedSize >> 24)
	buf[124] = byte(ln)
	buf[125] = byte(ln >> 8)
	buf[126] = byte(ln >> 16)
	buf[127] = byte(ln >> 24)
	copy(buf[distributionCenterFixedSize:], dc.Name[:])
	copy(buf[128:], dc.OrgID[:])
	buf[144] = byte(dc.Type)
	copy(buf[145:], dc.ThingID[:])
	copy(buf[161:], dc.CoordinateID[:])

	return
}

func (dc *DistributionCenter) syllabLen() uint64 {
	return distributionCenterFixedSize + uint64(len(dc.Name))
}
