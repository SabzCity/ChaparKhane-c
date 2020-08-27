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
	personAuthenticationStructureID uint64 = 2759743017265268907

	personAuthenticationFixedSize uint64 = 343 // 72 + 263 + (1 * 8) >> Common header + Unique data + vars add&&len
)

// PersonAuthentication store real person authenticate data.
type PersonAuthentication struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	PersonID         [16]byte // UUID of Person
	ReferentPersonID [16]byte
	Status           personAuthenticationStatus
	// Person Authentication Factors https://en.wikipedia.org/wiki/Authentication#Factors_and_identity
	PasswordHash  [32]byte // 256 bit SHA3-256 (Just SHA3)
	OTPPattern    [32]byte // https://tools.ietf.org/html/rfc6238
	OTPAdditional int32    // 4 to 7 digit. https://en.wikipedia.org/wiki/Personal_identification_number
	// RecoveryData
	RecoveryCode     [128]byte
	SecurityQuestion uint16 // https://en.wikipedia.org/wiki/Security_question
	SecurityAnswer   string // https://en.wikipedia.org/wiki/Security_question
}

type personAuthenticationStatus uint8

// PersonAuthentication status
const (
	// PersonAuthenticationInactive indicate person had been inactive and can't be use now!
	PersonAuthenticationInactive personAuthenticationStatus = iota
	// PersonAuthenticationBlocked indicate person had been blocked and can't be use now!
	PersonAuthenticationBlocked
	// PersonAuthenticationNotForceUse2Factor indicate authenticate person just with Password
	PersonAuthenticationNotForceUse2Factor
	// PersonAuthenticationForceUse2Factor indicate authenticate person with Password + OTP
	PersonAuthenticationForceUse2Factor
	// PersonAuthenticationMustChangePassword indicate user must change password to increase security!
	PersonAuthenticationMustChangePassword
)

