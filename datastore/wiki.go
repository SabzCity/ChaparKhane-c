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
	wikiStructureID uint64 = 4150904594571984896
)

var wikiStructure = ganjine.DataStructure{
	ID:                4150904594571984896,
	IssueDate:         1599455751,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         Wiki{},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Wiki",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: `store any text && pictures definition information about any topic!`,
	},
	TAGS: []string{
		"",
	},
}

// Wiki ---Read locale description in wikiStructure---
type Wiki struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         int64
	OwnerAppID        [16]byte

	/* Unique data */
	AppInstanceID    [16]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [16]byte // Store to remember which user connection set||chanaged this record!
	ID               [16]byte `ganjine:"Immutable,Unique" ganjine-index:"OwnerID,Title,Text" ganjine-list:"Language"` // Unique content ID in all languages!
	OwnerID          [16]byte `ganjine:"Immutable"`
	Language         uint32   `ganjine:"Immutable,Unique"` // language package
	Title            string   // It can be not unique in all wiki content.
	Text             string   // Text With Style. HTML & CSS is more expressive than markdown, so we use them in article text to style text.
	Pictures         [][16]byte
	Status           uint8 // Suggestion, Active,
}

// Set method set some data and write entire Wiki record!
func (w *Wiki) Set() (err error) {
	w.RecordStructureID = wikiStructureID
	w.RecordSize = w.syllabLen()
	w.WriteTime = etime.Now()
	w.OwnerAppID = server.Manifest.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: w.syllabEncoder(),
	}
	w.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], w.RecordID[:])

	err = gsdk.SetRecord(cluster, &req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (w *Wiki) GetByRecordID() (err error) {
	var req = gs.GetRecordReq{
		RecordID: w.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return err
	}

	err = w.syllabDecoder(res.Record)
	if err != nil {
		return err
	}

	if w.RecordStructureID != wikiStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetByIDLang method find and read last version of record by given ID+Lang
func (w *Wiki) GetByIDLang() (err error) {
	var indexReq = &gs.FindRecordsReq{
		IndexHash: w.HashIDLang(),
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
		w.RecordID = indexRes.RecordIDs[ln]
		err = w.GetByRecordID()
		if err != ganjine.ErrGanjineMisMatchedStructureID {
			return
		}
	}
	return ganjine.ErrGanjineRecordNotFound
}

/*
	-- PRIMARY INDEXES --
*/

// IndexIDLang index Unique-Field(ID+Language) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (w *Wiki) IndexIDLang() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: w.HashIDLang(),
		RecordID:  w.RecordID,
	}
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashIDLang hash wikiStructureID + w.ID + w.Language
func (w *Wiki) HashIDLang() (hash [32]byte) {
	var buf = make([]byte, 28) // 8+16+4
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.ID[:])
	syllab.SetUInt32(buf, 24, w.Language)
	return sha512.Sum512_256(buf[:])
}

/*
	-- SECONDARY INDEXES --
*/

// IndexOwner index to retrieve all Unique-Field(ID) owned by given OwnerID later.
// Don't call in update to an exiting record!
func (w *Wiki) IndexOwner() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: w.HashOwner(),
	}
	copy(indexRequest.RecordID[:], w.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashOwner hash wikiStructureID + w.OwnerID
func (w *Wiki) HashOwner() (hash [32]byte) {
	var buf = make([]byte, 24) // 8+16
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.OwnerID[:])
	return sha512.Sum512_256(buf)
}

