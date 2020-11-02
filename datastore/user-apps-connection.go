/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	"../libgo/achaemenid"
	"../libgo/authorization"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/log"
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
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	Status           UserAppsConnectionStatus
	Description      string // User custom text to identify connection easily.

	/* Connection data */
	ID     [32]byte `ganjine:"Unique" hash-index:"RecordID"`
	Weight achaemenid.Weight

	/* Peer data */
	// Peer Location
	SocietyID uint32
	RouterID  uint32
	GPAddr    [14]byte
	IPAddr    [16]byte
	ThingID   [32]byte `hash-index:"ID,ID[pair,UserID],UserID"`
	// Peer Identifiers
	UserID           [32]byte            `hash-index:"ID,ID[pair,DelegateUserID],ID[if,DelegateUserID],ThingID"`
	UserType         achaemenid.UserType `hash-index:"ID[daily]"`
	DelegateUserID   [32]byte            `hash-index:"ID,UserID"`
	DelegateUserType achaemenid.UserType

	/* Security data */
	PeerPublicKey [32]byte
	AccessControl authorization.AccessControl

	// Metrics data
	LastUsage             int64  // Last use of this connection
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
	UserAppsConnectionUnset UserAppsConnectionStatus = iota
	UserAppsConnectionIssued
	UserAppsConnectionUpdate
	UserAppsConnectionExpired
	UserAppsConnectionRevoked
)

// Set method set some data and write entire UserAppsConnection record!
func (uac *UserAppsConnection) Set() (err *er.Error) {
	uac.RecordStructureID = userAppsConnectionStructureID
	uac.RecordSize = uac.syllabLen()
	uac.WriteTime = etime.Now()
	uac.OwnerAppID = server.AppID

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
func (uac *UserAppsConnection) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID: uac.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = uac.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if uac.RecordStructureID != userAppsConnectionStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastByID find and read last version of record by given ID
func (uac *UserAppsConnection) GetLastByID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: uac.hashIDforRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	uac.RecordID = indexRes.IndexValues[0]
	err = uac.GetByRecordID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", userAppsConnectionStructureID)
	}
	return
}

// GetLastByUserIDThingID find and read last version of record by given UserID+ThingID
func (uac *UserAppsConnection) GetLastByUserIDThingID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: uac.hashUserIDThingIDforID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	uac.ID = indexRes.IndexValues[0]
	err = uac.GetLastByID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", userAppsConnectionStructureID)
	}
	return
}

// GetIDsByUserID return IDs by given UserID.
func (uac *UserAppsConnection) GetIDsByUserID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: uac.hashUserIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}
	return indexRes.IndexValues, nil
}

// GetIDsByGivenDelegate return IDs by given UserID.
func (uac *UserAppsConnection) GetIDsByGivenDelegate(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: uac.hashUserIDifDelegateUserIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}
	return indexRes.IndexValues, nil
}

// GetIDsByGottenDelegate return IDs by gotten DelegateUserID.
func (uac *UserAppsConnection) GetIDsByGottenDelegate(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: uac.hashDelegateUserIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}
	return indexRes.IndexValues, nil
}

/*
	-- PRIMARY INDEX --
*/

// IndexID index Unique-Field(ID) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (uac *UserAppsConnection) IndexID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashIDforRecordID(),
		IndexValue: uac.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashIDforRecordID() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexIDforThingID index to retrieve all Unique-Field(ID) owned by given ThingID later.
// Use in emergency to expire all connection on the Thing!
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforThingID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashThingIDforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashThingIDforID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.ThingID[:])
	return sha512.Sum512_256(buf)
}

// IndexIDforUserID index to retrieve all Unique-Field(ID) owned by given UserID.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserIDforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserIDforID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.UserID[:])
	return sha512.Sum512_256(buf)
}

