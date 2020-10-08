/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	"../libgo/authorization"
	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/syllab"
)

/*
Connection Authorization service
We have 2 types access to real user data.

Relations Access
Relations of users have many kind like friends(in many scope), followers, family(in many grade), ... . Relations have fixed types. These relations knowledge store in Relations MS.

Token Access
Users can make authentication token and give it to third party to access his/her data. So to restricted access, User can be set rules for tokens to limit to specific rules.
*/

const (
	userAppsConnectionStructureID uint64 = 5222171135412713418
)

var userAppsConnectionStructure = ganjine.DataStructure{
	ID:                5222171135412713418,
	IssueDate:         1601307949,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         UserAppsConnection{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "UserAppsConnection",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "Use by any type users itself or delegate to other users to act as the owner!",
	},
	TAGS: []string{
		"",
	},
}

// UserAppsConnection ---Read locale description in userAppsConnectionStructure---
type UserAppsConnection struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	ID               [16]byte `ganjine:"Immutable,Unique" ganjine-index:"OwnerID,DelegateUserID,AppID,ThingID,OwnerID-AppID-ThingID,OwnerID-AppID,OwnerID-ThingID"` // UserConnectionID
	OwnerID          [16]byte `ganjine:"Immutable" ganjine-list:"DelegateUserID,AppID,ThingID"`                                                                     // Owner User ID
	DelegateUserID   [16]byte `ganjine:"Immutable"`
	AppID            [16]byte `ganjine:"Immutable"`
	ThingID          [16]byte `ganjine:"Immutable"`
	Description      string   // User custom text to identify connection easily.
	AccessControl    authorization.AccessControl
	Status           UserAppsConnectionStatus

	// Metrics data
	PacketPayloadSize     uint16 // Always must respect max frame size, so usually packets can't be more than 8192Byte!
	MaxBandwidth          uint64 // Peer must respect this, otherwise connection will terminate and GP go to black list!
	ServiceCallCount      uint64 // Count successful or unsuccessful request.
	BytesSent             uint64 // Counts the bytes of payload data sent.
	PacketsSent           uint64 // Counts packets sent.
	BytesReceived         uint64 // Counts the bytes of payload data Receive.
	PacketsReceived       uint64 // Counts packets Receive.
	FailedPacketsReceived uint64 // Counts failed packets receive for firewalling server from some attack types!
	FailedServiceCall     uint64 // Counts failed service call e.g. data validation failed, ...
}

// UserAppsConnectionStatus use to indicate UserAppsConnection record status
type UserAppsConnectionStatus uint8

// UserAppsConnection status
const (
	// UserAppsConnectionRevoked indicate related public key was inactivated by person
	UserAppsConnectionIssued UserAppsConnectionStatus = iota
	UserAppsConnectionExpired
	UserAppsConnectionRevoked
)

// Set method set some data and write entire UserAppsConnection record!
func (uac *UserAppsConnection) Set() (err error) {
	uac.RecordStructureID = userAppsConnectionStructureID
	uac.RecordSize = uac.syllabLen()
	uac.WriteTime = etime.Now()
	uac.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: uac.syllabEncoder(),
	}
	uac.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], uac.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (uac *UserAppsConnection) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: uac.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = uac.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if uac.RecordStructureID != userAppsConnectionStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastByID method find and read last version of record by given ID
func (uac *UserAppsConnection) GetLastByID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: uac.HashID(),
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
		uac.RecordID = indexRes.RecordIDs[ln]
		err = uac.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

// GetLastByOwnerIDAppIDThingID method find and read last version of record by given OwnerID+AppID+ThingID
func (uac *UserAppsConnection) GetLastByOwnerIDAppIDThingID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: uac.HashOwnerIDAppIDThingID(),
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
		uac.RecordID = indexRes.RecordIDs[ln]
		err = uac.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index Unique-Field(ID) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (uac *UserAppsConnection) IndexID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashID(),
		RecordID:  uac.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashID hash userAppsConnectionStructureID + uac.ID
func (uac *UserAppsConnection) HashID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexOwner index to retrieve all Unique-Field(ID) owned by given OwnerID later.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexOwner() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerID(),
	}
	copy(indexRequest.RecordID[:], uac.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerID hash userAppsConnectionStructureID + uac.OwnerID
func (uac *UserAppsConnection) HashOwnerID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	return sha512.Sum512_256(buf)
}

