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
	"../libgo/math"
	"../libgo/pehrest"
	psdk "../libgo/pehrest-sdk"
	"../libgo/price"
	"../libgo/syllab"
)

const (
	productPriceStructureID uint64 = 7226862268582639719
)

var productPriceStructure = ganjine.DataStructure{
	ID:                7226862268582639719,
	IssueDate:         1607422485,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         ProductPrice{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Product Price",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"",
	},
}

// ProductPrice ---Read locale description in productPriceStructure---
type ProductPrice struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	OrgID            [32]byte `index-hash:"QuiddityID"` // Sell can be register just by producer organization
	QuiddityID       [32]byte `index-hash:"RecordID"`

	MaterialsPercent   math.PerMyriad
	MaterialsCost      price.Amount
	LaborPercent       math.PerMyriad
	LaborCost          price.Amount
	InvestmentsPercent math.PerMyriad
	InvestmentsCost    price.Amount
	TotalCost          price.Amount

	Markup          math.PerMyriad // the % added to the cost to determine the wholesale price
	WholesaleProfit price.Amount   // The amount of wholesale markup
	Margin          math.PerMyriad // the % added to the cost to determine the retail price
	RetailProfit    price.Amount   // The amount of retail markup

	TaxPercent math.PerMyriad // VAT, ...
	Tax        price.Amount

	Price price.Amount
}

// SaveNew method set some data and write entire ProductPrice record with all indexes!
func (pp *ProductPrice) SaveNew() (err *er.Error) {
	err = pp.Set()
	if err != nil {
		return
	}

	pp.IndexRecordIDForQuiddityID()
	pp.IndexQuiddityIDForOrgID()
	return
}

// Set method set some data and write entire ProductPrice record!
func (pp *ProductPrice) Set() (err *er.Error) {
	pp.RecordStructureID = productPriceStructureID
	pp.RecordSize = pp.syllabLen()
	pp.WriteTime = etime.Now()
	pp.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: pp.syllabEncoder(),
	}
	pp.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], pp.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Record Error:", err)
		}
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID read existing record data by given RecordID!
func (pp *ProductPrice) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          pp.RecordID,
		RecordStructureID: productPriceStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = pp.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if pp.RecordStructureID != productPriceStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", organizationAuthenticationStructureID)
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByQuiddityID method read last existing record data by given QuiddityID!
func (pp *ProductPrice) GetLastByQuiddityID() (err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: pp.hashQuiddityIDForRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	if err != nil {
		return
	}

	pp.RecordID = indexRes.IndexValues[0]
	err = pp.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", productAuctionStructureID)
	}
	return
}

/*
	-- Search Methods --
*/

// FindRecordsIDsByQuiddityID find RecordsIDs by given ID
func (pp *ProductPrice) FindRecordsIDsByQuiddityID(offset, limit uint64) (RecordsIDs [][32]byte, err *er.Error) {
	var indexRequest = &pehrest.HashGetValuesReq{
		IndexKey: pp.hashQuiddityIDForRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexRequest)
	RecordsIDs = indexRes.IndexValues
	return
}

