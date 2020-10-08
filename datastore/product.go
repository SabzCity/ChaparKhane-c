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
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID        [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID     [16]byte // Store to remember which user connection set||chanaged this record!
	ID                   [16]byte `ganjine:"Immutable,Unique" ganjine-index:"OwnerID,SellerID,WikiID,WikiID-DistributionCenterID[temp]"` // ProductID
	OwnerID              [16]byte // Who belong to! Just org can be first owner!
	SellerID             [16]byte // OrdererID, who places the order usually use for prescription(drug order) or sales agent!
	WikiID               [16]byte `ganjine:"Immutable"`
	ProductionID         [16]byte // It can also upper ID that this product split from it!
	DistributionCenterID [16]byte `ganjine-list:"WikiID[temp]"` // It will changed only on owner changed otherwise can be mobile location! if != Sale.WarehouseID means user wants to send item to this address!
	ProductAuctionID     [16]byte // can be 0 for just change owner without any auction or price but very rare situation!
	Status               ProductStatus
}

// ProductStatus indicate Product record status
type ProductStatus uint8

// Product status
const (
	ProductCreated ProductStatus = iota
	ProductVoid
	ProductPreSale // ProductPreSale use in budget analysis and also can be trade!

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

// Set method set some data and write entire Product record!
func (p *Product) Set() (err error) {
	p.RecordStructureID = productStructureID
	p.RecordSize = p.syllabLen()
	p.WriteTime = etime.Now()
	p.OwnerAppID = server.Manifest.AppID

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
func (p *Product) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: p.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = p.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if p.RecordStructureID != productStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByID method find and read last version of record by given ID
func (p *Product) GetByID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: p.HashID(),
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
		p.RecordID = indexRes.RecordIDs[ln]
		err = p.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index p.ID to retrieve record fast later.
func (p *Product) IndexID() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashID(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashID hash productStructureID + p.ID
func (p *Product) HashID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexOwnerDaily index to retrieve all ID owned by given p.Owner later in daily.
func (p *Product) IndexOwnerDaily() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashOwnerDaily(),
	}
	copy(indexRequest.RecordID[:], p.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerDaily hash productStructureID + p.OwnerID + p.WriteTime(round to day)
func (p *Product) HashOwnerDaily() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+8+16
	syllab.SetUInt64(buf, 0, productStructureID)
	syllab.SetInt64(buf, 8, etime.RoundToDay(p.WriteTime))
	copy(buf[16:], p.OwnerID[:])
	return sha512.Sum512_256(buf)
}

// IndexSellerDaily index to retrieve all ID owned by given p.Seller later in daily.
func (p *Product) IndexSellerDaily() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashSellerDaily(),
	}
	copy(indexRequest.RecordID[:], p.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashSellerDaily hash productStructureID + p.SellerID + p.WriteTime(round to day)
func (p *Product) HashSellerDaily() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+8+16
	syllab.SetUInt64(buf, 0, productStructureID)
	syllab.SetInt64(buf, 8, etime.RoundToDay(p.WriteTime))
	copy(buf[16:], p.SellerID[:])
	return sha512.Sum512_256(buf)
}

// IndexWikiDaily index to retrieve all ID owned by given p.Wiki later in daily.
func (p *Product) IndexWikiDaily() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiDaily(),
	}
	copy(indexRequest.RecordID[:], p.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashWikiDaily hash productStructureID + p.WikiID + p.WriteTime(round to day)
func (p *Product) HashWikiDaily() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+8+16
	syllab.SetUInt64(buf, 0, productStructureID)
	syllab.SetInt64(buf, 8, etime.RoundToDay(p.WriteTime))
	copy(buf[16:], p.WikiID[:])
	return sha512.Sum512_256(buf)
}