// IndexDelegateUser index to retrieve all Unique-Field(ID) owned by given DelegateUserID later.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexDelegateUser() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashDelegateUserID(),
	}
	copy(indexRequest.RecordID[:], uac.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashDelegateUserID hash userAppsConnectionStructureID + uac.DelegateUserID
func (uac *UserAppsConnection) HashDelegateUserID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.DelegateUserID[:])
	return sha512.Sum512_256(buf)
}

// IndexOwnerAppThing index to retrieve all Unique-Field(ID) owned by given OwnerID+AppID+ThingID later.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexOwnerAppThing() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerIDAppIDThingID(),
	}
	copy(indexRequest.RecordID[:], uac.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerIDAppIDThingID hash userAppsConnectionStructureID + uac.OwnerID + uac.AppID + uac.ThingID
func (uac *UserAppsConnection) HashOwnerIDAppIDThingID() (hash [32]byte) {
	var buf = make([]byte, 56) // 8+16+16+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	copy(buf[24:], uac.AppID[:])
	copy(buf[40:], uac.ThingID[:])
	return sha512.Sum512_256(buf)
}

// IndexOwnerApp index to retrieve all Unique-Field(ID) owned by given OwnerID+AppID later.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexOwnerApp() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerIDAppID(),
	}
	copy(indexRequest.RecordID[:], uac.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerIDAppID hash userAppsConnectionStructureID + uac.OwnerID + uac.AppID
func (uac *UserAppsConnection) HashOwnerIDAppID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+16+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	copy(buf[24:], uac.AppID[:])
	return sha512.Sum512_256(buf)
}

