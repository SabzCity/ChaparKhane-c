/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"time"

	ganjine "../libgo/ganjine-sdk"
	"../libgo/uuid"
)

const personAuthenticationStructureID uint64 = 2759743017265268907

// PersonAuthentication store real person authenticate data.
type PersonAuthentication struct {
	/* Common header data */
	Checksum          [32]byte
	RecordID          [16]byte
	RecordSize        uint64
	RecordStructureID uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppConnectionID  [16]byte // Store to remember which app instance connection set||chanaged this record!
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

// Set method use to write entire PersonAuthentication record!
// pa can't be nil otherwise panic occur!
func (pa *PersonAuthentication) Set() (err error) {
	pa.RecordID = uuid.NewV4()
	pa.RecordStructureID = personAuthenticationStructureID
	pa.WriteTime = time.Now().Unix()
	pa.OwnerAppID = server.Manifest.AppID

	var req = ganjine.SetRecordReq{
		RecordID: pa.RecordID,
	}
	req.Record = pa.syllabEncoder()

	err = ganjine.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// Get method use to read all existing record just by given RecordID!
// pa can't be nil otherwise panic occur!
func (pa *PersonAuthentication) Get() (err error) {
	// TODO::: First read from local OS (related lib) as cache

	// If not exist in cache read from DataStore
	var req = ganjine.GetRecordReq{
		RecordID: pa.RecordID,
	}
	var res *ganjine.GetRecordRes
	res, err = ganjine.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	// TODO::: Write to local OS as cache if not enough storage exist do GC(Garbage Collector)

	err = pa.syllabDecoder(res.Record)

	return
}

// FindByPersonID method use to find and read last version of record by given PersonID!
// pa can't be nil otherwise panic occur!
func (pa *PersonAuthentication) FindByPersonID() (err error) {
	return
}

func (pa *PersonAuthentication) syllabDecoder(buf []byte) (err error) {
	copy(pa.Checksum[:], buf[0:])
	copy(pa.RecordID[:], buf[32:])
	pa.RecordSize = uint64(buf[48]) | uint64(buf[49])<<8 | uint64(buf[50])<<16 | uint64(buf[51])<<24 | uint64(buf[52])<<32 | uint64(buf[53])<<40 | uint64(buf[54])<<48 | uint64(buf[55])<<56
	pa.RecordStructureID = uint64(buf[56]) | uint64(buf[57])<<8 | uint64(buf[58])<<16 | uint64(buf[59])<<24 | uint64(buf[60])<<32 | uint64(buf[61])<<40 | uint64(buf[62])<<48 | uint64(buf[63])<<56
	pa.WriteTime = int64(buf[64]) | int64(buf[65])<<8 | int64(buf[66])<<16 | int64(buf[67])<<24 | int64(buf[68])<<32 | int64(buf[69])<<40 | int64(buf[70])<<48 | int64(buf[71])<<56
	copy(pa.OwnerAppID[:], buf[72:])
	copy(pa.AppConnectionID[:], buf[88:])
	copy(pa.UserConnectionID[:], buf[104:])
	copy(pa.PersonID[:], buf[120:])
	copy(pa.ReferentPersonID[:], buf[136:])
	pa.Status = personAuthenticationStatus(buf[152])
	copy(pa.PasswordHash[:], buf[153:])
	copy(pa.OTPPattern[:], buf[185:])
	pa.OTPAdditional = int32(buf[217]) | int32(buf[218])<<8 | int32(buf[219])<<16 | int32(buf[220])<<24
	copy(pa.RecoveryCode[:], buf[221:])
	pa.SecurityQuestion = uint16(buf[349]) | uint16(buf[350])<<8
	var SecurityAnswerAdd = uint32(buf[351]) | uint32(buf[352])<<8 | uint32(buf[353])<<16 | uint32(buf[354])<<24
	var SecurityAnswerLen = uint32(buf[355]) | uint32(buf[356])<<8 | uint32(buf[357])<<16 | uint32(buf[358])<<24
	pa.SecurityAnswer = string(buf[SecurityAnswerAdd:SecurityAnswerLen])

	return
}

func (pa *PersonAuthentication) syllabEncoder() (buf []byte) {
	var hsi int = 359 // Heap start index
	var ln int        // len of string, slices, maps, ...
	ln = hsi + len(pa.SecurityAnswer)
	buf = make([]byte, ln)

	copy(buf[0:], pa.Checksum[:])
	copy(buf[32:], pa.RecordID[:])
	buf[48] = byte(pa.RecordSize)
	buf[49] = byte(pa.RecordSize >> 8)
	buf[50] = byte(pa.RecordSize >> 16)
	buf[51] = byte(pa.RecordSize >> 24)
	buf[52] = byte(pa.RecordSize >> 32)
	buf[53] = byte(pa.RecordSize >> 40)
	buf[54] = byte(pa.RecordSize >> 48)
	buf[55] = byte(pa.RecordSize >> 56)
	buf[56] = byte(pa.RecordStructureID)
	buf[57] = byte(pa.RecordStructureID >> 8)
	buf[58] = byte(pa.RecordStructureID >> 16)
	buf[59] = byte(pa.RecordStructureID >> 24)
	buf[60] = byte(pa.RecordStructureID >> 32)
	buf[61] = byte(pa.RecordStructureID >> 40)
	buf[62] = byte(pa.RecordStructureID >> 48)
	buf[63] = byte(pa.RecordStructureID >> 56)
	buf[64] = byte(pa.WriteTime)
	buf[65] = byte(pa.WriteTime >> 8)
	buf[66] = byte(pa.WriteTime >> 16)
	buf[67] = byte(pa.WriteTime >> 24)
	buf[68] = byte(pa.WriteTime >> 32)
	buf[69] = byte(pa.WriteTime >> 40)
	buf[70] = byte(pa.WriteTime >> 48)
	buf[71] = byte(pa.WriteTime >> 56)
	copy(buf[72:], pa.OwnerAppID[:])
	copy(buf[88:], pa.AppConnectionID[:])
	copy(buf[104:], pa.UserConnectionID[:])
	copy(buf[120:], pa.PersonID[:])
	copy(buf[136:], pa.ReferentPersonID[:])
	buf[152] = byte(pa.Status)
	copy(buf[153:], pa.PasswordHash[:])
	copy(buf[185:], pa.OTPPattern[:])
	buf[217] = byte(pa.OTPAdditional)
	buf[218] = byte(pa.OTPAdditional >> 8)
	buf[219] = byte(pa.OTPAdditional >> 16)
	buf[220] = byte(pa.OTPAdditional >> 24)
	copy(buf[221:], pa.RecoveryCode[:])
	buf[349] = byte(pa.SecurityQuestion)
	buf[350] = byte(pa.SecurityQuestion >> 8)
	ln = len(pa.SecurityAnswer)
	buf[351] = byte(hsi)
	buf[352] = byte(hsi >> 8)
	buf[353] = byte(hsi >> 16)
	buf[354] = byte(hsi >> 24)
	buf[355] = byte(ln)
	buf[356] = byte(ln >> 8)
	buf[357] = byte(ln >> 16)
	buf[358] = byte(ln >> 24)
	copy(buf[hsi:], pa.SecurityAnswer[:])
	hsi += ln

	return
}
