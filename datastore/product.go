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
	productStructureID uint64 = 2911412040892278928
)

var productStructure = ganjine.DataStructure{
	ID:                2911412040892278928,
	IssueDate:         1599279351,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         Product{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Product",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "store product(goods) details!",
	},
	TAGS: []string{
		"",
	},
}

// Product ---Read locale description in productStructure---
type Product struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID        [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID     [32]byte // Store to remember which user connection set||chanaged this record!
	ID                   [32]byte `ganjine:"Immutable,Unique" ganjine-index:"OwnerID,SellerID,WikiID,WikiID-DistributionCenterID[temp]"` // ProductID
	OwnerID              [32]byte // Who belong to! Just org can be first owner!
	SellerID             [32]byte // OrdererID, who places the order usually use for prescription(drug order) or sales agent!
	WikiID               [32]byte `ganjine:"Immutable"`
	ProductionID         [32]byte // It can also upper ID that this product split from it!
	DistributionCenterID [32]byte `ganjine-list:"WikiID[temp]"` // It will changed only on owner changed otherwise can be mobile location! if != Sale.WarehouseID means user wants to send item to this address!
	ProductAuctionID     [32]byte // can be 0 for just change owner without any auction or price but very rare situation!
	Status               ProductStatus
}

// Set method set some data and write entire Product record!
func (p *Product) Set() (err *er.Error) {
	p.RecordStructureID = productStructureID
	p.RecordSize = p.syllabLen()
	p.WriteTime = etime.Now()
	p.OwnerAppID = server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: p.syllabEncoder(),
	}
	p.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], p.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (p *Product) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID: p.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = p.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if p.RecordStructureID != productStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByID find and read last version of record by given ID
func (p *Product) GetLastByID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: p.hashIDforRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	p.RecordID = indexRes.IndexValues[0]
	err = p.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", productStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index p.ID to retrieve record fast later.
func (p *Product) IndexID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashIDforRecordID(),
		IndexValue: p.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashIDforRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexOwnerDaily index to retrieve all ID owned by given p.Owner later in daily.
func (p *Product) IndexOwnerDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashOwnerIDforIDDaily(),
		IndexValue: p.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashOwnerIDforIDDaily() (hash [32]byte) {
	var buf = make([]byte, 48) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.OwnerID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	return sha512.Sum512_256(buf)
}

// IndexSellerDaily index to retrieve all ID owned by given p.Seller later in daily.
func (p *Product) IndexSellerDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashSellerIDforIDDaily(),
		IndexValue: p.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashSellerIDforIDDaily() (hash [32]byte) {
	var buf = make([]byte, 48) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.SellerID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	return sha512.Sum512_256(buf)
}

// IndexWikiDaily index to retrieve all ID owned by given p.Wiki later in daily.
func (p *Product) IndexWikiDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashWikiIDforIDDaily(),
		IndexValue: p.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashWikiIDforIDDaily() (hash [32]byte) {
	var buf = make([]byte, 48) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.WikiID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	return sha512.Sum512_256(buf)
}

// IndexWikiDCDaily index to retrieve all ID owned by given p.Wiki + p.DistributionCenterID later.
// Each year is 365 days that indicate we have 365 index record each year per product wiki on each distribution center!
func (p *Product) IndexWikiDCDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashWikiIDDistributionCenterIDforIDDaily(),
		IndexValue: p.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashWikiIDDistributionCenterIDforIDDaily() (hash [32]byte) {
	var buf = make([]byte, 80) // 8+32+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.WikiID[:])
	copy(buf[40:], p.DistributionCenterID[:])
	syllab.SetInt64(buf, 72, p.WriteTime.RoundToDay())
	return sha512.Sum512_256(buf)
}

// IndexProductAuction index to retrieve all ID owned by given p.AuctionID later.
// Use to indiacate product sell by specific auction.
// Don't call in update to an exiting record!
func (p *Product) IndexProductAuction() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashAuctionIDforID(),
		IndexValue: p.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashAuctionIDforID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.ProductAuctionID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

/*
	-- Temporary INDEXES & LIST --
*/

// TempIndexWikiDC index to retrieve all ID owned by given p.WikiID + p.DistributionCenterID later.
// Use to indiacate product stock in the DC. temp index to Pop from it!
func (p *Product) TempIndexWikiDC() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashWikiIDDistributionCenterIDforID(),
		IndexValue: p.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashWikiIDDistributionCenterIDforID() (hash [32]byte) {
	const field = "TempWikiID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.WikiID[:])
	copy(buf[40:], p.DistributionCenterID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// TempListWikiDC store all p.DistributionCenterID related to specific p.WikiID.
// Use to indiacate global product stock. temp index to delete from it!
// Don't call in update to an exiting record!
func (p *Product) TempListWikiDC() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashWikiIDforDistributionCenterID(),
		IndexValue: p.DistributionCenterID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashWikiIDforDistributionCenterID() (hash [32]byte) {
	const field = "TempListWikiID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.WikiID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (p *Product) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < p.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(p.RecordID[:], buf[0:])
	p.RecordStructureID = syllab.GetUInt64(buf, 32)
	p.RecordSize = syllab.GetUInt64(buf, 40)
	p.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(p.OwnerAppID[:], buf[56:])

	copy(p.AppInstanceID[:], buf[88:])
	copy(p.UserConnectionID[:], buf[120:])
	copy(p.ID[:], buf[152:])
	copy(p.OwnerID[:], buf[184:])
	copy(p.SellerID[:], buf[216:])
	copy(p.WikiID[:], buf[248:])
	copy(p.ProductionID[:], buf[280:])
	copy(p.DistributionCenterID[:], buf[312:])
	copy(p.ProductAuctionID[:], buf[344:])
	p.Status = ProductStatus(syllab.GetUInt8(buf, 376))
	return
}

func (p *Product) syllabEncoder() (buf []byte) {
	buf = make([]byte, p.syllabLen())

	// copy(buf[0:], p.RecordID[:])
	syllab.SetUInt64(buf, 32, p.RecordStructureID)
	syllab.SetUInt64(buf, 40, p.RecordSize)
	syllab.SetInt64(buf, 48, int64(p.WriteTime))
	copy(buf[56:], p.OwnerAppID[:])

	copy(buf[88:], p.AppInstanceID[:])
	copy(buf[120:], p.UserConnectionID[:])
	copy(buf[152:], p.ID[:])
	copy(buf[184:], p.OwnerID[:])
	copy(buf[216:], p.SellerID[:])
	copy(buf[248:], p.WikiID[:])
	copy(buf[280:], p.ProductionID[:])
	copy(buf[312:], p.DistributionCenterID[:])
	copy(buf[344:], p.ProductAuctionID[:])
	syllab.SetUInt8(buf, 376, uint8(p.Status))
	return
}

func (p *Product) syllabStackLen() (ln uint32) {
	return 377
}

func (p *Product) syllabHeapLen() (ln uint32) {
	return
}

func (p *Product) syllabLen() (ln uint64) {
	return uint64(p.syllabStackLen() + p.syllabHeapLen())
}

/*
	-- Record types --
*/

// ProductStatus indicate Product record status
type ProductStatus uint8

// Product status
const (
	ProductCreated ProductStatus = iota
	ProductVoid
	ProductPreSale // use in budget analysis and also can be trade!
	// 0x0 for non expire record, 0x1 for sell to first above SuggestPrice||buy first below it!

	// Split from upper size product!
// Split to small size product!

// ManagerApprove
// WarehouseApprove
// ActiveWithApprove
// ActiveWithoutApprove
// ActiveWithOrderer
// Inactive
// Normal
// Reject
)
