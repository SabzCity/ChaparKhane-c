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
	"../libgo/math"
	"../libgo/price"
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
		lang.EnglishLanguage: " store the product auction data",
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
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	OrgID            [32]byte `hash-index:"ID"` // Sell can be register just by producer organization
	ID               [32]byte `ganjine:"Unique" hash-index:"RecordID"`
	WikiID           [32]byte `hash-index:"ID[pair,Currency],ID[pair,DistributionCenterID],ID[pair,GroupID]"`

	// Price
	Currency                     price.Currency
	SuggestPrice                 price.Amount
	DistributionCenterCommission math.PerMyriad // plus to SuggestPrice || minus from Discount
	SellerCommission             math.PerMyriad // plus to SuggestPrice || minus from Discount
	Discount                     math.PerMyriad // minus from SuggestPrice
	PayablePrice                 price.Amount   // Some number base on currency is Decimal part e.g. 8099 >> 80.99$

	// Authorization
	DistributionCenterID [32]byte `hash-index:"ID"` // Same as OrgID and if not 0 means this sale is just for specific DistributionCenter!
	GroupID              [32]byte `hash-index:"ID"` // it can be 0 and means sale is global!
	MinNumBuy            uint64   // Minimum number to buy in this auction to use for sale-off(Discount)
	StockNumber          uint64   // 0 for unlimited until related product exist to sell!
	LiveUntil            etime.Time
	AllowWeekdays        etime.Weekdays
	AllowDayhours        etime.Dayhours

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

	pa.HashIndexRecordIDForID()
	pa.HashIndexIDForOrgID()
	pa.HashIndexIDForWikiIDCurrency()
	if pa.DistributionCenterID != [32]byte{} {
		pa.HashIndexIDForWikiIDDistributionCenterID()
		pa.HashIndexIDForDistributionCenterID()
	}
	if pa.GroupID != [32]byte{} {
		pa.HashIndexIDForWikiIDGroupID()
		pa.HashIndexIDForGroupID()
	}
	return
}

// CalculatePayablePrice method set PayablePrice by given price data
func (pa *ProductAuction) CalculatePayablePrice() {
	pa.PayablePrice = pa.SuggestPrice - pa.SuggestPrice.PerMyriad(pa.Discount)
}

// Set method set some data and write entire ProductAuction record!
func (pa *ProductAuction) Set() (err *er.Error) {
	pa.RecordStructureID = productAuctionStructureID
	pa.RecordSize = pa.syllabLen()
	pa.WriteTime = etime.Now()
	pa.OwnerAppID = server.AppID

	pa.CalculatePayablePrice()

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
		RecordID:          pa.RecordID,
		RecordStructureID: productAuctionStructureID,
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
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByIDByHashIndex method find and read last version of record by given pa.ID
func (pa *ProductAuction) GetLastByIDByHashIndex() (err *er.Error) {
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
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", productAuctionStructureID)
	}
	return
}

/*
	-- Search Methods --
*/

// FindRecordsIDsByIDByHashIndex find RecordsIDs by given ID
func (pa *ProductAuction) FindRecordsIDsByIDByHashIndex(offset, limit uint64) (RecordsIDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashIDforRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	RecordsIDs = indexRes.IndexValues
	return
}

// FindIDsByOrgIDByHashIndex find IDs by given OrgID
func (pa *ProductAuction) FindIDsByOrgIDByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashOrgIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByWikiIDCurrencyByHashIndex find IDs by given WikiID+Currency
func (pa *ProductAuction) FindIDsByWikiIDCurrencyByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashWikiIDCurrencyforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByWikiIDDistributionCenterIDByHashIndex find IDs by given WikiID+DistributionCenterID
func (pa *ProductAuction) FindIDsByWikiIDDistributionCenterIDByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashWikiIDDistributionCenterIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByWikiIDGroupIDByHashIndex find IDs by given WikiID+GroupID
func (pa *ProductAuction) FindIDsByWikiIDGroupIDByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashWikiIDGroupIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByDistributionCenterIDByHashIndex find IDs by given DistributionCenterID
func (pa *ProductAuction) FindIDsByDistributionCenterIDByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashDistributionCenterIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByGroupIDByHashIndex find IDs by given GroupID
func (pa *ProductAuction) FindIDsByGroupIDByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: pa.hashGroupIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

/*
	-- PRIMARY INDEXES --
*/

