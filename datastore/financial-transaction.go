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
		lang.EnglishLanguage: "FinancialTransaction",
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
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID         [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID      [16]byte // Store to remember which user connection set||chanaged this record!
	OwnerID               [16]byte `ganjine:"Immutable"`
	Currency              uint16   `ganjine:"Immutable,Unique"`
	ReferenceID           [16]byte // ProductID || OwnerID Transferred || ForeignExchangeID || JusticeID || ...
	ReferenceType         uint8    // 0:Product 1:Transfer 2:Transfer-bank 3:Block 4:donate
	PreviousTransactionID [32]byte // Last RecordID of same user with same currency
	Amount                int64    // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
	Balance               uint64   // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
}

// Set method set some data and write entire FinancialTransaction record!
func (ft *FinancialTransaction) Set() (err error) {
	ft.RecordStructureID = financialTransactionStructureID
	ft.RecordSize = ft.syllabLen()
	ft.WriteTime = etime.Now()
	ft.OwnerAppID = server.Manifest.AppID

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
func (ft *FinancialTransaction) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: ft.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = ft.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if ft.RecordStructureID != financialTransactionStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastTransactionByOwnerCurrency method find and read last version of record by given OwnerID + Currency
func (ft *FinancialTransaction) GetLastTransactionByOwnerCurrency() (err error) {
	// TODO::: set this month in ft!!
	// ft.WriteTime =
	var indexReq = &gs.FindRecordsReq{
		IndexHash: ft.HashOwnerCurrencyTransactionMonthly(),
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
		ft.RecordID = indexRes.RecordIDs[ln]
		err = ft.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexOwnerCurrencyTransactionMonthly index ft.OwnerID + ft.Currency on monthly base to retrieve record fast later.
func (ft *FinancialTransaction) IndexOwnerCurrencyTransactionMonthly() {
	var userCurrencyMonthlyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ft.HashOwnerCurrencyTransactionMonthly(),
		RecordID:  ft.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &userCurrencyMonthlyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerCurrencyTransactionMonthly hash financialTransactionStructureID + ft.OwnerID + ft.Currency
func (ft *FinancialTransaction) HashOwnerCurrencyTransactionMonthly() (hash [32]byte) {
	var buf = make([]byte, 34) // 8+8+16+2
	syllab.SetUInt64(buf, 0, financialTransactionStructureID)
	syllab.SetInt64(buf, 8, etime.RoundToMonth(ft.WriteTime))
	copy(buf[16:], ft.OwnerID[:])
	syllab.SetUInt16(buf, 32, ft.Currency)
	return sha512.Sum512_256(buf)
}

/*
	-- SECONDARY INDEXES --
*/

// ??

/*
	-- LIST INDEXES --
*/

// IndexOwnerCurrencies index ft.OwnerID + ft.Currency.
// Just call in first create currency record not in each transaction!
func (ft *FinancialTransaction) IndexOwnerCurrencies() {
	var userCurrenciesIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ft.HashOwnerCurrencyField(),
	}
	syllab.SetUInt16(userCurrenciesIndex.RecordID[:], 0, ft.Currency)
	var err = gsdk.SetIndexHash(cluster, &userCurrenciesIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwnerCurrencyField hash financialTransactionStructureID + OwnerID + Currency field
func (ft *FinancialTransaction) HashOwnerCurrencyField() (hash [32]byte) {
	const field = "Currency"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, financialTransactionStructureID)
	copy(buf[8:], ft.OwnerID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (ft *FinancialTransaction) syllabDecoder(buf []byte) (err error) {
	if uint32(len(buf)) < ft.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(ft.RecordID[:], buf[0:])
	ft.RecordStructureID = syllab.GetUInt64(buf, 32)
	ft.RecordSize = syllab.GetUInt64(buf, 40)
	ft.WriteTime = syllab.GetInt64(buf, 48)
	copy(ft.OwnerAppID[:], buf[56:])

	copy(ft.AppInstanceID[:], buf[72:])
	copy(ft.UserConnectionID[:], buf[88:])
	copy(ft.OwnerID[:], buf[104:])
	ft.Currency = syllab.GetUInt16(buf, 120)
	copy(ft.ReferenceID[:], buf[122:])
	ft.ReferenceType = syllab.GetUInt8(buf, 138)
	copy(ft.PreviousTransactionID[:], buf[139:])
	ft.Amount = syllab.GetInt64(buf, 171)
	ft.Balance = syllab.GetUInt64(buf, 179)
	return
}

func (ft *FinancialTransaction) syllabEncoder() (buf []byte) {
	buf = make([]byte, ft.syllabLen())

	// copy(buf[0:], ft.RecordID[:])
	syllab.SetUInt64(buf, 32, ft.RecordStructureID)
	syllab.SetUInt64(buf, 40, ft.RecordSize)
	syllab.SetInt64(buf, 48, ft.WriteTime)
	copy(buf[56:], ft.OwnerAppID[:])

	copy(buf[72:], ft.AppInstanceID[:])
	copy(buf[88:], ft.UserConnectionID[:])
	copy(buf[104:], ft.OwnerID[:])
	syllab.SetUInt16(buf, 120, ft.Currency)
	copy(buf[122:], ft.ReferenceID[:])
	syllab.SetUInt8(buf, 138, ft.ReferenceType)
	copy(buf[139:], ft.PreviousTransactionID[:])
	syllab.SetInt64(buf, 171, ft.Amount)
	syllab.SetUInt64(buf, 179, ft.Balance)
	return
}

func (ft *FinancialTransaction) syllabStackLen() (ln uint32) {
	return 187 // 72 + 115 + (0 * 8) >> Common header + Unique data + vars add&&len
}

func (ft *FinancialTransaction) syllabHeapLen() (ln uint32) {
	return
}

func (ft *FinancialTransaction) syllabLen() (ln uint64) {
	return uint64(ft.syllabStackLen() + ft.syllabHeapLen())
}
