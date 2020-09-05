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
	financialTransactionStructureID uint64 = 15981238345012607782
	financialTransactionFixedSize   uint64 = 187 // 72 + 115 + (0 * 8) >> Common header + Unique data + vars add&&len
	financialTransactionState       uint8  = ganjine.DataStructureStatePreAlpha
)

// FinancialTransaction store each financial transaction.
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
	UserID                [16]byte
	Currency              uint16
	ReferenceID           [16]byte // ProductID || UserID Transferred || ForeignExchangeID || JusticeID || ...
	ReferenceType         uint8    // 0:Product 1:Transfer 2:Transfer-bank 3:Block 4:donate
	PreviousTransactionID [32]byte // Last RecordID of same user with same currency
	Amount                int64    // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
	Balance               uint64   // Some number part base on currency is Decimal part e.g. 8099 >> 80.99$
}

// Set method set some data and write entire FinancialTransaction record!
func (ft *FinancialTransaction) Set() (err error) {
	ft.RecordStructureID = financialTransactionStructureID
	ft.RecordSize = financialTransactionFixedSize
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
		err = ganjine.ErrRecordMisMatchedStructureID
	}
	return
}

// GetByUserCurrency method find and read last version of record by given UserID + Currency
func (ft *FinancialTransaction) GetByUserCurrency() (err error) {
	// TODO::: set this month in ft!!
	// ft.WriteTime =
	var indexReq = &gs.FindRecordsReq{
		IndexHash: ft.HashUserCurrencyMonthly(),
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
		ft.RecordID = indexRes.RecordIDs[ln]
		err = ft.GetByRecordID()
		if err != ganjine.ErrRecordMisMatchedStructureID {
			return
		}
	}
}

// IndexUserCurrencyMonthly index ft.UserID + ft.Currency on monthly base to retrieve record fast later.
func (ft *FinancialTransaction) IndexUserCurrencyMonthly() {
	var userCurrencyMonthlyIndex = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: ft.HashUserCurrencyMonthly(),
		RecordID:  ft.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &userCurrencyMonthlyIndex)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashUserCurrencyMonthly hash financialTransactionStructureID + ft.UserID + ft.Currency
func (ft *FinancialTransaction) HashUserCurrencyMonthly() (hash [32]byte) {
	var buf = make([]byte, 34) // 8+8+16+2

	buf[0] = byte(ft.RecordStructureID)
	buf[1] = byte(ft.RecordStructureID >> 8)
	buf[2] = byte(ft.RecordStructureID >> 16)
	buf[3] = byte(ft.RecordStructureID >> 24)
	buf[4] = byte(ft.RecordStructureID >> 32)
	buf[5] = byte(ft.RecordStructureID >> 40)
	buf[6] = byte(ft.RecordStructureID >> 48)
	buf[7] = byte(ft.RecordStructureID >> 56)

	var roundedTime = etime.RoundToMonth(ft.WriteTime)
	buf[8] = byte(roundedTime)
	buf[9] = byte(roundedTime >> 8)
	buf[10] = byte(roundedTime >> 16)
	buf[11] = byte(roundedTime >> 24)
	buf[12] = byte(roundedTime >> 32)
	buf[13] = byte(roundedTime >> 40)
	buf[14] = byte(roundedTime >> 48)
	buf[15] = byte(roundedTime >> 56)

	copy(buf[16:], ft.UserID[:])

	buf[32] = byte(ft.Currency)
	buf[33] = byte(ft.Currency >> 8)

	return sha512.Sum512_256(buf)
}

