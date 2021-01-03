/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	"../libgo/achaemenid"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	lang "../libgo/language"
	"../libgo/log"
	"../libgo/pehrest"
	psdk "../libgo/pehrest-sdk"
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
		lang.LanguageEnglish: "Product",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "store product(goods) details!",
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
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	ID               [32]byte `index-hash:"RecordID"`  // ProductID
	OwnerID          [32]byte `index-hash:"ID[daily]"` // Who belong to! Just org can be first owner!
	QuiddityID       [32]byte `index-hash:"ID[daily],ID[pair,DCID,daily],ID[pair,DCID,temp],DCID[temp]"`

	SellerID         [32]byte `index-hash:"ID[daily]"` // OrdererID, who places the order usually use for prescription(drug order) or sales agent!
	ProductionID     [32]byte // It can also upper ID that this product split from it!
	DCID             [32]byte `index-hash:"QuiddityID"` // DistributionCenterID
	ProductAuctionID [32]byte `index-hash:"ID"`         // can be 0 for just change owner without any auction or price but very rare situation!
	Status           ProductStatus
}

// SaveNew method set some data and write entire Product record with all indexes!
func (p *Product) SaveNew() (err *er.Error) {
	err = p.Set()
	if err != nil {
		return
	}

	p.IndexRecordIDForID()
	p.IndexIDForOwnerIDDaily()
	p.IndexIDForQuiddityIDDaily()
	p.IndexIDForQuiddityIDDCIDDaily()
	if p.SellerID != [32]byte{} {
		p.IndexIDForSellerIDDaily()
	}
	if p.ProductAuctionID != [32]byte{} {
		p.IndexIDForProductAuctionID()
	}
	p.ListQuiddityIDForDCIDDaily()
	p.TempIndexIDForQuiddityIDDCID()
	p.TempIndexDCIDForQuiddityID()
	return
}

// Set method set some data and write entire Product record!
func (p *Product) Set() (err *er.Error) {
	p.RecordStructureID = productStructureID
	p.RecordSize = p.syllabLen()
	p.WriteTime = etime.Now()
	p.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: p.syllabEncoder(),
	}
	p.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], p.RecordID[:])

	err = gsdk.SetRecord(&req)
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
	res, err = gsdk.GetRecord(&req)
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
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: p.hashIDForRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
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

// IndexRecordIDForID save RecordID chain for ID
// Call in each update to the exiting record!
func (p *Product) IndexRecordIDForID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashIDForRecordID(),
		IndexValue: p.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashIDForRecordID() (hash [32]byte) {
	const field = "ID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.ID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexIDForOwnerIDDaily save ID chain for OwnerID daily
func (p *Product) IndexIDForOwnerIDDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashOwnerIDForIDDaily(),
		IndexValue: p.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashOwnerIDForIDDaily() (hash [32]byte) {
	const field = "OwnerID"
	var buf = make([]byte, 48+len(field)) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.OwnerID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	copy(buf[48:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForQuiddityIDDaily save ID chain for QuiddityID daily
func (p *Product) IndexIDForQuiddityIDDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashQuiddityIDForIDDaily(),
		IndexValue: p.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashQuiddityIDForIDDaily() (hash [32]byte) {
	const field = "QuiddityID"
	var buf = make([]byte, 48+len(field)) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.QuiddityID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	copy(buf[48:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForQuiddityIDDCIDDaily save ID chain for QuiddityID+DCID
func (p *Product) IndexIDForQuiddityIDDCIDDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashQuiddityIDDCIDForIDDaily(),
		IndexValue: p.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashQuiddityIDDCIDForIDDaily() (hash [32]byte) {
	const field = "QuiddityIDDCID"
	var buf = make([]byte, 80+len(field)) // 8+32+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.QuiddityID[:])
	copy(buf[40:], p.DCID[:])
	syllab.SetInt64(buf, 72, p.WriteTime.RoundToDay())
	copy(buf[80:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForSellerIDDaily save ID chain for SellerID daily
func (p *Product) IndexIDForSellerIDDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashSellerIDForIDDaily(),
		IndexValue: p.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashSellerIDForIDDaily() (hash [32]byte) {
	const field = "SellerID"
	var buf = make([]byte, 48+len(field)) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.SellerID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	copy(buf[48:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForProductAuctionID save ID chain for ProductAuctionID
// Use to indiacate product sell by specific auction. it is better to remove this record each month or year!
// Don't call in update to an exiting record!
func (p *Product) IndexIDForProductAuctionID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashAuctionIDForID(),
		IndexValue: p.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashAuctionIDForID() (hash [32]byte) {
	const field = "ProductAuctionID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.ProductAuctionID[:])
	copy(buf[80:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListQuiddityIDForDCIDDaily save QuiddityID chain for DCID daily
func (p *Product) ListQuiddityIDForDCIDDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashDCIDForQuiddityIDDaily(),
		IndexValue: p.QuiddityID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashDCIDForQuiddityIDDaily() (hash [32]byte) {
	const field = "ListDCID"
	var buf = make([]byte, 48+len(field)) // 8+32+8
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.DCID[:])
	syllab.SetInt64(buf, 40, p.WriteTime.RoundToDay())
	copy(buf[48:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Temporary INDEXES & LIST --
*/

// TempIndexIDForQuiddityIDDCID save temporary ID chain for QuiddityID+DCID
// Use to indiacate product stock in the DC. temp index to Pop from it!
func (p *Product) TempIndexIDForQuiddityIDDCID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashQuiddityIDDCIDForID(),
		IndexValue: p.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashQuiddityIDDCIDForID() (hash [32]byte) {
	const field = "TempQuiddityIDDCID"
	var buf = make([]byte, 72+len(field)) // 8+32+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.QuiddityID[:])
	copy(buf[40:], p.DCID[:])
	copy(buf[72:], field)
	return sha512.Sum512_256(buf)
}

// TempIndexDCIDForQuiddityID save temporary DCID chain for QuiddityID
// Use to indiacate global product stock. temp index to delete from it!
// Don't call in update to an exiting record!
func (p *Product) TempIndexDCIDForQuiddityID() {
	// TODO::: first check if given DCID exist in QuiddityID temp chain
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   p.hashQuiddityIDForDCID(),
		IndexValue: p.DCID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (p *Product) hashQuiddityIDForDCID() (hash [32]byte) {
	const field = "TempQuiddityID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productStructureID)
	copy(buf[8:], p.QuiddityID[:])
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
	copy(p.QuiddityID[:], buf[216:])

	copy(p.SellerID[:], buf[248:])
	copy(p.ProductionID[:], buf[280:])
	copy(p.DCID[:], buf[312:])
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
	copy(buf[216:], p.QuiddityID[:])

	copy(buf[248:], p.SellerID[:])
	copy(buf[280:], p.ProductionID[:])
	copy(buf[312:], p.DCID[:])
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
	ProductNotSet ProductStatus = iota
	ProductCreated
	ProductChangeDC
	ProductChangeQuiddity // Split to small size product! || Split from upper size product!
	ProductChangeOwner
	ProductVoid
	ProductPreSale // use in budget analysis and also can be trade!

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