// Set method set some data and write entire PersonAuthentication record!
func (pa *PersonAuthentication) Set() (err error) {
	pa.RecordStructureID = personAuthenticationStructureID
	pa.RecordSize = personAuthenticationFixedSize + uint64(len(pa.SecurityAnswer))
	pa.WriteTime = etime.Now()
	pa.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pa.syllabEncoder(),
	}
	pa.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pa.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pa *PersonAuthentication) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: pa.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = pa.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if pa.RecordStructureID != personAuthenticationStructureID {
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByPersonID method find and read last version of record by given PersonID!
func (pa *PersonAuthentication) GetByPersonID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pa.hashPersonID(),
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
		pa.RecordID = indexRes.RecordIDs[ln]
		err = pa.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexPersonID index pa.PersonID to retrieve record fast later.
func (pa *PersonAuthentication) IndexPersonID() {
	var personIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.hashPersonID(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &personIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexRegisterTime index pa.WriteTime to retrieve all register person on specific time in hour rate.
// Each year is 8760 hour (365*24) that indicate we have 8760 index record each year!
func (pa *PersonAuthentication) IndexRegisterTime() {
	var personIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.hashWriteTime(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &personIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully before this func!
	}
}

// IndexReferentPersonID index pa.ReferentPersonID to retrieve record fast later.
func (pa *PersonAuthentication) IndexReferentPersonID() {
	var personIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.hashReferentPersonID(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &personIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashPersonID hash personAuthenticationStructureID + pa.PersonID
func (pa *PersonAuthentication) hashPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	copy(buf[8:], pa.PersonID[:])

	return sha512.Sum512_256(buf)
}

// HashPersonID hash personAuthenticationStructureID + pa.WriteTime(round to hour)
func (pa *PersonAuthentication) hashWriteTime() (hash [32]byte) {
	var buf = make([]byte, 16) // 8+8

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	var roundedTime = etime.RoundToHour(pa.WriteTime)
	buf[8] = byte(roundedTime)
	buf[9] = byte(roundedTime >> 8)
	buf[10] = byte(roundedTime >> 16)
	buf[11] = byte(roundedTime >> 24)
	buf[12] = byte(roundedTime >> 32)
	buf[13] = byte(roundedTime >> 40)
	buf[14] = byte(roundedTime >> 48)
	buf[15] = byte(roundedTime >> 56)

	return sha512.Sum512_256(buf)
}

// hashReferentPersonID hash personAuthenticationStructureID + pa.ReferentPersonID
func (pa *PersonAuthentication) hashReferentPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	copy(buf[8:], pa.ReferentPersonID[:])

	return sha512.Sum512_256(buf)
}

func (pa *PersonAuthentication) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < personAuthenticationFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(pa.RecordID[:], buf[:])
	pa.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	pa.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	pa.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(pa.OwnerAppID[:], buf[56:])

	copy(pa.AppInstanceID[:], buf[72:])
	copy(pa.UserConnectionID[:], buf[88:])
	copy(pa.PersonID[:], buf[104:])
	copy(pa.ReferentPersonID[:], buf[120:])
	pa.Status = personAuthenticationStatus(buf[136])
	copy(pa.PasswordHash[:], buf[137:])
	copy(pa.OTPPattern[:], buf[169:])
	pa.OTPAdditional = int32(buf[201]) | int32(buf[202])<<8 | int32(buf[203])<<16 | int32(buf[204])<<24
	copy(pa.RecoveryCode[:], buf[205:])
	pa.SecurityQuestion = uint16(buf[333]) | uint16(buf[334])<<8
	var SecurityAnswerAdd = uint32(buf[335]) | uint32(buf[336])<<8 | uint32(buf[337])<<16 | uint32(buf[338])<<24
	var SecurityAnswerLen = uint32(buf[339]) | uint32(buf[340])<<8 | uint32(buf[341])<<16 | uint32(buf[342])<<24
	// It must check len of every heap access but due to encode of data is safe proccess, skip it here!
	pa.SecurityAnswer = string(buf[SecurityAnswerAdd:SecurityAnswerLen])

	return
}

func (pa *PersonAuthentication) syllabEncoder() (buf []byte) {
	var hsi int = int(personAuthenticationFixedSize) // Heap start index
	var ln int                                       // len of buf that include len of string, slices, maps, ...
	ln = hsi + len(pa.SecurityAnswer)
	buf = make([]byte, ln)

	// copy(buf[0:], pa.RecordID[:])
	buf[32] = byte(pa.RecordSize)
	buf[33] = byte(pa.RecordSize >> 8)
	buf[34] = byte(pa.RecordSize >> 16)
	buf[35] = byte(pa.RecordSize >> 24)
	buf[36] = byte(pa.RecordSize >> 32)
	buf[37] = byte(pa.RecordSize >> 40)
	buf[38] = byte(pa.RecordSize >> 48)
	buf[39] = byte(pa.RecordSize >> 56)
	buf[40] = byte(pa.RecordStructureID)
	buf[41] = byte(pa.RecordStructureID >> 8)
	buf[42] = byte(pa.RecordStructureID >> 16)
	buf[43] = byte(pa.RecordStructureID >> 24)
	buf[44] = byte(pa.RecordStructureID >> 32)
	buf[45] = byte(pa.RecordStructureID >> 40)
	buf[46] = byte(pa.RecordStructureID >> 48)
	buf[47] = byte(pa.RecordStructureID >> 56)
	buf[48] = byte(pa.WriteTime)
	buf[49] = byte(pa.WriteTime >> 8)
	buf[50] = byte(pa.WriteTime >> 16)
	buf[51] = byte(pa.WriteTime >> 24)
	buf[52] = byte(pa.WriteTime >> 32)
	buf[53] = byte(pa.WriteTime >> 40)
	buf[54] = byte(pa.WriteTime >> 48)
	buf[55] = byte(pa.WriteTime >> 56)
	copy(buf[56:], pa.OwnerAppID[:])

	copy(buf[72:], pa.AppInstanceID[:])
	copy(buf[88:], pa.UserConnectionID[:])
	copy(buf[104:], pa.PersonID[:])
	copy(buf[120:], pa.ReferentPersonID[:])
	buf[136] = byte(pa.Status)
	copy(buf[137:], pa.PasswordHash[:])
	copy(buf[169:], pa.OTPPattern[:])
	buf[201] = byte(pa.OTPAdditional)
	buf[202] = byte(pa.OTPAdditional >> 8)
	buf[203] = byte(pa.OTPAdditional >> 16)
	buf[204] = byte(pa.OTPAdditional >> 24)
	copy(buf[205:], pa.RecoveryCode[:])
	buf[333] = byte(pa.SecurityQuestion)
	buf[334] = byte(pa.SecurityQuestion >> 8)
	ln = len(pa.SecurityAnswer)
	buf[335] = byte(hsi)
	buf[336] = byte(hsi >> 8)
	buf[337] = byte(hsi >> 16)
	buf[338] = byte(hsi >> 24)
	buf[339] = byte(ln)
	buf[340] = byte(ln >> 8)
	buf[341] = byte(ln >> 16)
	buf[342] = byte(ln >> 24)
	copy(buf[343:], pa.SecurityAnswer[:])
	// hsi += ln

	return
}