// IndexIDforDelegateUserID index to retrieve all Unique-Field(ID) owned by given DelegateUserID.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforDelegateUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashDelegateUserIDforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashDelegateUserIDforID() (hash [32]byte) {
	const field = "DelegateUserID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.DelegateUserID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDforUserIDDelegateUserID index to retrieve all Unique-Field(ID) owned by given UserID + DelegateUserID
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforUserIDDelegateUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserIDDelegateUserIDforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserIDDelegateUserIDforID() (hash [32]byte) {
	const field = "UserIDDelegateUserID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.UserID[:])
	copy(buf[40:], uac.DelegateUserID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDforUserIDifDelegateUserID index to retrieve all Unique-Field(ID) owned by given UserID if DelegateUserID exist
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforUserIDifDelegateUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserIDifDelegateUserIDforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserIDifDelegateUserIDforID() (hash [32]byte) {
	const field = "IfDelegateUserID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.UserID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDforUserIDThingID index to retrieve all Unique-Field(ID) owned by given UserID+ThingID.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforUserIDThingID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserIDThingIDforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserIDThingIDforID() (hash [32]byte) {
	var buf = make([]byte, 72) // 8+32+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.UserID[:])
	copy(buf[40:], uac.ThingID[:])
	return sha512.Sum512_256(buf)
}

// IndexIDforUserTypeDaily index uac.ID that belong to UserType on daily base.
// Mostly use to index GuestType connections to research on them!
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) IndexIDforUserTypeDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserTypeDailyforID(),
		IndexValue: uac.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserTypeDailyforID() (hash [32]byte) {
	var buf = make([]byte, 17) // 8+1+8
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	syllab.SetUInt8(buf, 8, uint8(uac.UserType))
	syllab.SetInt64(buf, 9, etime.RoundToDay(uac.WriteTime))
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListThingIDforUserID list all ThingID own by specific UserID.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) ListThingIDforUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserIDforThingID(),
		IndexValue: uac.ThingID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserIDforThingID() (hash [32]byte) {
	const field = "ListThingID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.UserID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// ListUserIDforThingID store all UserID own by specific ThingID.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) ListUserIDforThingID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashThindIDforUserID(),
		IndexValue: uac.UserID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashThindIDforUserID() (hash [32]byte) {
	const field = "ListUserID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.ThingID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// ListDelegateUserIDforUserID list all DelegateUserID own by specific UserID.
// Don't call in update to an exiting record!
func (uac *UserAppsConnection) ListDelegateUserIDforUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   uac.hashUserIDforDelegateUserID(),
		IndexValue: uac.DelegateUserID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (uac *UserAppsConnection) hashUserIDforDelegateUserID() (hash [32]byte) {
	const field = "ListDelegateUserID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, userAppsConnectionStructureID)
	copy(buf[8:], uac.UserID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (uac *UserAppsConnection) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < uac.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(uac.RecordID[:], buf[0:])
	uac.RecordStructureID = syllab.GetUInt64(buf, 32)
	uac.RecordSize = syllab.GetUInt64(buf, 40)
	uac.WriteTime = syllab.GetInt64(buf, 48)
	copy(uac.OwnerAppID[:], buf[56:])

	copy(uac.AppInstanceID[:], buf[88:])
	copy(uac.UserConnectionID[:], buf[120:])
	uac.Status = UserAppsConnectionStatus(syllab.GetUInt8(buf, 152))
	uac.Description = syllab.UnsafeGetString(buf, 153)

	copy(uac.ID[:], buf[161:])
	uac.Weight = achaemenid.Weight(syllab.GetUInt8(buf, 193))

	uac.SocietyID = syllab.GetUInt32(buf, 194)
	uac.RouterID = syllab.GetUInt32(buf, 198)
	copy(uac.GPAddr[:], buf[202:])
	copy(uac.IPAddr[:], buf[216:])
	copy(uac.ThingID[:], buf[232:])

	copy(uac.UserID[:], buf[264:])
	uac.UserType = achaemenid.UserType(syllab.GetUInt8(buf, 296))
	copy(uac.DelegateUserID[:], buf[297:])
	uac.DelegateUserType = achaemenid.UserType(syllab.GetUInt8(buf, 329))

	copy(uac.PeerPublicKey[:], buf[330:])
	uac.AccessControl.SyllabDecoder(buf, 362)

	uac.LastUsage = syllab.GetInt64(buf, 362+uac.AccessControl.SyllabStackLen())
	uac.PacketPayloadSize = syllab.GetUInt16(buf, 370+uac.AccessControl.SyllabStackLen())
	uac.MaxBandwidth = syllab.GetUInt64(buf, 372+uac.AccessControl.SyllabStackLen())
	uac.ServiceCallCount = syllab.GetUInt64(buf, 380+uac.AccessControl.SyllabStackLen())
	uac.BytesSent = syllab.GetUInt64(buf, 388+uac.AccessControl.SyllabStackLen())
	uac.PacketsSent = syllab.GetUInt64(buf, 396+uac.AccessControl.SyllabStackLen())
	uac.BytesReceived = syllab.GetUInt64(buf, 404+uac.AccessControl.SyllabStackLen())
	uac.PacketsReceived = syllab.GetUInt64(buf, 412+uac.AccessControl.SyllabStackLen())
	uac.FailedPacketsReceived = syllab.GetUInt64(buf, 420+uac.AccessControl.SyllabStackLen())
	uac.FailedServiceCall = syllab.GetUInt64(buf, 428+uac.AccessControl.SyllabStackLen())
	return
}

func (uac *UserAppsConnection) syllabEncoder() (buf []byte) {
	buf = make([]byte, uac.syllabLen())
	var hsi uint32 = uac.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], uac.RecordID[:])
	syllab.SetUInt64(buf, 32, uac.RecordStructureID)
	syllab.SetUInt64(buf, 40, uac.RecordSize)
	syllab.SetInt64(buf, 48, uac.WriteTime)
	copy(buf[56:], uac.OwnerAppID[:])

	copy(buf[88:], uac.AppInstanceID[:])
	copy(buf[120:], uac.UserConnectionID[:])
	syllab.SetUInt8(buf, 152, uint8(uac.Status))
	hsi = syllab.SetString(buf, uac.Description, 153, hsi)

	copy(buf[161:], uac.ID[:])
	syllab.SetUInt8(buf, 193, uint8(uac.Weight))

	syllab.SetUInt32(buf, 194, uac.SocietyID)
	syllab.SetUInt32(buf, 198, uac.RouterID)
	copy(buf[202:], uac.GPAddr[:])
	copy(buf[216:], uac.IPAddr[:])
	copy(buf[232:], uac.ThingID[:])

	copy(buf[264:], uac.UserID[:])
	syllab.SetUInt8(buf, 296, uint8(uac.UserType))
	copy(buf[297:], uac.DelegateUserID[:])
	syllab.SetUInt8(buf, 329, uint8(uac.DelegateUserType))

	copy(buf[330:], uac.PeerPublicKey[:])
	uac.AccessControl.SyllabEncoder(buf, 362, hsi)

	syllab.SetInt64(buf, 362+uac.AccessControl.SyllabStackLen(), uac.LastUsage)
	syllab.SetUInt16(buf, 370+uac.AccessControl.SyllabStackLen(), uac.PacketPayloadSize)
	syllab.SetUInt64(buf, 372+uac.AccessControl.SyllabStackLen(), uac.MaxBandwidth)
	syllab.SetUInt64(buf, 380+uac.AccessControl.SyllabStackLen(), uac.ServiceCallCount)
	syllab.SetUInt64(buf, 388+uac.AccessControl.SyllabStackLen(), uac.BytesSent)
	syllab.SetUInt64(buf, 396+uac.AccessControl.SyllabStackLen(), uac.PacketsSent)
	syllab.SetUInt64(buf, 404+uac.AccessControl.SyllabStackLen(), uac.BytesReceived)
	syllab.SetUInt64(buf, 412+uac.AccessControl.SyllabStackLen(), uac.PacketsReceived)
	syllab.SetUInt64(buf, 420+uac.AccessControl.SyllabStackLen(), uac.FailedPacketsReceived)
	syllab.SetUInt64(buf, 428+uac.AccessControl.SyllabStackLen(), uac.FailedServiceCall)
	return
}

func (uac *UserAppsConnection) syllabStackLen() (ln uint32) {
	return 436 + uac.AccessControl.SyllabStackLen()
}

func (uac *UserAppsConnection) syllabHeapLen() (ln uint32) {
	ln += uint32(len(uac.Description))
	ln += uac.AccessControl.SyllabHeapLen()
	return
}

func (uac *UserAppsConnection) syllabLen() (ln uint64) {
	return uint64(uac.syllabStackLen() + uac.syllabHeapLen())
}