// IndexWikiDCDaily index to retrieve all ID owned by given p.Wiki + p.DistributionCenterID later.
// Each year is 365 days that indicate we have 365 index record each year per product wiki on each distribution center!
func (p *Product) IndexWikiDCDaily() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiDCDaily(),
	}
	copy(indexRequest.RecordID[:], p.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashWikiDCDaily hash productStructureID+ p.WikiID + p.DistributionCenterID + p.WriteTime(round to day)
func (p *Product) HashWikiDCDaily() (hash [32]byte) {
	var buf = make([]byte, 48) // 8+8+16+16
	syllab.SetUInt64(buf, 0, productStructureID)
	syllab.SetInt64(buf, 8, etime.RoundToDay(p.WriteTime))
	copy(buf[16:], p.WikiID[:])
	copy(buf[32:], p.DistributionCenterID[:])
	return sha512.Sum512_256(buf)
}

// IndexProductAuction index to retrieve all ID owned by given p.AuctionID later.
// Use to indiacate product sell by specific auction.
// Don't call in update to an exiting record!
func (p *Product) IndexProductAuction() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashAuctionID(),
	}
	copy(indexRequest.RecordID[:], p.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashAuctionID hash productStructureID + p.ProductAuctionID
func (p *Product) HashAuctionID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
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
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiDC(),
	}
	copy(indexRequest.RecordID[:], p.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashWikiDC hash productStructureID + p.WikiID + p.DistributionCenterID
func (p *Product) HashWikiDC() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+16+16
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.WikiID[:])
	copy(buf[24:], p.DistributionCenterID[:])
	return sha512.Sum512_256(buf)
}

// TempListWikiDC store all p.DistributionCenterID related to specific p.WikiID.
// Use to indiacate global product stock. temp index to delete from it!
// Don't call in update to an exiting record!
func (p *Product) TempListWikiDC() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiField(),
	}
	copy(indexRequest.RecordID[:], p.DistributionCenterID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashWikiField hash productStructureID + p.WikiID + "WikiID" field
func (p *Product) HashWikiField() (hash [32]byte) {
	const field = "WikiID"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.WikiID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (p *Product) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < p.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(p.RecordID[:], buf[0:])
	p.RecordStructureID = syllab.GetUInt64(buf, 32)
	p.RecordSize = syllab.GetUInt64(buf, 40)
	p.WriteTime = syllab.GetInt64(buf, 48)
	copy(p.OwnerAppID[:], buf[56:])

	copy(p.AppInstanceID[:], buf[72:])
	copy(p.UserConnectionID[:], buf[88:])
	copy(p.ID[:], buf[104:])
	copy(p.OwnerID[:], buf[120:])
	copy(p.SellerID[:], buf[136:])
	copy(p.WikiID[:], buf[152:])
	copy(p.ProductionID[:], buf[168:])
	copy(p.DistributionCenterID[:], buf[184:])
	copy(p.ProductAuctionID[:], buf[200:])
	p.Status = ProductStatus(syllab.GetUInt8(buf, 216))
	return
}

func (p *Product) syllabEncoder() (buf []byte) {
	buf = make([]byte, p.syllabLen())

	// copy(buf[0:], p.RecordID[:])
	syllab.SetUInt64(buf, 32, p.RecordStructureID)
	syllab.SetUInt64(buf, 40, p.RecordSize)
	syllab.SetInt64(buf, 48, p.WriteTime)
	copy(buf[56:], p.OwnerAppID[:])

	copy(buf[72:], p.AppInstanceID[:])
	copy(buf[88:], p.UserConnectionID[:])
	copy(buf[104:], p.ID[:])
	copy(buf[120:], p.OwnerID[:])
	copy(buf[136:], p.SellerID[:])
	copy(buf[152:], p.WikiID[:])
	copy(buf[168:], p.ProductionID[:])
	copy(buf[184:], p.DistributionCenterID[:])
	copy(buf[200:], p.ProductAuctionID[:])
	syllab.SetUInt8(buf, 216, uint8(p.Status))
	return
}

func (p *Product) syllabStackLen() (ln uint32) {
	return 217 // fixed size data + variables data add&&len
}

func (p *Product) syllabHeapLen() (ln uint32) {
	return
}

func (p *Product) syllabLen() (ln uint64) {
	return uint64(p.syllabStackLen() + p.syllabHeapLen())
}