func (ft *FinancialTransaction) syllabDecoder(buf []byte) (err error) {
	if uint64(len(buf)) < financialTransactionFixedSize {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	copy(ft.RecordID[:], buf[:])
	ft.RecordStructureID = uint64(buf[32]) | uint64(buf[33])<<8 | uint64(buf[34])<<16 | uint64(buf[35])<<24 | uint64(buf[36])<<32 | uint64(buf[37])<<40 | uint64(buf[38])<<48 | uint64(buf[39])<<56
	ft.RecordSize = uint64(buf[40]) | uint64(buf[41])<<8 | uint64(buf[42])<<16 | uint64(buf[43])<<24 | uint64(buf[44])<<32 | uint64(buf[45])<<40 | uint64(buf[46])<<48 | uint64(buf[47])<<56
	ft.WriteTime = int64(buf[48]) | int64(buf[49])<<8 | int64(buf[50])<<16 | int64(buf[51])<<24 | int64(buf[52])<<32 | int64(buf[53])<<40 | int64(buf[54])<<48 | int64(buf[55])<<56
	copy(ft.OwnerAppID[:], buf[56:])

	copy(ft.AppInstanceID[:], buf[72:])
	copy(ft.UserConnectionID[:], buf[88:])
	copy(ft.UserID[:], buf[104:])
	ft.Currency = uint16(buf[120]) | uint16(buf[121])<<8
	copy(ft.ReferenceID[:], buf[122:])
	ft.ReferenceType = uint8(buf[138])
	copy(ft.PreviousTransactionID[:], buf[139:])
	ft.Amount = int64(buf[171]) | int64(buf[172])<<8 | int64(buf[173])<<16 | int64(buf[174])<<24 | int64(buf[175])<<32 | int64(buf[176])<<40 | int64(buf[177])<<48 | int64(buf[178])<<56
	ft.Balance = uint64(buf[179]) | uint64(buf[180])<<8 | uint64(buf[181])<<16 | uint64(buf[182])<<24 | uint64(buf[183])<<32 | uint64(buf[184])<<40 | uint64(buf[185])<<48 | uint64(buf[186])<<56

	return
}

func (ft *FinancialTransaction) syllabEncoder() (buf []byte) {
	buf = make([]byte, financialTransactionFixedSize)

	// copy(buf[0:], ft.RecordID[:])
	buf[32] = byte(ft.RecordStructureID)
	buf[33] = byte(ft.RecordStructureID >> 8)
	buf[34] = byte(ft.RecordStructureID >> 16)
	buf[35] = byte(ft.RecordStructureID >> 24)
	buf[36] = byte(ft.RecordStructureID >> 32)
	buf[37] = byte(ft.RecordStructureID >> 40)
	buf[38] = byte(ft.RecordStructureID >> 48)
	buf[39] = byte(ft.RecordStructureID >> 56)
	buf[40] = byte(ft.RecordSize)
	buf[41] = byte(ft.RecordSize >> 8)
	buf[42] = byte(ft.RecordSize >> 16)
	buf[43] = byte(ft.RecordSize >> 24)
	buf[44] = byte(ft.RecordSize >> 32)
	buf[45] = byte(ft.RecordSize >> 40)
	buf[46] = byte(ft.RecordSize >> 48)
	buf[47] = byte(ft.RecordSize >> 56)
	buf[48] = byte(ft.WriteTime)
	buf[49] = byte(ft.WriteTime >> 8)
	buf[50] = byte(ft.WriteTime >> 16)
	buf[51] = byte(ft.WriteTime >> 24)
	buf[52] = byte(ft.WriteTime >> 32)
	buf[53] = byte(ft.WriteTime >> 40)
	buf[54] = byte(ft.WriteTime >> 48)
	buf[55] = byte(ft.WriteTime >> 56)
	copy(buf[56:], ft.OwnerAppID[:])

	copy(buf[72:], ft.AppInstanceID[:])
	copy(buf[88:], ft.UserConnectionID[:])
	copy(buf[104:], ft.UserID[:])
	buf[120] = byte(ft.Currency)
	buf[121] = byte(ft.Currency >> 8)
	copy(buf[122:], ft.ReferenceID[:])
	buf[138] = byte(ft.ReferenceType)
	copy(buf[139:], ft.PreviousTransactionID[:])
	buf[171] = byte(ft.Amount)
	buf[172] = byte(ft.Amount >> 8)
	buf[173] = byte(ft.Amount >> 16)
	buf[174] = byte(ft.Amount >> 24)
	buf[175] = byte(ft.Amount >> 32)
	buf[176] = byte(ft.Amount >> 40)
	buf[177] = byte(ft.Amount >> 48)
	buf[178] = byte(ft.Amount >> 56)
	buf[179] = byte(ft.Balance)
	buf[180] = byte(ft.Balance >> 8)
	buf[181] = byte(ft.Balance >> 16)
	buf[182] = byte(ft.Balance >> 24)
	buf[183] = byte(ft.Balance >> 32)
	buf[184] = byte(ft.Balance >> 40)
	buf[185] = byte(ft.Balance >> 48)
	buf[186] = byte(ft.Balance >> 56)

	return
}
