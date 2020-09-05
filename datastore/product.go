/* For license and copyright information please see LEGAL file in repository */

package datastore

import (
	"crypto/sha512"

	etime "../libgo/earth-time"
	"../libgo/ganjine"
	gsdk "../libgo/ganjine-sdk"
	gs "../libgo/ganjine-services"
	"../libgo/syllab"
)

const (
	productStructureID uint64 = 2911412040892278928
	productFixedSize   uint64 = 217 // 72 + 145 + (0 * 8) >> Common header + Unique data + vars add&&len
	productState       uint8  = ganjine.DataStructureStatePreAlpha
)

// Product store product(goods) details!
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
	ProductID            [16]byte // Immutable, can be same on multi record refer to same product with diffrent owner or status!
	OwnerID              [16]byte // Who belong to! Just org can be first owner!
	SellerID             [16]byte // OrdererID, who places the order usually use for prescription(drug order) or sales agent!
	WikiID               [16]byte
	ProductionID         [16]byte // It can also upper ProductID that this product split from it!
	DistributionCenterID [16]byte // It will changed only on owner changed otherwise can be mobile location! if != Sale.WarehouseID means user wants to send item to this address!
	ProductAuctionID     [16]byte // can be 0 for just change owner without any auction or price but very rare situation!
	Status               uint8
}

