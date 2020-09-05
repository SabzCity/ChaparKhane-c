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
	productAuctionStructureID uint64 = 7190740114066546952
	productAuctionFixedSize   uint64 = 233 // 72 + 161 + (0 * 8) >> Common header + Unique data + vars add&&len
	productAuctionState       uint8  = ganjine.DataStructureStatePreAlpha
)

// ProductAuction store the product auction.
type ProductAuction struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID                [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID             [16]byte // Store to remember which user connection set||chanaged this record!
	ProductAuctionID             [16]byte
	OrgID                        [16]byte // Sell can be register just by producer organization
	WikiID                       [16]byte

	Currency                     uint16
	SuggestPrice                 uint64   // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
	DiscountPerMyriad            uint16   // ‱ PerMyriad Mines from SuggestPrice
	DistributionCenterCommission uint16   // ‱ PerMyriad decrease DiscountPerMyriad
	SellerCommission             uint16   // ‱ PerMyriad decrease DiscountPerMyriad
	PayablePrice                 uint64   // Some number base on currency is Decimal part e.g. 8099 >> 80.99$

	DistributionCenterID         [16]byte // if not 0 means this sale is just for specific DistributionCenter!
	MinNumBuy                    uint64   // Minimum number to buy in this auction use for sale-off,...
	StockNumber                  uint64   // 0 for unlimited until related product exist to sell!
	GroupID                      [16]byte // it can be 0 and means sale is global!
	LiveUntil                    int64
	Status                       uint8
}

// ProductSale status
const (
	// ProductCreated indicate
	ProductCreated uint8 = iota
	// ProductVoid
	ProductVoid
	// ProductPreSale use in budget analysis and also can be trade!
	ProductPreSale
	// 0x0 for non expire record, 0x1 for sell to first above SuggestPrice||buy first below it!
)

// ManagerApprove
// WarehouseApprove
// ActiveWithApprove
// ActiveWithoutApprove
// ActiveWithOrderer
// Inactive
// Normal
// Reject

