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
	financialTransactionStructureID uint64 = 15981238345012607782
)

var financialTransactionStructure = ganjine.DataStructure{
	ID:                15981238345012607782,
	IssueDate:         1599291620,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         FinancialTransaction{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Financial Transaction",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: `store each financial transaction.`,
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
	WriteTime         int64
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID         [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID      [32]byte // Store to remember which user connection set||chanaged this record!
	UserID                [32]byte `ganjine:"Unique" hash-index:"RecordID[daily,pair,Currency]"`
	Currency              uint16   `ganjine:"Unique" hash-index:"UserID"`
	ReferenceID           [32]byte // ProductID || UserID Transferred || ForeignExchangeID || JusticeID || ...
	ReferenceType         uint8    // 0:Product 1:Transfer 2:Transfer-bank 3:Block 4:donate
	PreviousTransactionID [32]byte // Last RecordID of same user with same currency
	Amount                int64    // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
	Balance               uint64   // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
}

// Set method set some data and write entire FinancialTransaction record!
func (ft *FinancialTransaction) Set() (err *er.Error) {
	ft.RecordStructureID = financialTransactionStructureID
	ft.RecordSize = ft.syllabLen()
	ft.WriteTime = etime.Now()
	ft.OwnerAppID = server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: ft.syllabEncoder(),
	}
	ft.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], ft.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (ft *FinancialTransaction) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID: ft.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = ft.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if ft.RecordStructureID != financialTransactionStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastTransactionByUserCurrency method find and read last version of record by given UserID + Currency.
// It returns error if can't find any record in last 90 days! longer than this period must do carefully in service logic!
func (ft *FinancialTransaction) GetLastTransactionByUserCurrency() (err *er.Error) {
	var indexRequest = gs.HashIndexGetValuesReq{
		IndexKey: ft.hashUserIDCurrencyforRecordIDDaily(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	for i := 0; i < 91; i++ {
		indexRes, err = gsdk.HashIndexGetValues(cluster, &indexRequest)
		if err == nil {
			break
		}
		ft.WriteTime -= (24 * 60 * 60)
		indexRequest.IndexKey = ft.hashUserIDCurrencyforRecordIDDaily()
	}
	if err != nil {
		return
	}

	ft.RecordID = indexRes.IndexValues[0]
	err = ft.GetByRecordID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", financialTransactionStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexUserCurrencyTransactionDaily index ft.UserID + ft.Currency on daily base to retrieve record fast later.
func (ft *FinancialTransaction) IndexUserCurrencyTransactionDaily() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   ft.hashUserIDCurrencyforRecordIDDaily(),
		IndexValue: ft.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ft *FinancialTransaction) hashUserIDCurrencyforRecordIDDaily() (hash [32]byte) {
	var buf = make([]byte, 50) // 8+32+2+8
	syllab.SetUInt64(buf, 0, financialTransactionStructureID)
	copy(buf[8:], ft.UserID[:])
	syllab.SetUInt16(buf, 40, ft.Currency)
	syllab.SetInt64(buf, 42, etime.RoundToDay(ft.WriteTime))
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// ??

/*
	-- LIST FIELDS --
*/

// ListCurrencyforUserID list all Currency own by specific UserID.
// Just call in first create currency record not in each transaction!
func (ft *FinancialTransaction) ListCurrencyforUserID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:     gs.RequestTypeBroadcast,
		IndexKey: ft.hashUserIDforCurrency(),
	}
	syllab.SetUInt16(indexRequest.IndexValue[:], 0, ft.Currency)
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (ft *FinancialTransaction) hashUserIDforCurrency() (hash [32]byte) {
	const field = "ListCurrency"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, financialTransactionStructureID)
	copy(buf[8:], ft.UserID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

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
	ft.WriteTime = syllab.GetInt64(buf, 48)
	copy(ft.OwnerAppID[:], buf[56:])

	copy(ft.AppInstanceID[:], buf[88:])
	copy(ft.UserConnectionID[:], buf[120:])
	copy(ft.UserID[:], buf[152:])
	ft.Currency = syllab.GetUInt16(buf, 184)
	copy(ft.ReferenceID[:], buf[186:])
	ft.ReferenceType = syllab.GetUInt8(buf, 218)
	copy(ft.PreviousTransactionID[:], buf[219:])
	ft.Amount = syllab.GetInt64(buf, 251)
	ft.Balance = syllab.GetUInt64(buf, 259)
	return
}

func (ft *FinancialTransaction) syllabEncoder() (buf []byte) {
	buf = make([]byte, ft.syllabLen())

	// copy(buf[0:], ft.RecordID[:])
	syllab.SetUInt64(buf, 32, ft.RecordStructureID)
	syllab.SetUInt64(buf, 40, ft.RecordSize)
	syllab.SetInt64(buf, 48, ft.WriteTime)
	copy(buf[56:], ft.OwnerAppID[:])

	copy(buf[88:], ft.AppInstanceID[:])
	copy(buf[120:], ft.UserConnectionID[:])
	copy(buf[152:], ft.UserID[:])
	syllab.SetUInt16(buf, 184, ft.Currency)
	copy(buf[186:], ft.ReferenceID[:])
	syllab.SetUInt8(buf, 218, ft.ReferenceType)
	copy(buf[219:], ft.PreviousTransactionID[:])
	syllab.SetInt64(buf, 251, ft.Amount)
	syllab.SetUInt64(buf, 259, ft.Balance)
	return
}

func (ft *FinancialTransaction) syllabStackLen() (ln uint32) {
	return 267
}

func (ft *FinancialTransaction) syllabHeapLen() (ln uint32) {
	return
}

func (ft *FinancialTransaction) syllabLen() (ln uint64) {
	return uint64(ft.syllabStackLen() + ft.syllabHeapLen())
}
