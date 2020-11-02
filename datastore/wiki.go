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
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	ID               [32]byte `ganjine:"Immutable,Unique" ganjine-index:"OwnerID,Title,Text" ganjine-list:"Language"` // Unique content ID in all languages!
	OwnerID          [32]byte `ganjine:"Immutable"`

	Language uint32 `ganjine:"Immutable,Unique"` // language package
	Title    string // It can be not unique in all wiki content.
	Text     string // Text With Style. HTML & CSS is more expressive than markdown, so we use them in article text to style text.
	Pictures [][32]byte
	Status   uint8 // Suggestion, Active,
}

// Set method set some data and write entire Wiki record!
func (w *Wiki) Set() (err *er.Error) {
	w.RecordStructureID = wikiStructureID
	w.RecordSize = w.syllabLen()
	w.WriteTime = etime.Now()
	w.OwnerAppID = server.AppID

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
func (w *Wiki) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID: w.RecordID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(cluster, &req)
	if err != nil {
		return
	}

	err = w.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if w.RecordStructureID != wikiStructureID {
		err = ganjine.ErrGanjineMisMatchedStructureID
	}
	return
}

// GetLastByIDLang method find and read last version of record by given ID+Lang
func (w *Wiki) GetLastByIDLang() (err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: w.hashIDLanguageforRecordID(),
		Offset:   18446744073709551615,
		Limit:    1,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	if err != nil {
		return
	}

	w.RecordID = indexRes.IndexValues[0]
	err = w.GetByRecordID()
	if err == ganjine.ErrGanjineMisMatchedStructureID {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", wikiStructureID)
	}
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexIDLang index Unique-Field(ID+Language) chain to retrieve last record version fast later.
// Call in each update to the exiting record!
func (w *Wiki) IndexIDLang() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   w.hashIDLanguageforRecordID(),
		IndexValue: w.RecordID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (w *Wiki) hashIDLanguageforRecordID() (hash [32]byte) {
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
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   w.hashOwnerIDforID(),
		IndexValue: w.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (w *Wiki) hashOwnerIDforID() (hash [32]byte) {
	var buf = make([]byte, 40) // 8+32
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.OwnerID[:])
	return sha512.Sum512_256(buf)
}

// IndexTitleString index w.Title to retrieve record fast later.
func (w *Wiki) IndexTitleString() {
	// TODO::: Title parts must index too!
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   w.hashTitleforID(),
		IndexValue: w.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (w *Wiki) hashTitleforID() (hash [32]byte) {
	const field = "Title"
	var buf = make([]byte, 8+len(field)+len(w.Title))
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], field)
	copy(buf[8+len(field):], w.Title)
	return sha512.Sum512_256(buf)
}

// IndexTextString index w.Text to retrieve record fast later.
func (w *Wiki) IndexTextString() {
	// TODO::: ???

	// var indexRequest = gs.HashIndexSetValueReq{
	// 	Type:      gs.RequestTypeBroadcast,
	// 	IndexKey: w.hashTextforID(),
	// 	IndexValue:  w.ID,
	// }
	// var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	// if err != nil {
	// 	// TODO::: we must retry more due to record wrote successfully!
	// }
}

func (w *Wiki) hashTextforID() (hash [32]byte) {
	const field = "Text"
	var buf = make([]byte, 8+len(field)+len(w.Text))
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], field)
	copy(buf[8+len(field):], w.Text)
	return sha512.Sum512_256(buf)
}

/*
	-- LIST FIELDS --
*/

// ListLangforID list all Language own by specific ID
// Don't call in update to an exiting record!
func (w *Wiki) ListLangforID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:     gs.RequestTypeBroadcast,
		IndexKey: w.hashIDforLanguage(),
	}
	syllab.SetUInt32(indexRequest.IndexValue[:], 0, w.Language)
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (w *Wiki) hashIDforLanguage() (hash [32]byte) {
	const field = "ListLanguage"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.ID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (w *Wiki) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < w.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(w.RecordID[:], buf[0:])
	w.RecordStructureID = syllab.GetUInt64(buf, 32)
	w.RecordSize = syllab.GetUInt64(buf, 40)
	w.WriteTime = syllab.GetInt64(buf, 48)
	copy(w.OwnerAppID[:], buf[56:])

	copy(w.AppInstanceID[:], buf[88:])
	copy(w.UserConnectionID[:], buf[120:])
	copy(w.ID[:], buf[152:])
	copy(w.OwnerID[:], buf[184:])

	w.Language = syllab.GetUInt32(buf, 216)
	w.Title = syllab.UnsafeGetString(buf, 220)
	w.Text = syllab.UnsafeGetString(buf, 228)
	w.Pictures = syllab.UnsafeGet32ByteArrayArray(buf, 236)
	w.Status = syllab.GetUInt8(buf, 244)
	return
}

func (w *Wiki) syllabEncoder() (buf []byte) {
	buf = make([]byte, w.syllabLen())
	var hsi uint32 = w.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], w.RecordID[:])
	syllab.SetUInt64(buf, 32, w.RecordStructureID)
	syllab.SetUInt64(buf, 40, w.RecordSize)
	syllab.SetInt64(buf, 48, w.WriteTime)
	copy(buf[56:], w.OwnerAppID[:])

	copy(buf[88:], w.AppInstanceID[:])
	copy(buf[120:], w.UserConnectionID[:])
	copy(buf[152:], w.ID[:])
	copy(buf[184:], w.OwnerID[:])

	syllab.SetUInt32(buf, 216, w.Language)
	hsi = syllab.SetString(buf, w.Title, 220, hsi)
	hsi = syllab.SetString(buf, w.Text, 228, hsi)
	syllab.Set32ByteArrayArray(buf, w.Pictures, 156, hsi)
	syllab.SetUInt8(buf, 244, w.Status)
	return
}

func (w *Wiki) syllabStackLen() (ln uint32) {
	return 254
}

func (w *Wiki) syllabHeapLen() (ln uint32) {
	ln += uint32(len(w.Title))
	ln += uint32(len(w.Text))
	ln += uint32(len(w.Pictures) * 32)
	return
}

func (w *Wiki) syllabLen() (ln uint64) {
	return uint64(w.syllabStackLen() + w.syllabHeapLen())
}