// Set method set some data and write entire Product record!
func (p *Product) Set() (err error) {
	p.RecordStructureID = productStructureID
	p.RecordSize = productFixedSize
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
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByProductID method find and read last version of record by given ProductID
func (p *Product) GetByProductID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: p.HashProductID(),
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
	for {
		ln--
		p.RecordID = indexRes.RecordIDs[ln]
		err = p.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexProductID index p.ProductID to retrieve record fast later.
func (p *Product) IndexProductID() {
	var productIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashProductID(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &productIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexOwnerDaily index p.Owner to retrieve record fast later.
func (p *Product) IndexOwnerDaily() {
	var ownerDailyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashOwnerDaily(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &ownerDailyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexSellerDaily index p.SellerDaily to retrieve record fast later.
func (p *Product) IndexSellerDaily() {
	var sellerDailyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashSellerDaily(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &sellerDailyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexWiki index p.Wiki to retrieve record fast later.
// Use to indiacate global product stock. temp index to delete from it!
// func (p *Product) IndexWiki() {
// 	var wikiIndex = gs.SetIndexHashReq{
// 		Type:      gs.RequestTypeBroadcast,
// 		IndexHash: p.HashWiki(),
// 		RecordID:  p.DistributionCenterID,
// 	}
// 	var err = gsdk.SetIndexHash(cluster, &wikiIndex)
// 	if err != nil {
// 		// TODO::: we must retry more due to record wrote successfully!
// 	}
// }

// IndexWikiDaily index p.WikiDaily to retrieve record fast later.
func (p *Product) IndexWikiDaily() {
	var wikiDailyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiDaily(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &wikiDailyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexWikiDCDaily index p.WikiDCDaily to retrieve record fast later.
// Each year is 365 days that indicate we have 365 index record each year per product wiki on each distribution center!
func (p *Product) IndexWikiDCDaily() {
	var wikiDCDailyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiDCDaily(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &wikiDCDailyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexWikiDC index p.WikiDC to retrieve record fast later.
// Use to indiacate product stock in the DC. temp index to Pop from it!
func (p *Product) IndexWikiDC() {
	var wikiDCIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashWikiDC(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &wikiDCIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexAuctionID index p.AuctionID to retrieve record fast later.
// Use to indiacate product sell by specific auction
func (p *Product) IndexAuctionID() {
	var auctionIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: p.HashAuctionID(),
		RecordID:  p.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &auctionIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashProductID hash productStructureID + p.ProductID
func (p *Product) HashProductID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	copy(buf[8:], p.ProductID[:])

	return sha512.Sum512_256(buf)
}

// HashOwnerDaily hash productStructureID + p.OwnerID + p.WriteTime(round to day)
func (p *Product) HashOwnerDaily() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+8+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	var roundedTime = etime.RoundToDay(p.WriteTime)
	buf[8] = byte(roundedTime)
	buf[9] = byte(roundedTime >> 8)
	buf[10] = byte(roundedTime >> 16)
	buf[11] = byte(roundedTime >> 24)
	buf[12] = byte(roundedTime >> 32)
	buf[13] = byte(roundedTime >> 40)
	buf[14] = byte(roundedTime >> 48)
	buf[15] = byte(roundedTime >> 56)

	copy(buf[16:], p.OwnerID[:])

	return sha512.Sum512_256(buf)
}

// HashSellerDaily hash productStructureID + p.SellerID + p.WriteTime(round to day)
func (p *Product) HashSellerDaily() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+8+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	var roundedTime = etime.RoundToDay(p.WriteTime)
	buf[8] = byte(roundedTime)
	buf[9] = byte(roundedTime >> 8)
	buf[10] = byte(roundedTime >> 16)
	buf[11] = byte(roundedTime >> 24)
	buf[12] = byte(roundedTime >> 32)
	buf[13] = byte(roundedTime >> 40)
	buf[14] = byte(roundedTime >> 48)
	buf[15] = byte(roundedTime >> 56)

	copy(buf[16:], p.SellerID[:])

	return sha512.Sum512_256(buf)
}

// HashWiki hash productStructureID + p.WikiID
func (p *Product) HashWiki() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	copy(buf[8:], p.WikiID[:])

	return sha512.Sum512_256(buf)
}

// HashWikiDaily hash productStructureID + p.WikiID + p.WriteTime(round to day)
func (p *Product) HashWikiDaily() (hash [32]byte) {
	var buf = make([]byte, 32) // 8+8+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	var roundedTime = etime.RoundToDay(p.WriteTime)
	buf[8] = byte(roundedTime)
	buf[9] = byte(roundedTime >> 8)
	buf[10] = byte(roundedTime >> 16)
	buf[11] = byte(roundedTime >> 24)
	buf[12] = byte(roundedTime >> 32)
	buf[13] = byte(roundedTime >> 40)
	buf[14] = byte(roundedTime >> 48)
	buf[15] = byte(roundedTime >> 56)

	copy(buf[16:], p.WikiID[:])

	return sha512.Sum512_256(buf)
}

// HashWikiDCDaily hash pa.StructureID + p.WriteTime(round to day) + p.WikiID + p.DistributionCenterID
func (p *Product) HashWikiDCDaily() (hash [32]byte) {
	var buf = make([]byte, 48) // 8+8+16+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	var roundedTime = etime.RoundToDay(p.WriteTime)
	buf[8] = byte(roundedTime)
	buf[9] = byte(roundedTime >> 8)
	buf[10] = byte(roundedTime >> 16)
	buf[11] = byte(roundedTime >> 24)
	buf[12] = byte(roundedTime >> 32)
	buf[13] = byte(roundedTime >> 40)
	buf[14] = byte(roundedTime >> 48)
	buf[15] = byte(roundedTime >> 56)

	copy(buf[16:], p.WikiID[:])
	copy(buf[32:], p.DistributionCenterID[:])

	return sha512.Sum512_256(buf)
}

// HashWikiDC hash productStructureID + p.WikiID + p.DistributionCenterID
func (p *Product) HashWikiDC() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+16+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	copy(buf[8:], p.WikiID[:])
	copy(buf[24:], p.DistributionCenterID[:])

	return sha512.Sum512_256(buf)
}

// HashAuctionID hash productStructureID + p.ProductAuctionID
func (p *Product) HashAuctionID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(p.RecordStructureID)
	buf[1] = byte(p.RecordStructureID >> 8)
	buf[2] = byte(p.RecordStructureID >> 16)
	buf[3] = byte(p.RecordStructureID >> 24)
	buf[4] = byte(p.RecordStructureID >> 32)
	buf[5] = byte(p.RecordStructureID >> 40)
	buf[6] = byte(p.RecordStructureID >> 48)
	buf[7] = byte(p.RecordStructureID >> 56)

	copy(buf[8:], p.ProductAuctionID[:])

	return sha512.Sum512_256(buf)
}

func (p *Product) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < productFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(p.RecordID[:], buf[:])
	p.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	p.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	p.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(p.OwnerAppID[:], buf[56:])

	copy(p.AppInstanceID[:], buf[72:])
	copy(p.UserConnectionID[:], buf[88:])
	copy(p.ProductID[:], buf[104:])
	copy(p.OwnerID[:], buf[120:])
	copy(p.SellerID[:], buf[136:])
	copy(p.WikiID[:], buf[152:])
	copy(p.ProductionID[:], buf[168:])
	copy(p.DistributionCenterID[:], buf[184:])
	copy(p.ProductAuctionID[:], buf[200:])
	p.Status = buf[216]

	return
}

func (p *Product) syllabEncoder() (buf []byte) {
	buf = make([]byte, productFixedSize)

	// copy(buf[0:], p.RecordID[:])
	buf[32] = byte(p.RecordStructureID)
	buf[33] = byte(p.RecordStructureID >> 8)
	buf[34] = byte(p.RecordStructureID >> 16)
	buf[35] = byte(p.RecordStructureID >> 24)
	buf[36] = byte(p.RecordStructureID >> 32)
	buf[37] = byte(p.RecordStructureID >> 40)
	buf[38] = byte(p.RecordStructureID >> 48)
	buf[39] = byte(p.RecordStructureID >> 56)
	buf[40] = byte(p.RecordSize)
	buf[41] = byte(p.RecordSize >> 8)
	buf[42] = byte(p.RecordSize >> 16)
	buf[43] = byte(p.RecordSize >> 24)
	buf[44] = byte(p.RecordSize >> 32)
	buf[45] = byte(p.RecordSize >> 40)
	buf[46] = byte(p.RecordSize >> 48)
	buf[47] = byte(p.RecordSize >> 56)
	buf[48] = byte(p.WriteTime)
	buf[49] = byte(p.WriteTime >> 8)
	buf[50] = byte(p.WriteTime >> 16)
	buf[51] = byte(p.WriteTime >> 24)
	buf[52] = byte(p.WriteTime >> 32)
	buf[53] = byte(p.WriteTime >> 40)
	buf[54] = byte(p.WriteTime >> 48)
	buf[55] = byte(p.WriteTime >> 56)
	copy(buf[56:], p.OwnerAppID[:])

	copy(buf[72:], p.AppInstanceID[:])
	copy(buf[88:], p.UserConnectionID[:])
	copy(buf[104:], p.ProductID[:])
	copy(buf[120:], p.OwnerID[:])
	copy(buf[136:], p.SellerID[:])
	copy(buf[152:], p.WikiID[:])
	copy(buf[168:], p.ProductionID[:])
	copy(buf[184:], p.DistributionCenterID[:])
	copy(buf[200:], p.ProductAuctionID[:])
	buf[215] = p.Status

	return
}