// FindQuiddityIDsByOrgID find QuiddityIDs by given OrgID
func (pp *ProductPrice) FindQuiddityIDsByOrgID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexRequest = &pehrest.HashGetValuesReq{
		IndexKey: pp.hashOrgIDForQuiddityID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexRequest)
	IDs = indexRes.IndexValues
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexRecordIDForQuiddityID save RecordID chain for QuiddityID
// Call in each update to the exiting record!
func (pp *ProductPrice) IndexRecordIDForQuiddityID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pp.hashQuiddityIDForRecordID(),
		IndexValue: pp.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index Error:", err)
		}
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pp *ProductPrice) hashQuiddityIDForRecordID() (hash [32]byte) {
	const field = "QuiddityID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productPriceStructureID)
	copy(buf[8:], pp.QuiddityID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// IndexQuiddityIDForOrgID save QuiddityID chain for OrgID.
// Don't call in update to an exiting record!
func (pp *ProductPrice) IndexQuiddityIDForOrgID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   pp.hashOrgIDForQuiddityID(),
		IndexValue: pp.QuiddityID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		if log.DebugMode {
			log.Debug("Ganjine - Set Index Error:", err)
		}
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (pp *ProductPrice) hashOrgIDForQuiddityID() (hash [32]byte) {
	const field = "OrgID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, productPriceStructureID)
	copy(buf[8:], pp.OrgID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ??

/*
	-- Temporary INDEXES & LIST --
*/

// ??

/*
	-- Syllab Encoder & Decoder --
*/

func (pp *ProductPrice) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < pp.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(pp.RecordID[:], buf[0:])
	pp.RecordStructureID = syllab.GetUInt64(buf, 32)
	pp.RecordSize = syllab.GetUInt64(buf, 40)
	pp.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(pp.OwnerAppID[:], buf[56:])

	copy(pp.AppInstanceID[:], buf[88:])
	copy(pp.UserConnectionID[:], buf[120:])
	copy(pp.OrgID[:], buf[152:])
	copy(pp.QuiddityID[:], buf[184:])

	pp.MaterialsPercent = math.PerMyriad(syllab.GetUInt16(buf, 216))
	pp.MaterialsCost = price.Amount(syllab.GetInt64(buf, 218))
	pp.LaborPercent = math.PerMyriad(syllab.GetUInt16(buf, 226))
	pp.LaborCost = price.Amount(syllab.GetInt64(buf, 228))
	pp.InvestmentsPercent = math.PerMyriad(syllab.GetUInt16(buf, 236))
	pp.InvestmentsCost = price.Amount(syllab.GetInt64(buf, 238))
	pp.TotalCost = price.Amount(syllab.GetInt64(buf, 246))

	pp.Markup = math.PerMyriad(syllab.GetUInt16(buf, 254))
	pp.WholesaleProfit = price.Amount(syllab.GetInt64(buf, 256))
	pp.Margin = math.PerMyriad(syllab.GetUInt16(buf, 264))
	pp.RetailProfit = price.Amount(syllab.GetInt64(buf, 266))

	pp.TaxPercent = math.PerMyriad(syllab.GetUInt16(buf, 274))
	pp.Tax = price.Amount(syllab.GetInt64(buf, 276))

	pp.Price = price.Amount(syllab.GetInt64(buf, 284))
	return
}

func (pp *ProductPrice) syllabEncoder() (buf []byte) {
	copy(buf[0:], pp.RecordID[:])
	syllab.SetUInt64(buf, 32, pp.RecordStructureID)
	syllab.SetUInt64(buf, 40, pp.RecordSize)
	syllab.SetInt64(buf, 48, int64(pp.WriteTime))
	copy(buf[56:], pp.OwnerAppID[:])

	copy(buf[88:], pp.AppInstanceID[:])
	copy(buf[120:], pp.UserConnectionID[:])
	copy(buf[152:], pp.OrgID[:])
	copy(buf[184:], pp.QuiddityID[:])

	syllab.SetUInt16(buf, 216, uint16(pp.MaterialsPercent))
	syllab.SetInt64(buf, 218, int64(pp.MaterialsCost))
	syllab.SetUInt16(buf, 226, uint16(pp.LaborPercent))
	syllab.SetInt64(buf, 228, int64(pp.LaborCost))
	syllab.SetUInt16(buf, 236, uint16(pp.InvestmentsPercent))
	syllab.SetInt64(buf, 238, int64(pp.InvestmentsCost))
	syllab.SetInt64(buf, 246, int64(pp.TotalCost))

	syllab.SetUInt16(buf, 254, uint16(pp.Markup))
	syllab.SetInt64(buf, 256, int64(pp.WholesaleProfit))
	syllab.SetUInt16(buf, 264, uint16(pp.Margin))
	syllab.SetInt64(buf, 266, int64(pp.RetailProfit))

	syllab.SetUInt16(buf, 274, uint16(pp.TaxPercent))
	syllab.SetInt64(buf, 276, int64(pp.Tax))

	syllab.SetInt64(buf, 284, int64(pp.Price))
	return
}

func (pp *ProductPrice) syllabStackLen() (ln uint32) {
	return 292
}

func (pp *ProductPrice) syllabHeapLen() (ln uint32) {
	return
}

func (pp *ProductPrice) syllabLen() (ln uint64) {
	return uint64(pp.syllabStackLen() + pp.syllabHeapLen())
}
