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
	personAuthenticationStructureID uint64 = 2759743017265268907
)

var personAuthenticationStructure = ganjine.DataStructure{
	ID:                2759743017265268907,
	IssueDate:         1595888151,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         PersonAuthentication{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "PersonAuthentication",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store real person authenticate data",
	},
	TAGS: []string{
		"",
	},
}

// PersonAuthentication ---Read locale description in userAppsConnectionStructure---
type PersonAuthentication struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time `hash-index:"PersonID[hourly]"`
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	PersonID         [32]byte `ganjine:"Unique" hash-index:"RecordID"` // UUID of Person
	ReferentPersonID [32]byte `hash-index:"PersonID[if,ReferentPersonID]"`
	Status           PersonAuthenticationStatus

	// Person Authentication Factors https://en.wikipedia.org/wiki/Authentication#Factors_and_identity
	PasswordHash  [32]byte
	OTPPattern    [32]byte // https://tools.ietf.org/html/rfc6238
	OTPAdditional int32    // easy to be 2 to 7 digit. https://en.wikipedia.org/wiki/Personal_identification_number
	SecurityKey   [32]byte
}

// PersonAuthenticationStatus indicate PersonAuthentication record status
type PersonAuthenticationStatus uint8

// PersonAuthentication status
const (
	PersonAuthenticationUnset              PersonAuthenticationStatus = iota
	PersonAuthenticationInactive                                      // person had been inactive and can't be use now!
	PersonAuthenticationBlocked                                       // person had been blocked and can't be use now!
	PersonAuthenticationNotForceUse2Factor                            // authenticate person just with Password
	PersonAuthenticationForceUse2Factor                               // authenticate person with Password + OTP
	PersonAuthenticationMustChangePassword                            // user must change password to increase security!
)

// Set method set some data and write entire PersonAuthentication record!
func (pa *PersonAuthentication) Set() (err *er.Error) {
	pa.RecordStructureID = personAuthenticationStructureID
	pa.RecordSize = pa.syllabLen()
	pa.WriteTime = etime.Now()
	pa.OwnerAppID = server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pa.syllabEncoder(),
	}
	pa.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pa.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Record:", err)
		}
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pa *PersonAuthentication) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          pa.RecordID,
		RecordStructureID: personAuthenticationStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = pa.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if pa.RecordStructureID != personAuthenticationStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByPersonID method find and read last version of record by given PersonID!
func (pa *PersonAuthentication) GetLastByPersonID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashPersonIDforRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	pa.RecordID = indexRes.IndexValues[0]
	err = pa.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", personAuthenticationStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexPersonID index pa.PersonID to retrieve record fast later.
func (pa *PersonAuthentication) IndexPersonID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashPersonIDforRecordID(),
		IndexValue: pa.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index:", err)
		}
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *PersonAuthentication) hashPersonIDforRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	copy(buf[8:], pa.PersonID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexPersonIDforRegisterTime index pa.WriteTime to retrieve all register person on specific time in hour rate.
// Each year is 8760 hour (365*24) that indicate we have 8760 index record each year!
func (pa *PersonAuthentication) IndexPersonIDforRegisterTime() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashWriteTimeforPersonIDHourly(),
		IndexValue: pa.PersonID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index:", err)
		}
		// TODO::: we must retry more due to record wrote successfully before this func!
	}
}

func (pa *PersonAuthentication) hashWriteTimeforPersonIDHourly() (hash [32]byte) {
	const field = "WriteTime"
	var buf = make([]byte, 16) // 8+8
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	syllab.SetInt64(buf, 8, pa.WriteTime.RoundToHour())
	copy(buf[16:], field)
	return sha512.Sum512_256(buf)
}

// IndexPersonIDforReferentPersonID index pa.ReferentPersonID to retrieve record fast later.
func (pa *PersonAuthentication) IndexPersonIDforReferentPersonID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashIfReferentPersonIDforPersonID(),
		IndexValue: pa.PersonID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index:", err)
		}
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *PersonAuthentication) hashIfReferentPersonIDforPersonID() (hash [32]byte) {
	const field = "IfReferentPersonID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	copy(buf[8:], pa.ReferentPersonID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (pa *PersonAuthentication) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < pa.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(pa.RecordID[:], buf[0:])
	pa.RecordStructureID = syllab.GetUInt64(buf, 32)
	pa.RecordSize = syllab.GetUInt64(buf, 40)
	pa.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(pa.OwnerAppID[:], buf[56:])

	copy(pa.AppInstanceID[:], buf[88:])
	copy(pa.UserConnectionID[:], buf[120:])
	copy(pa.PersonID[:], buf[152:])
	copy(pa.ReferentPersonID[:], buf[184:])
	pa.Status = PersonAuthenticationStatus(syllab.GetUInt8(buf, 216))

	copy(pa.PasswordHash[:], buf[217:])
	copy(pa.OTPPattern[:], buf[249:])
	pa.OTPAdditional = syllab.GetInt32(buf, 281)
	copy(pa.SecurityKey[:], buf[285:])
	return
}

func (pa *PersonAuthentication) syllabEncoder() (buf []byte) {
	buf = make([]byte, pa.syllabLen())

	// copy(buf[0:], pa.RecordID[:])
	syllab.SetUInt64(buf, 32, pa.RecordStructureID)
	syllab.SetUInt64(buf, 40, pa.RecordSize)
	syllab.SetInt64(buf, 48, int64(pa.WriteTime))
	copy(buf[56:], pa.OwnerAppID[:])

	copy(buf[88:], pa.AppInstanceID[:])
	copy(buf[120:], pa.UserConnectionID[:])
	copy(buf[152:], pa.PersonID[:])
	copy(buf[184:], pa.ReferentPersonID[:])
	syllab.SetUInt8(buf, 216, uint8(pa.Status))

	copy(buf[217:], pa.PasswordHash[:])
	copy(buf[249:], pa.OTPPattern[:])
	syllab.SetInt32(buf, 281, pa.OTPAdditional)
	copy(buf[285:], pa.SecurityKey[:])
	return
}

func (pa *PersonAuthentication) syllabStackLen() (ln uint32) {
	return 317
}

func (pa *PersonAuthentication) syllabHeapLen() (ln uint32) {
	return
}

func (pa *PersonAuthentication) syllabLen() (ln uint64) {
	return uint64(pa.syllabStackLen() + pa.syllabHeapLen())
}
