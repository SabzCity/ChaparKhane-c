/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
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
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	PersonID         [16]byte `ganjine:"Immutable,Unique"` // UUID of Person
	ReferentPersonID [16]byte `ganjine:"Immutable"`
	Status           PersonAuthenticationStatus

	// Person Authentication Factors https://en.wikipedia.org/wiki/Authentication#Factors_and_identity
	PasswordHash  [32]byte
	OTPPattern    [32]byte // https://tools.ietf.org/html/rfc6238
	OTPAdditional int32    // 4 to 7 digit. https://en.wikipedia.org/wiki/Personal_identification_number
	SecurityKey   [32]byte
}

// PersonAuthenticationStatus indicate PersonAuthentication record status
type PersonAuthenticationStatus uint8

// PersonAuthentication status
const (
	// PersonAuthenticationInactive indicate person had been inactive and can't be use now!
	PersonAuthenticationInactive PersonAuthenticationStatus = iota
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
	pa.RecordSize = pa.syllabLen()
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
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByPersonID method find and read last version of record by given PersonID!
func (pa *PersonAuthentication) GetByPersonID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pa.HashPersonID(),
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
		pa.RecordID = indexRes.RecordIDs[ln]
		err = pa.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexPersonID index pa.PersonID to retrieve record fast later.
func (pa *PersonAuthentication) IndexPersonID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashPersonID(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashPersonID hash personAuthenticationStructureID + pa.PersonID
func (pa *PersonAuthentication) HashPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	copy(buf[8:], pa.PersonID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexRegisterTime index pa.WriteTime to retrieve all register person on specific time in hour rate.
// Each year is 8760 hour (365*24) that indicate we have 8760 index record each year!
func (pa *PersonAuthentication) IndexRegisterTime() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashWriteTimeHourly(),
		RecordID:  pa.RecordID,
	}
	copy(indexRequest.RecordID[:], pa.PersonID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully before this func!
	}
}

// HashWriteTimeHourly hash personAuthenticationStructureID + pa.WriteTime(round to hour)
func (pa *PersonAuthentication) HashWriteTimeHourly() (hash [32]byte) {
	var buf = make([]byte, 16) // 8+8
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	syllab.SetInt64(buf, 8, etime.RoundToHour(pa.WriteTime))
	return sha512.Sum512_256(buf)
}

// IndexReferentPersonID index pa.ReferentPersonID to retrieve record fast later.
func (pa *PersonAuthentication) IndexReferentPersonID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashReferentPersonID(),
	}
	copy(indexRequest.RecordID[:], pa.PersonID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashReferentPersonID hash personAuthenticationStructureID + pa.ReferentPersonID
func (pa *PersonAuthentication) HashReferentPersonID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, personAuthenticationStructureID)
	copy(buf[8:], pa.ReferentPersonID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (pa *PersonAuthentication) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < pa.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(pa.RecordID[:], buf[0:])
	pa.RecordStructureID = syllab.GetUInt64(buf, 32)
	pa.RecordSize = syllab.GetUInt64(buf, 40)
	pa.WriteTime = syllab.GetInt64(buf, 48)
	copy(pa.OwnerAppID[:], buf[56:])

	copy(pa.AppInstanceID[:], buf[72:])
	copy(pa.UserConnectionID[:], buf[88:])
	copy(pa.PersonID[:], buf[104:])
	copy(pa.ReferentPersonID[:], buf[120:])
	pa.Status = PersonAuthenticationStatus(syllab.GetUInt8(buf, 136))

	copy(pa.PasswordHash[:], buf[137:])
	copy(pa.OTPPattern[:], buf[169:])
	pa.OTPAdditional = syllab.GetInt32(buf, 201)
	copy(pa.SecurityKey[:], buf[205:])
	return
}

func (pa *PersonAuthentication) syllabEncoder() (buf []byte) {
	buf = make([]byte, pa.syllabLen())

	// copy(buf[0:], pa.RecordID[:])
	syllab.SetUInt64(buf, 32, pa.RecordStructureID)
	syllab.SetUInt64(buf, 40, pa.RecordSize)
	syllab.SetInt64(buf, 48, pa.WriteTime)
	copy(buf[56:], pa.OwnerAppID[:])

	copy(buf[72:], pa.AppInstanceID[:])
	copy(buf[88:], pa.UserConnectionID[:])
	copy(buf[104:], pa.PersonID[:])
	copy(buf[120:], pa.ReferentPersonID[:])
	syllab.SetUInt8(buf, 136, uint8(pa.Status))

	copy(buf[137:], pa.PasswordHash[:])
	copy(buf[169:], pa.OTPPattern[:])
	syllab.SetInt32(buf, 201, pa.OTPAdditional)
	copy(buf[205:], pa.SecurityKey[:])
	return
}

func (pa *PersonAuthentication) syllabStackLen() (ln uint32) {
	return 237 // fixed size data + variables data add&&len
}

func (pa *PersonAuthentication) syllabHeapLen() (ln uint32) {
	return
}

func (pa *PersonAuthentication) syllabLen() (ln uint64) {
	return uint64(pa.syllabStackLen() + pa.syllabHeapLen())
}