// HashIndexRecordIDForID save RecordID chain for ID
func (pa *ProductAuction) HashIndexRecordIDForID() {
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

// HashIndexIDForOrgID save ID chain for OrgID
func (pa *ProductAuction) HashIndexIDForOrgID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashOrgIDforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashOrgIDforID() (hash [32]byte) {
	const field = "OrgID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.OrgID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForWikiIDCurrency save ID chain for WikiID+Currency
func (pa *ProductAuction) HashIndexIDForWikiIDCurrency() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashWikiIDCurrencyforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashWikiIDCurrencyforID() (hash [32]byte) {
	const field = "WikiIDCurrency"
	var buf = make([]byte, 44+len(field)) // 8+32+4
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.WikiID[:])
	syllab.SetUInt16(buf, 40, uint16(pa.Currency))
	copy(buf[44:], field)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForWikiIDDistributionCenterID save ID chain for WikiID+DistributionCenterID
func (pa *ProductAuction) HashIndexIDForWikiIDDistributionCenterID() {
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
	const field = "WikiIDDistributionCenterID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.WikiID[:])
	copy(buf[40:], pa.DistributionCenterID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForWikiIDGroupID save ID chain for WikiID+GroupID
func (pa *ProductAuction) HashIndexIDForWikiIDGroupID() {
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
	const field = "WikiIDGroupID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.WikiID[:])
	copy(buf[40:], pa.GroupID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForDistributionCenterID save ID chain for DistributionCenterID
func (pa *ProductAuction) HashIndexIDForDistributionCenterID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashDistributionCenterIDforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashDistributionCenterIDforID() (hash [32]byte) {
	const field = "DistributionCenterID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.DistributionCenterID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForGroupID save ID chain for GroupID
func (pa *ProductAuction) HashIndexIDForGroupID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pa.hashGroupIDforID(),
		IndexValue: pa.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pa *ProductAuction) hashGroupIDforID() (hash [32]byte) {
	const field = "GroupID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productAuctionStructureID)
	copy(buf[8:], pa.GroupID[:])
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
	copy(pa.WikiID[:], buf[216:])

	pa.Currency = price.Currency(syllab.GetUInt16(buf, 248))
	pa.SuggestPrice = price.Amount(syllab.GetUInt64(buf, 250))
	pa.DistributionCenterCommission = math.PerMyriad(syllab.GetUInt16(buf, 258))
	pa.SellerCommission = math.PerMyriad(syllab.GetUInt16(buf, 260))
	pa.Discount = math.PerMyriad(syllab.GetUInt16(buf, 262))
	pa.PayablePrice = price.Amount(syllab.GetUInt64(buf, 264))

	copy(pa.DistributionCenterID[:], buf[272:])
	copy(pa.GroupID[:], buf[304:])
	pa.MinNumBuy = syllab.GetUInt64(buf, 336)
	pa.StockNumber = syllab.GetUInt64(buf, 344)
	pa.LiveUntil = etime.Time(syllab.GetInt64(buf, 352))
	pa.AllowWeekdays = etime.Weekdays(syllab.GetUInt8(buf, 360))
	pa.AllowDayhours = etime.Dayhours(syllab.GetUInt32(buf, 361))

	pa.Description = syllab.UnsafeGetString(buf, 365)
	pa.Type = ProductAuctionType(syllab.GetUInt8(buf, 373))
	pa.Status = ProductAuctionStatus(syllab.GetUInt8(buf, 374))
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
	copy(buf[216:], pa.WikiID[:])

	syllab.SetUInt16(buf, 248, uint16(pa.Currency))
	syllab.SetUInt64(buf, 250, uint64(pa.SuggestPrice))
	syllab.SetUInt16(buf, 258, uint16(pa.DistributionCenterCommission))
	syllab.SetUInt16(buf, 260, uint16(pa.SellerCommission))
	syllab.SetUInt16(buf, 262, uint16(pa.Discount))
	syllab.SetUInt64(buf, 264, uint64(pa.PayablePrice))

	copy(buf[272:], pa.DistributionCenterID[:])
	copy(buf[304:], pa.GroupID[:])
	syllab.SetUInt64(buf, 336, pa.MinNumBuy)
	syllab.SetUInt64(buf, 344, pa.StockNumber)
	syllab.SetInt64(buf, 352, int64(pa.LiveUntil))
	syllab.SetUInt8(buf, 360, uint8(pa.AllowWeekdays))
	syllab.SetUInt32(buf, 361, uint32(pa.AllowDayhours))

	syllab.SetString(buf, pa.Description, 365, hsi)
	syllab.SetUInt8(buf, 373, uint8(pa.Type))
	syllab.SetUInt8(buf, 374, uint8(pa.Status))
	return
}

func (pa *ProductAuction) syllabStackLen() (ln uint32) {
	return 375
}

func (pa *ProductAuction) syllabHeapLen() (ln uint32) {
	ln = uint32(len(pa.Description))
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