// IndexTitleString index w.Title to retrieve record fast later.
func (w *Wiki) IndexTitleString() {
	// TODO::: Title parts must index too!
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: w.HashTitleString(),
	}
	copy(indexRequest.RecordID[:], w.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashTitleString hash wikiStructureID + w.Title that it can be just some word not whole title
func (w *Wiki) HashTitleString() (hash [32]byte) {
	var buf = make([]byte, 9+len(w.Title)) // 8+1
	syllab.SetUInt64(buf, 0, wikiStructureID)
	syllab.SetByte(buf, 8, 'T')
	copy(buf[9:], w.Title)
	return sha512.Sum512_256(buf)
}

// IndexTextString index w.Text to retrieve record fast later.
func (w *Wiki) IndexTextString() {
	// TODO::: ???

	// var indexRequest = gs.SetIndexHashReq{
	// 	Type:      gs.RequestTypeBroadcast,
	// 	IndexHash: w.HashTextString(),
	// 	RecordID:  w.RecordID,
	// }
	// var err = gsdk.SetIndexHash(cluster, &indexRequest)
	// if err != nil {
	// 	// TODO::: we must retry more due to record wrote successfully!
	// }
}

// HashTextString hash wikiStructureID + w.Text that it can be just some word not whole article
func (w *Wiki) HashTextString() (hash [32]byte) {
	var buf = make([]byte, 9+len(w.Text)) // 8+1
	syllab.SetUInt64(buf, 0, wikiStructureID)
	buf[8] = byte('A')
	copy(buf[9:], w.Text)
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListIDLang store all w.Language own by specific w.ID.
// Don't call in update to an exiting record!
func (w *Wiki) ListIDLang() {
	var indexRequest = gs.SetIndexHashReq{
		Type:      gs.RequestTypeBroadcast,
		IndexHash: w.HashIDLangField(),
	}
	copy(indexRequest.RecordID[:], w.ID[:])
	var err = gsdk.SetIndexHash(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

// HashIDLangField hash wikiStructureID + w.ID + "Language" field
func (w *Wiki) HashIDLangField() (hash [32]byte) {
	const field = "Language"
	var buf = make([]byte, 24+len(field)) // 8+16
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.ID[:])
	copy(buf[24:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (w *Wiki) syllabDecoder(buf []byte) (err error) {
	var add, ln uint32

	if uint32(len(buf)) < w.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(w.RecordID[:], buf[0:])
	w.RecordStructureID = syllab.GetUInt64(buf, 32)
	w.RecordSize = syllab.GetUInt64(buf, 40)
	w.WriteTime = syllab.GetInt64(buf, 48)
	copy(w.OwnerAppID[:], buf[56:])

	copy(w.AppInstanceID[:], buf[72:])
	copy(w.UserConnectionID[:], buf[88:])
	copy(w.ID[:], buf[104:])
	copy(w.OwnerID[:], buf[120:])
	w.Language = syllab.GetUInt32(buf, 136)
	add = syllab.GetUInt32(buf, 140)
	ln = syllab.GetUInt32(buf, 144)
	w.Title = string(buf[add : add+ln])
	add = syllab.GetUInt32(buf, 148)
	ln = syllab.GetUInt32(buf, 152)
	w.Text = string(buf[add : add+ln])
	w.Status = syllab.GetUInt8(buf, 164)
	return
}

func (w *Wiki) syllabEncoder() (buf []byte) {
	buf = make([]byte, w.syllabLen())
	var hsi uint32 = w.syllabStackLen() // Heap start index || Stack size!
	var i, ln uint32                    // len of strings, slices, maps, ...

	// copy(buf[0:], w.RecordID[:])
	syllab.SetUInt64(buf, 32, w.RecordStructureID)
	syllab.SetUInt64(buf, 40, w.RecordSize)
	syllab.SetInt64(buf, 48, w.WriteTime)
	copy(buf[56:], w.OwnerAppID[:])

	copy(buf[72:], w.AppInstanceID[:])
	copy(buf[88:], w.UserConnectionID[:])
	copy(buf[104:], w.ID[:])
	copy(buf[120:], w.OwnerID[:])

	syllab.SetUInt32(buf, 136, w.Language)
	ln = uint32(len(w.Title))
	syllab.SetUInt32(buf, 140, hsi)
	syllab.SetUInt32(buf, 144, ln)
	copy(buf[hsi:], w.Title)
	hsi += ln
	ln = uint32(len(w.Text))
	syllab.SetUInt32(buf, 148, hsi)
	syllab.SetUInt32(buf, 152, ln)
	copy(buf[hsi:], w.Text)
	hsi += ln
	ln = uint32(len(w.Pictures))
	syllab.SetUInt32(buf, 156, hsi)
	syllab.SetUInt32(buf, 160, ln)
	for i = 0; i < ln; i++ {
		copy(buf[hsi:], w.Pictures[i][:])
		hsi += 16
	}
	syllab.SetUInt8(buf, 164, w.Status)
	return
}

func (w *Wiki) syllabStackLen() (ln uint32) {
	return 165 // 72 + 69 + (3 * 8) >> Common header + Unique data + vars add&&len
}

func (w *Wiki) syllabHeapLen() (ln uint32) {
	ln += uint32(len(w.Title))
	ln += uint32(len(w.Text))
	ln += uint32(len(w.Pictures) * 16)
	return
}

func (w *Wiki) syllabLen() (ln uint64) {
	return uint64(w.syllabStackLen() + w.syllabHeapLen())
}
