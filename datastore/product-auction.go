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
	"../libgo/math"
	"../libgo/pehrest"
	psdk "../libgo/pehrest-sdk"
	"../libgo/syllab"
)

const (
	productAuctionStructureID uint64 = 14218566705593294138
)

var productAuctionStructure = ganjine.DataStructure{
	ID:                14218566705593294138,
	IssueDate:         1599286551,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         ProductAuction{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Product Auction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Store the product auction data",
	},
	TAGS: []string{
		"",
	},
}

// ProductAuction ---Read locale description in productAuctionStructure---
type ProductAuction struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	OrgID            [32]byte `index-hash:"ID"` // Sell can be register just by producer organization
	ID               [32]byte `index-hash:"RecordID"`
	QuiddityID       [32]byte `index-hash:"ID,ID[pair,Authorization.AllowUserID],ID[pair,Authorization.GroupID]"`

	// Price
	Discount         math.PerMyriad // minus from product price
	DCCommission     math.PerMyriad // Distribution-Center(Org)
	SellerCommission math.PerMyriad

	// Authorization
	Authorization authorization.Product

	Description string // User custom text to identify Product Auction easily by each other for org and other users.
	Type        ProductAuctionType
	Status      ProductAuctionStatus
}

// SaveNew method set some data and write entire ProductAuction record with all indexes!
func (pa *ProductAuction) SaveNew() (err *er.Error) {
	err = pa.Set()
	if err != nil {
		return
	}

	pa.IndexRecordIDForID()
	pa.IndexIDForOrgID()
	pa.IndexIDForQuiddityID()
	if pa.Authorization.AllowUserID != [32]byte{} {
		pa.IndexIDForQuiddityIDAllowUserID()
		pa.IndexIDForAllowUserID()
	}
	if pa.Authorization.GroupID != [32]byte{} {
		pa.IndexIDForQuiddityIDGroupID()
		pa.IndexIDForGroupID()
	}
	return
}

// Set method set some data and write entire ProductAuction record!
func (pa *ProductAuction) Set() (err *er.Error) {
	pa.RecordStructureID = productAuctionStructureID
	pa.RecordSize = pa.syllabLen()
	pa.WriteTime = etime.Now()
	pa.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pa.syllabEncoder(),
	}
	pa.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pa.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pa *ProductAuction) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          pa.RecordID,
		RecordStructureID: productAuctionStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = pa.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if pa.RecordStructureID != productAuctionStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByID method find and read last version of record by given pa.ID
func (pa *ProductAuction) GetLastByID() (err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashIDForRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	if err != nil {
		return
	}

	pa.RecordID = indexRes.IndexValues[0]
	err = pa.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", productAuctionStructureID)
	}
	return
}

/*
	-- Search Methods --
*/

// FindRecordsIDsByID find RecordsIDs by given ID
func (pa *ProductAuction) FindRecordsIDsByID(offset, limit uint64) (RecordsIDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashIDForRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	RecordsIDs = indexRes.IndexValues
	return
}