// Set method set some data and write entire ProductAuction record!
func (pa *ProductAuction) Set() (err error) {
	pa.RecordStructureID = productAuctionStructureID
	pa.RecordSize = productAuctionFixedSize
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
func (pa *ProductAuction) GetByRecordID() (err error) {
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

	if pa.RecordStructureID != productAuctionStructureID {
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByAuctionID method find and read last version of record by given AuctionID
func (pa *ProductAuction) GetByAuctionID() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: pa.HashAuctionID(),
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
		pa.RecordID = indexRes.RecordIDs[ln]
		err = pa.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexAuctionID index pa.AuctionID to retrieve record fast later.
func (pa *ProductAuction) IndexAuctionID() {
	var auctionIDIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashAuctionID(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &auctionIDIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexWiki index pa.Wiki to retrieve record fast later.
func (pa *ProductAuction) IndexWiki() {
	var wikiIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashWiki(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &wikiIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexWikiDC index pa.WikiDC to retrieve record fast later.
func (pa *ProductAuction) IndexWikiDC() {
	var wikiDCIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashWikiDC(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &wikiDCIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// IndexWikiGroup index pa.WikiGroup to retrieve record fast later.
func (pa *ProductAuction) IndexWikiGroup() {
	var wikiGroupIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: pa.HashWikiGroup(),
		RecordID:  pa.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &wikiGroupIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashAuctionID hash productAuctionStructureID + pa.AuctionID
func (pa *ProductAuction) HashAuctionID() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	copy(buf[8:], pa.ProductAuctionID[:])

	return sha512.Sum512_256(buf)
}

// HashWiki hash productAuctionStructureID + pa.WikiID
func (pa *ProductAuction) HashWiki() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	copy(buf[8:], pa.WikiID[:])

	return sha512.Sum512_256(buf)
}

// HashWikiDC hash productAuctionStructureID + pa.WikiID + pa.DistributionCenterID
func (pa *ProductAuction) HashWikiDC() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+16+16

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	copy(buf[8:], pa.WikiID[:])
	copy(buf[24:], pa.DistributionCenterID[:])

	return sha512.Sum512_256(buf)
}

// HashWikiGroup hash productAuctionStructureID + pa.WikiID + pa.GroupID
func (pa *ProductAuction) HashWikiGroup() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+16+16

	buf[0] = byte(pa.RecordStructureID)
	buf[1] = byte(pa.RecordStructureID >> 8)
	buf[2] = byte(pa.RecordStructureID >> 16)
	buf[3] = byte(pa.RecordStructureID >> 24)
	buf[4] = byte(pa.RecordStructureID >> 32)
	buf[5] = byte(pa.RecordStructureID >> 40)
	buf[6] = byte(pa.RecordStructureID >> 48)
	buf[7] = byte(pa.RecordStructureID >> 56)

	copy(buf[8:], pa.WikiID[:])
	copy(buf[24:], pa.GroupID[:])

	return sha512.Sum512_256(buf)
}

func (pa *ProductAuction) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < productAuctionFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(pa.RecordID[:], buf[:])
	pa.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	pa.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	pa.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(pa.OwnerAppID[:], buf[56:])

	copy(pa.AppInstanceID[:], buf[72:])
	copy(pa.UserConnectionID[:], buf[88:])
	copy(pa.ProductAuctionID[:], buf[104:])
	copy(pa.OrgID[:], buf[120:])
	copy(pa.WikiID[:], buf[136:])
	pa.Currency = uint16(buf[152]) | uint16(buf[153])<<8
	pa.SuggestPrice = uint64(buf[154]) | uint64(buf[155])<<8 | uint64(buf[156])<<16 | uint64(buf[157])<<24 | uint64(buf[158])<<32 | uint64(buf[159])<<40 | uint64(buf[160])<<48 | uint64(buf[161])<<56
	pa.DiscountPerMyriad = uint16(buf[162]) | uint16(buf[163])<<8
	pa.DistributionCenterCommission = uint16(buf[164]) | uint16(buf[165])<<8
	pa.SellerCommission = uint16(buf[166]) | uint16(buf[167])<<8
	pa.PayablePrice = uint64(buf[168]) | uint64(buf[169])<<8 | uint64(buf[170])<<16 | uint64(buf[171])<<24 | uint64(buf[172])<<32 | uint64(buf[173])<<40 | uint64(buf[174])<<48 | uint64(buf[175])<<56
	copy(pa.DistributionCenterID[:], buf[176:])
	pa.MinNumBuy = uint64(buf[192]) | uint64(buf[193])<<8 | uint64(buf[194])<<16 | uint64(buf[195])<<24 | uint64(buf[196])<<32 | uint64(buf[197])<<40 | uint64(buf[198])<<48 | uint64(buf[199])<<56
	pa.StockNumber = uint64(buf[200]) | uint64(buf[201])<<8 | uint64(buf[202])<<16 | uint64(buf[203])<<24 | uint64(buf[204])<<32 | uint64(buf[205])<<40 | uint64(buf[206])<<48 | uint64(buf[207])<<56
	copy(pa.GroupID[:], buf[208:])
	pa.LiveUntil = int64(buf[224]) | int64(buf[225])<<8 | int64(buf[226])<<16 | int64(buf[227])<<24 | int64(buf[228])<<32 | int64(buf[229])<<40 | int64(buf[230])<<48 | int64(buf[231])<<56
	pa.Status = uint8(buf[232])

	return
}

func (pa *ProductAuction) syllabEncoder() (buf []byte) {
	buf = make([]byte, productAuctionFixedSize)

	// copy(buf[0:], pa.RecordID[:])
	buf[32] = byte(pa.RecordStructureID)
	buf[33] = byte(pa.RecordStructureID >> 8)
	buf[34] = byte(pa.RecordStructureID >> 16)
	buf[35] = byte(pa.RecordStructureID >> 24)
	buf[36] = byte(pa.RecordStructureID >> 32)
	buf[37] = byte(pa.RecordStructureID >> 40)
	buf[38] = byte(pa.RecordStructureID >> 48)
	buf[39] = byte(pa.RecordStructureID >> 56)
	buf[40] = byte(pa.RecordSize)
	buf[41] = byte(pa.RecordSize >> 8)
	buf[42] = byte(pa.RecordSize >> 16)
	buf[43] = byte(pa.RecordSize >> 24)
	buf[44] = byte(pa.RecordSize >> 32)
	buf[45] = byte(pa.RecordSize >> 40)
	buf[46] = byte(pa.RecordSize >> 48)
	buf[47] = byte(pa.RecordSize >> 56)
	buf[48] = byte(pa.WriteTime)
	buf[49] = byte(pa.WriteTime >> 8)
	buf[50] = byte(pa.WriteTime >> 16)
	buf[51] = byte(pa.WriteTime >> 24)
	buf[52] = byte(pa.WriteTime >> 32)
	buf[53] = byte(pa.WriteTime >> 40)
	buf[54] = byte(pa.WriteTime >> 48)
	buf[55] = byte(pa.WriteTime >> 56)
	copy(buf[56:], pa.OwnerAppID[:])

	copy(buf[72:], pa.AppInstanceID[:])
	copy(buf[88:], pa.UserConnectionID[:])
	copy(buf[104:], pa.ProductAuctionID[:])
	copy(buf[120:], pa.OrgID[:])
	copy(buf[136:], pa.WikiID[:])
	buf[152] = byte(pa.Currency)
	buf[153] = byte(pa.Currency >> 8)
	buf[154] = byte(pa.SuggestPrice)
	buf[155] = byte(pa.SuggestPrice >> 8)
	buf[156] = byte(pa.SuggestPrice >> 16)
	buf[157] = byte(pa.SuggestPrice >> 24)
	buf[158] = byte(pa.SuggestPrice >> 32)
	buf[159] = byte(pa.SuggestPrice >> 40)
	buf[160] = byte(pa.SuggestPrice >> 48)
	buf[161] = byte(pa.SuggestPrice >> 56)
	buf[162] = byte(pa.DiscountPerMyriad)
	buf[163] = byte(pa.DiscountPerMyriad >> 8)
	buf[164] = byte(pa.DistributionCenterCommission)
	buf[165] = byte(pa.DistributionCenterCommission >> 8)
	buf[166] = byte(pa.SellerCommission)
	buf[167] = byte(pa.SellerCommission >> 8)
	buf[168] = byte(pa.PayablePrice)
	buf[169] = byte(pa.PayablePrice >> 8)
	buf[170] = byte(pa.PayablePrice >> 16)
	buf[171] = byte(pa.PayablePrice >> 24)
	buf[172] = byte(pa.PayablePrice >> 32)
	buf[173] = byte(pa.PayablePrice >> 40)
	buf[174] = byte(pa.PayablePrice >> 48)
	buf[175] = byte(pa.PayablePrice >> 56)
	copy(buf[176:], pa.DistributionCenterID[:])
	buf[192] = byte(pa.MinNumBuy)
	buf[193] = byte(pa.MinNumBuy >> 8)
	buf[194] = byte(pa.MinNumBuy >> 16)
	buf[195] = byte(pa.MinNumBuy >> 24)
	buf[196] = byte(pa.MinNumBuy >> 32)
	buf[197] = byte(pa.MinNumBuy >> 40)
	buf[198] = byte(pa.MinNumBuy >> 48)
	buf[199] = byte(pa.MinNumBuy >> 56)
	buf[200] = byte(pa.StockNumber)
	buf[201] = byte(pa.StockNumber >> 8)
	buf[202] = byte(pa.StockNumber >> 16)
	buf[203] = byte(pa.StockNumber >> 24)
	buf[204] = byte(pa.StockNumber >> 32)
	buf[205] = byte(pa.StockNumber >> 40)
	buf[206] = byte(pa.StockNumber >> 48)
	buf[207] = byte(pa.StockNumber >> 56)
	copy(buf[208:], pa.GroupID[:])
	buf[224] = byte(pa.LiveUntil)
	buf[225] = byte(pa.LiveUntil >> 8)
	buf[226] = byte(pa.LiveUntil >> 16)
	buf[227] = byte(pa.LiveUntil >> 24)
	buf[228] = byte(pa.LiveUntil >> 32)
	buf[229] = byte(pa.LiveUntil >> 40)
	buf[230] = byte(pa.LiveUntil >> 48)
	buf[231] = byte(pa.LiveUntil >> 56)
	buf[232] = byte(pa.Status)

	return
}
