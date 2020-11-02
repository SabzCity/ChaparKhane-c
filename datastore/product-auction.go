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
	productAuctionStructureID uint64 = 7190740114066546952
)

var productAuctionStructure = ganjine.DataStructure{
	ID:                7190740114066546952,
	IssueDate:         1599286551,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         ProductAuction{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "ProductAuction",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: " store the product auction",
	},
	TAGS: []string{
		"",
	},
}

// ProductAuction ---Read locale description in wikiStructure---
type ProductAuction struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	OrgID            [32]byte // Sell can be register just by producer organization
	ID               [32]byte `ganjine:"Immutable,Unique"`
	WikiID           [32]byte `ganjine:"Immutable"`

	Currency                     uint16 `ganjine:"Immutable"`
	SuggestPrice                 uint64 // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
	DiscountPerMyriad            uint16 // ‱ PerMyriad Mines from SuggestPrice
	DistributionCenterCommission uint16 // ‱ PerMyriad decrease DiscountPerMyriad
	SellerCommission             uint16 // ‱ PerMyriad decrease DiscountPerMyriad
	PayablePrice                 uint64 // Some number base on currency is Decimal part e.g. 8099 >> 80.99$

	DistributionCenterID [32]byte `ganjine:"Immutable"` // if not 0 means this sale is just for specific DistributionCenter!
	MinNumBuy            uint64   // Minimum number to buy in this auction use for sale-off,...
	StockNumber          uint64   // 0 for unlimited until related product exist to sell!
	GroupID              [32]byte `ganjine:"Immutable"` // it can be 0 and means sale is global!
	LiveUntil            int64
	Type                 uint8 // https://en.wikipedia.org/wiki/Auction_theory
	Status               ProductAuctionStatus
}

// ProductAuctionStatus indicate ProductAuction record status
type ProductAuctionStatus uint8

// ProductAuction status
const (
	ProductAuctionCreated ProductAuctionStatus = iota
	ProductAuctionVoid
	ProductAuctionPreSale // use in budget analysis and also can be trade!
	// 0x0 for non expire record, 0x1 for sell to first above SuggestPrice||buy first below it!

// ManagerApprove
// WarehouseApprove
// ActiveWithApprove
// ActiveWithoutApprove
// ActiveWithOrderer
// Inactive
// Normal
// Reject
)

// Set method set some data and write entire ProductAuction record!
func (pa *ProductAuction) Set() (err *er.Error) {
	pa.RecordStructureID = productAuctionStructureID
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
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (pa *ProductAuction) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID: pa.RecordID,
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

	if pa.RecordStructureID != productAuctionStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLatByID method find and read last version of record by given pa.ID
func (pa *ProductAuction) GetLatByID() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashIDforRecordID(),
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
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", productAuctionStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexID index pa.ID to retrieve record fast later.
func (pa *ProductAuction) IndexID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashIDforRecordID(),
		IndexValue: pa.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashIDforRecordID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.ID[:])
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexWiki index pa.Wiki to retrieve record fast later.
func (pa *ProductAuction) IndexWiki() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashWikiIDforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashWikiIDforID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.WikiID[:])
	return sha512.Sum512_256(buf)
}

// IndexWikiDC index pa.WikiDC to retrieve record fast later.
func (pa *ProductAuction) IndexWikiDC() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashWikiIDDistributionCenterIDforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashWikiIDDistributionCenterIDforID() (hash [32]byte) {
	var buf = make([]byte, 72) // 8+32+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.WikiID[:])
	copy(buf[40:], pa.DistributionCenterID[:])
	return sha512.Sum512_256(buf)
}

// IndexWikiGroup index pa.WikiGroup to retrieve record fast later.
func (pa *ProductAuction) IndexWikiGroup() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashWikiIDGroupIDforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashWikiIDGroupIDforID() (hash [32]byte) {
	var buf = make([]byte, 72) // 8+32+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.WikiID[:])
	copy(buf[40:], pa.GroupID[:])
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
	pa.WriteTime = syllab.GetInt64(buf, 48)
	copy(pa.OwnerAppID[:], buf[56:])

	copy(pa.AppInstanceID[:], buf[88:])
	copy(pa.UserConnectionID[:], buf[120:])
	copy(pa.OrgID[:], buf[152:])
	copy(pa.ID[:], buf[184:])
	copy(pa.WikiID[:], buf[216:])

	pa.Currency = syllab.GetUInt16(buf, 248)
	pa.SuggestPrice = syllab.GetUInt64(buf, 250)
	pa.DiscountPerMyriad = syllab.GetUInt16(buf, 258)
	pa.DistributionCenterCommission = syllab.GetUInt16(buf, 260)
	pa.SellerCommission = syllab.GetUInt16(buf, 262)
	pa.PayablePrice = syllab.GetUInt64(buf, 264)

	copy(pa.DistributionCenterID[:], buf[272:])
	pa.MinNumBuy = syllab.GetUInt64(buf, 304)
	pa.StockNumber = syllab.GetUInt64(buf, 312)
	copy(pa.GroupID[:], buf[320:])
	pa.LiveUntil = syllab.GetInt64(buf, 352)
	pa.Type = syllab.GetUInt8(buf, 360)
	pa.Status = ProductAuctionStatus(syllab.GetUInt8(buf, 361))
	return
}

func (pa *ProductAuction) syllabEncoder() (buf []byte) {
	buf = make([]byte, pa.syllabLen())

	// copy(buf[0:], pa.RecordID[:])
	syllab.SetUInt64(buf, 32, pa.RecordStructureID)
	syllab.SetUInt64(buf, 40, pa.RecordSize)
	syllab.SetInt64(buf, 48, pa.WriteTime)
	copy(buf[56:], pa.OwnerAppID[:])

	copy(buf[88:], pa.AppInstanceID[:])
	copy(buf[120:], pa.UserConnectionID[:])
	copy(buf[152:], pa.OrgID[:])
	copy(buf[184:], pa.ID[:])
	copy(buf[216:], pa.WikiID[:])

	syllab.SetUInt16(buf, 248, pa.Currency)
	syllab.SetUInt64(buf, 250, pa.SuggestPrice)
	syllab.SetUInt16(buf, 258, pa.DiscountPerMyriad)
	syllab.SetUInt16(buf, 260, pa.DistributionCenterCommission)
	syllab.SetUInt16(buf, 262, pa.SellerCommission)
	syllab.SetUInt64(buf, 264, pa.PayablePrice)

	copy(buf[272:], pa.DistributionCenterID[:])
	syllab.SetUInt64(buf, 304, pa.MinNumBuy)
	syllab.SetUInt64(buf, 312, pa.StockNumber)
	copy(buf[320:], pa.GroupID[:])
	syllab.SetInt64(buf, 352, pa.LiveUntil)
	syllab.SetUInt8(buf, 360, pa.Type)
	syllab.SetUInt8(buf, 361, uint8(pa.Status))
	return
}

func (pa *ProductAuction) syllabStackLen() (ln uint32) {
	return 362
}

func (pa *ProductAuction) syllabHeapLen() (ln uint32) {
	return
}

func (pa *ProductAuction) syllabLen() (ln uint64) {
	return uint64(pa.syllabStackLen() + pa.syllabHeapLen())
}