// FindIDsByOrgID find IDs by given OrgID
func (pa *ProductAuction) FindIDsByOrgID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashOrgIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByQuiddityID find IDs by given QuiddityID
func (pa *ProductAuction) FindIDsByQuiddityID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashQuiddityIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByQuiddityIDAllowUserID find IDs by given QuiddityID+AllowUserID
func (pa *ProductAuction) FindIDsByQuiddityIDAllowUserID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashQuiddityIDAllowUserIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByQuiddityIDGroupID find IDs by given QuiddityID+GroupID
func (pa *ProductAuction) FindIDsByQuiddityIDGroupID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashQuiddityIDGroupIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByAllowUserID find IDs by given AllowUserID
func (pa *ProductAuction) FindIDsByAllowUserID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashAllowUserIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByGroupID find IDs by given GroupID
func (pa *ProductAuction) FindIDsByGroupID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pa.hashGroupIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexRecordIDForID save RecordID chain for ID
func (pa *ProductAuction) IndexRecordIDForID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashIDForRecordID(),
		IndexValue: pa.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashIDForRecordID() (hash [32]byte) {
	const field = "ID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.ID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexIDForOrgID save ID chain for OrgID
func (pa *ProductAuction) IndexIDForOrgID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashOrgIDForID(),
		IndexValue: pa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashOrgIDForID() (hash [32]byte) {
	const field = "OrgID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.OrgID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForQuiddityID save ID chain for QuiddityID
func (pa *ProductAuction) IndexIDForQuiddityID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashQuiddityIDForID(),
		IndexValue: pa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashQuiddityIDForID() (hash [32]byte) {
	const field = "QuiddityID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.QuiddityID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForQuiddityIDAllowUserID save ID chain for QuiddityID+AllowUserID
func (pa *ProductAuction) IndexIDForQuiddityIDAllowUserID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashQuiddityIDAllowUserIDForID(),
		IndexValue: pa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashQuiddityIDAllowUserIDForID() (hash [32]byte) {
	const field = "QuiddityIDAllowUserID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.QuiddityID[:])
	copy(buf[40:], pa.Authorization.AllowUserID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForQuiddityIDGroupID save ID chain for QuiddityID+GroupID
func (pa *ProductAuction) IndexIDForQuiddityIDGroupID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashQuiddityIDGroupIDForID(),
		IndexValue: pa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashQuiddityIDGroupIDForID() (hash [32]byte) {
	const field = "QuiddityIDGroupID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.QuiddityID[:])
	copy(buf[40:], pa.Authorization.GroupID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForAllowUserID save ID chain for AllowUserID
func (pa *ProductAuction) IndexIDForAllowUserID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashAllowUserIDForID(),
		IndexValue: pa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashAllowUserIDForID() (hash [32]byte) {
	const field = "AllowUserID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.Authorization.AllowUserID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForGroupID save ID chain for GroupID
func (pa *ProductAuction) IndexIDForGroupID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashGroupIDForID(),
		IndexValue: pa.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashGroupIDForID() (hash [32]byte) {
	const field = "GroupID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.Authorization.GroupID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (pa *ProductAuction) syllabDecoder(buf []byte) (err *er.Error) {
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
	copy(pa.OrgID[:], buf[152:])
	copy(pa.ID[:], buf[184:])
	copy(pa.QuiddityID[:], buf[216:])

	pa.Discount = math.PerMyriad(syllab.GetUInt16(buf, 248))
	pa.DCCommission = math.PerMyriad(syllab.GetUInt16(buf, 250))
	pa.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 252))

	pa.Authorization.SyllabDecoder(buf, 254)

	pa.Description = syllab.UnsafeGetString(buf, 254+pa.Authorization.SyllabStackLen())
	pa.Type = ProductAuctionType(syllab.GetUInt8(buf, 262+pa.Authorization.SyllabStackLen()))
	pa.Status = ProductAuctionStatus(syllab.GetUInt8(buf, 263+pa.Authorization.SyllabStackLen()))
	return
}

func (pa *ProductAuction) syllabEncoder() (buf []byte) {
	buf = make([]byte, pa.syllabLen())
	var hsi uint32 = pa.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], pa.RecordID[:])
	syllab.SetUInt64(buf, 32, pa.RecordStructureID)
	syllab.SetUInt64(buf, 40, pa.RecordSize)
	syllab.SetInt64(buf, 48, int64(pa.WriteTime))
	copy(buf[56:], pa.OwnerAppID[:])

	copy(buf[88:], pa.AppInstanceID[:])
	copy(buf[120:], pa.UserConnectionID[:])
	copy(buf[152:], pa.OrgID[:])
	copy(buf[184:], pa.ID[:])
	copy(buf[216:], pa.QuiddityID[:])

	syllab.SetUInt16(buf, 248, uint16(pa.Discount))
	syllab.SetUInt16(buf, 250, uint16(pa.DCCommission))
	syllab.SetUInt16(buf, 252, uint16(pa.SellerCommission))

	hsi = pa.Authorization.SyllabEncoder(buf, 254, hsi)

	hsi = syllab.SetString(buf, pa.Description, 254+pa.Authorization.SyllabStackLen(), hsi)
	syllab.SetUInt8(buf, 262+pa.Authorization.SyllabStackLen(), uint8(pa.Type))
	syllab.SetUInt8(buf, 263+pa.Authorization.SyllabStackLen(), uint8(pa.Status))
	return
}

func (pa *ProductAuction) syllabStackLen() (ln uint32) {
	return 264 + pa.Authorization.SyllabStackLen()
}

func (pa *ProductAuction) syllabHeapLen() (ln uint32) {
	ln += pa.Authorization.SyllabHeapLen()
	ln += uint32(len(pa.Description))
	return
}

func (pa *ProductAuction) syllabLen() (ln uint64) {
	return uint64(pa.syllabStackLen() + pa.syllabHeapLen())
}

/*
	-- Record types --
*/

// ProductAuctionType indicate ProductAuction record type
// https://en.wikipedia.org/wiki/Auction_theory
type ProductAuctionType uint8

// ProductAuctionStatus indicate ProductAuction record status
type ProductAuctionStatus uint8

// ProductAuction status
const (
	ProductAuctionUnset ProductAuctionStatus = iota
	ProductAuctionRegistered
	ProductAuctionUpdated
	ProductAuctionExpired
	ProductAuctionBlocked
)
