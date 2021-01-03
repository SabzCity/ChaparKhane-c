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
	"../libgo/price"
	"../libgo/syllab"
)

const (
	financialTransactionStructureID uint64 = 11180411632961596298
)

var financialTransactionStructure = ganjine.DataStructure{
	ID:                11180411632961596298,
	IssueDate:         1599291620,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         FinancialTransaction{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Financial Transaction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: `store each financial transaction.`,
	},
	TAGS: []string{
		"",
	},
}

// FinancialTransaction ---Read locale description in financialTransactionStructure---
type FinancialTransaction struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID         [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID      [32]byte // Store to remember which user connection set||chanaged this record!
	UserID                [32]byte `index-hash:"RecordID[daily]"`
	ReferenceID           [32]byte // data base on ReferenceType
	ReferenceType         FinancialTransactionType
	PreviousTransactionID [32]byte     // Last RecordID this transaction base on it!
	Amount                price.Amount // Some number base on currency is Decimal part e.g. 8099 >> 80.99$
	Balance               price.Amount // Some number base on currency is Decimal part e.g. 8099 >> 80.99$
}

// Lock read last RecordID and lock userID to set in-time transactions!
func (ft *FinancialTransaction) Lock() (err *er.Error) {
	err = ft.GetLastTransactionByUserID()
	return
}

// UnLock will unlock userID to let other services set new in-time transaction records!
func (ft *FinancialTransaction) UnLock() (err *er.Error) {
	err = ft.SaveNew()
	if err != nil {
		return
	}
	return
}

// SaveNew method set some data and write entire FinancialTransaction record with all indexes!
func (ft *FinancialTransaction) SaveNew() (err *er.Error) {
	err = ft.Set()
	if err != nil {
		return
	}

	ft.IndexUserIDForRecordIDDaily()
	return
}

// Set method set some data and write entire FinancialTransaction record!
func (ft *FinancialTransaction) Set() (err *er.Error) {
	ft.RecordStructureID = financialTransactionStructureID
	ft.RecordSize = ft.syllabLen()
	ft.WriteTime = etime.Now()
	ft.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: ft.syllabEncoder(),
	}
	ft.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], ft.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (ft *FinancialTransaction) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          ft.RecordID,
		RecordStructureID: financialTransactionStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = ft.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if ft.RecordStructureID != financialTransactionStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

/*
	-- Get Last Methods --
*/

// GetLastTransactionByUserID method find and read last version of record by given UserID.
// It returns error if can't find any record in last 90 days! longer than this period must do carefully in service logic!
func (ft *FinancialTransaction) GetLastTransactionByUserID() (err *er.Error) {
	var indexRequest = pehrest.HashGetValuesReq{
		IndexKey: ft.hashUserIDForRecordIDDaily(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *pehrest.HashGetValuesRes
	for i := 0; i < 91; i++ {
		indexRes, err = psdk.HashGetValues(&indexRequest)
		if err == nil {
			break
		}
		ft.WriteTime -= (24 * 60 * 60)
		indexRequest.IndexKey = ft.hashUserIDForRecordIDDaily()
	}
	if err != nil {
		return
	}

	ft.RecordID = indexRes.IndexValues[0]
	err = ft.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", financialTransactionStructureID)
	}
	return
}

/*
	-- Search Methods --
*/

// FindRecordIDsByUserIDWriteTime find RecordsIDs by given UserID + WriteTime(round to daily)
func (ft *FinancialTransaction) FindRecordIDsByUserIDWriteTime(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: ft.hashUserIDForRecordIDDaily(),
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

// IndexUserIDForRecordIDDaily index ft.UserID on daily base to retrieve record fast later.
func (ft *FinancialTransaction) IndexUserIDForRecordIDDaily() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   ft.hashUserIDForRecordIDDaily(),
		IndexValue: ft.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ft *FinancialTransaction) hashUserIDForRecordIDDaily() (hash [32]byte) {
	const field = "UserID"
	var buf = make([]byte, 48+len(field)) // 8+32+8
	syllab.SetUInt64(buf, 0, financialTransactionStructureID)
	copy(buf[8:], ft.UserID[:])
	syllab.SetInt64(buf, 40, ft.WriteTime.RoundToDay())
	copy(buf[48:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// ??

/*
	-- LIST FIELDS --
*/

// ??

/*
	-- Syllab Encoder & Decoder --
*/

func (ft *FinancialTransaction) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < ft.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(ft.RecordID[:], buf[0:])
	ft.RecordStructureID = syllab.GetUInt64(buf, 32)
	ft.RecordSize = syllab.GetUInt64(buf, 40)
	ft.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(ft.OwnerAppID[:], buf[56:])

	copy(ft.AppInstanceID[:], buf[88:])
	copy(ft.UserConnectionID[:], buf[120:])
	copy(ft.UserID[:], buf[152:])
	copy(ft.ReferenceID[:], buf[184:])
	ft.ReferenceType = FinancialTransactionType(syllab.GetUInt8(buf, 216))
	copy(ft.PreviousTransactionID[:], buf[217:])
	ft.Amount = price.Amount(syllab.GetInt64(buf, 249))
	ft.Balance = price.Amount(syllab.GetUInt64(buf, 257))
	return
}

func (ft *FinancialTransaction) syllabEncoder() (buf []byte) {
	buf = make([]byte, ft.syllabLen())

	// copy(buf[0:], ft.RecordID[:])
	syllab.SetUInt64(buf, 32, ft.RecordStructureID)
	syllab.SetUInt64(buf, 40, ft.RecordSize)
	syllab.SetInt64(buf, 48, int64(ft.WriteTime))
	copy(buf[56:], ft.OwnerAppID[:])

	copy(buf[88:], ft.AppInstanceID[:])
	copy(buf[120:], ft.UserConnectionID[:])
	copy(buf[152:], ft.UserID[:])
	copy(buf[184:], ft.ReferenceID[:])
	syllab.SetUInt8(buf, 216, uint8(ft.ReferenceType))
	copy(buf[217:], ft.PreviousTransactionID[:])
	syllab.SetInt64(buf, 249, int64(ft.Amount))
	syllab.SetInt64(buf, 257, int64(ft.Balance))
	return
}

func (ft *FinancialTransaction) syllabStackLen() (ln uint32) {
	return 265
}

func (ft *FinancialTransaction) syllabHeapLen() (ln uint32) {
	return
}

func (ft *FinancialTransaction) syllabLen() (ln uint64) {
	return uint64(ft.syllabStackLen() + ft.syllabHeapLen())
}

/*
	-- Record types --
*/

// FinancialTransactionType indicate FinancialTransaction record type
type FinancialTransactionType uint8

// FinancialTransaction types
const (
	FinancialTransactionUnset FinancialTransactionType = iota
	FinancialTransactionFailed
	FinancialTransactionBlocked //  JusticeID
	FinancialTransactionDonate  // UserID Transferred
	FinancialTransactionBankTransfer
	FinancialTransactionPOSTransfer // ForeignExchangeID
	FinancialTransactionWebTransfer // ForeignExchangeID
	FinancialTransactionProductAuctionCommission
	FinancialTransactionProductAuctionPrice // ProductID
)