// IndexOwnerThing index to retrieve all Unique-Field(ID) owned by given OwnerID+ThingID later.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexOwnerThing() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerIDThingID(),
	}
	copy(indexRequest.RecordID[:], uac.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerIDThingID hash userAppsConnectionStructureID + uac.OwnerID + uac.ThingID
func (uac *UserAppsConnection) HashOwnerIDThingID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+16+16+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	copy(buf[24:], uac.ThingID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListOwnerDelegate store all DelegateUserID own by specific Owner.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) ListOwnerDelegate() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerDelegateField(),
	}
	copy(indexRequest.RecordID[:], uac.OwnerID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerDelegateField hash userAppsConnectionStructureID + OwnerID + "DelegateUserID" field
func (uac *UserAppsConnection) HashOwnerDelegateField() (hash [32]byte) {
	const field = "DelegateUserID"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

// ListOwnerApp store all AppID own by specific Owner.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) ListOwnerApp() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerAppField(),
	}
	copy(indexRequest.RecordID[:], uac.OwnerID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerAppField hash userAppsConnectionStructureID + OwnerID + "AppID" field
func (uac *UserAppsConnection) HashOwnerAppField() (hash [32]byte) {
	const field = "AppID"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

// ListOwnerThing store all ThingID own by specific Owner.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) ListOwnerThing() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: uac.HashOwnerThingField(),
	}
	copy(indexRequest.RecordID[:], uac.OwnerID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerThingField hash userAppsConnectionStructureID + OwnerID + "ThingID" field
func (uac *UserAppsConnection) HashOwnerThingField() (hash [32]byte) {
	const field = "ThingID"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.OwnerID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (uac *UserAppsConnection) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < uac.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(uac.RecordID[:], buf[0:])
	uac.RecordStructureID = syllab.GetUInt64(buf, 32)
	uac.RecordSize = syllab.GetUInt64(buf, 40)
	uac.WriteTime = syllab.GetInt64(buf, 48)
	copy(uac.OwnerAppID[:], buf[56:])

	copy(uac.AppInstanceID[:], buf[72:])
	copy(uac.UserConnectionID[:], buf[88:])
	copy(uac.ID[:], buf[104:])
	copy(uac.OwnerID[:], buf[120:])
	copy(uac.DelegateUserID[:], buf[136:])
	copy(uac.AppID[:], buf[152:])
	copy(uac.ThingID[:], buf[168:])
	uac.Description = syllab.UnsafeGetString(buf, 184)
	uac.AccessControl.SyllabDecoder(buf, 192)
	uac.Status = UserAppsConnectionStatus(syllab.GetUInt8(buf, 192+uac.AccessControl.SyllabStackLen()))

	uac.PacketPayloadSize = syllab.GetUInt16(buf, 193+uac.AccessControl.SyllabStackLen())
	uac.MaxBandwidth = syllab.GetUInt64(buf, 195+uac.AccessControl.SyllabStackLen())
	uac.ServiceCallCount = syllab.GetUInt64(buf, 203+uac.AccessControl.SyllabStackLen())
	uac.BytesSent = syllab.GetUInt64(buf, 211+uac.AccessControl.SyllabStackLen())
	uac.PacketsSent = syllab.GetUInt64(buf, 219+uac.AccessControl.SyllabStackLen())
	uac.BytesReceived = syllab.GetUInt64(buf, 227+uac.AccessControl.SyllabStackLen())
	uac.PacketsReceived = syllab.GetUInt64(buf, 235+uac.AccessControl.SyllabStackLen())
	uac.FailedPacketsReceived = syllab.GetUInt64(buf, 243+uac.AccessControl.SyllabStackLen())
	uac.FailedServiceCall = syllab.GetUInt64(buf, 251+uac.AccessControl.SyllabStackLen())
	return
}

func (uac *UserAppsConnection) syllabEncoder() (buf []byte) {
	buf = make([]byte, uac.syllabLen())
	var hsi uint32 = uac.syllabStackLen() // Heap start index || Stack size!
	var ln uint32                         // len of strings, slices, maps, ...

	// copy(buf[0:], uac.RecordID[:])
	syllab.SetUInt64(buf, 32, uac.RecordStructureID)
	syllab.SetUInt64(buf, 40, uac.RecordSize)
	syllab.SetInt64(buf, 48, uac.WriteTime)
	copy(buf[56:], uac.OwnerAppID[:])

	copy(buf[72:], uac.AppInstanceID[:])
	copy(buf[88:], uac.UserConnectionID[:])
	copy(buf[104:], uac.ID[:])
	copy(buf[120:], uac.OwnerID[:])
	copy(buf[136:], uac.DelegateUserID[:])
	copy(buf[152:], uac.AppID[:])
	copy(buf[168:], uac.ThingID[:])
	ln = uint32(len(uac.Description))
	syllab.SetUInt32(buf, 184, hsi)
	syllab.SetUInt32(buf, 188, ln)
	copy(buf[hsi:], uac.Description)
	hsi += ln
	hsi = uac.AccessControl.SyllabEncoder(buf, 192, hsi)
	syllab.SetUInt8(buf, 192+uac.AccessControl.SyllabStackLen(), uint8(uac.Status))

	syllab.SetUInt16(buf, 193+uac.AccessControl.SyllabStackLen(), uac.PacketPayloadSize)
	syllab.SetUInt64(buf, 195+uac.AccessControl.SyllabStackLen(), uac.MaxBandwidth)
	syllab.SetUInt64(buf, 203+uac.AccessControl.SyllabStackLen(), uac.ServiceCallCount)
	syllab.SetUInt64(buf, 211+uac.AccessControl.SyllabStackLen(), uac.BytesSent)
	syllab.SetUInt64(buf, 219+uac.AccessControl.SyllabStackLen(), uac.PacketsSent)
	syllab.SetUInt64(buf, 227+uac.AccessControl.SyllabStackLen(), uac.BytesReceived)
	syllab.SetUInt64(buf, 235+uac.AccessControl.SyllabStackLen(), uac.PacketsReceived)
	syllab.SetUInt64(buf, 243+uac.AccessControl.SyllabStackLen(), uac.FailedPacketsReceived)
	syllab.SetUInt64(buf, 251+uac.AccessControl.SyllabStackLen(), uac.FailedServiceCall)
	return
}

func (uac *UserAppsConnection) syllabStackLen() (ln uint32) {
	return 259 + uac.AccessControl.SyllabStackLen() // fixed size data + variables data add&&len
}

func (uac *UserAppsConnection) syllabHeapLen() (ln uint32) {
	ln += uint32(len(uac.Description))
	ln += uac.AccessControl.SyllabHeapLen()
	return
}

func (uac *UserAppsConnection) syllabLen() (ln uint64) {
	return uint64(uac.syllabStackLen() + uac.syllabHeapLen())
}
