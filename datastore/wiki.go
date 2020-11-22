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
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	ID               [32]byte `ganjine:"Unique" hash-index:"RecordID[pair,Language],Language"` // Unique content ID in all languages!
	OrgID            [32]byte `hash-index:"ID"`

	Language lang.Language
	URI      string `hash-index:"ID"` // https://en.wikipedia.org/wiki/Uniform_Resource_Identifier && https://en.wikipedia.org/wiki/Uniform_Resource_Name && https://en.wikipedia.org/wiki/Electronic_Product_Code
	Title    string `hash-index:"ID"` // It can be not unique in all wiki content.
	Text     string `hash-index:"ID"` // Text With Style. HTML & CSS is more expressive than markdown, so we use them in article text to style text.
	Pictures [][32]byte
	Status   WikiStatus
}

// SaveNew method set some data and write entire Wiki record with all indexes!
func (w *Wiki) SaveNew() (err *er.Error) {
	err = w.Set()
	if err != nil {
		return
	}
	w.HashIndexRecordIDForIDLanguage()
	w.HashListLanguageForID()
	w.HashIndexIDForURI()
	w.HashIndexIDForOrgID()
	w.HashIndexIDForTitle()
	return
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
		RecordID:          w.RecordID,
		RecordStructureID: wikiStructureID,
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
		err = ganjine.ErrMisMatchedStructureID
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
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", wikiStructureID)
	}
	return
}

// GetRecordsIDByIDLanguageByHashIndex find RecordsID by given ID
func (w *Wiki) GetRecordsIDByIDLanguageByHashIndex(offset, limit uint64) (RecordsID [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: w.hashIDLanguageforRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	RecordsID = indexRes.IndexValues
	return
}

// FindIDsByOrgIDByHashIndex find IDs by given OrgID
func (w *Wiki) FindIDsByOrgIDByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: w.hashOrgIDforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// GetIDsByURIByHashIndex find IDs by given URI
func (w *Wiki) GetIDsByURIByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: w.hashURIForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// GetIDsByTitleByHashIndex find IDs by given Title
func (w *Wiki) GetIDsByTitleByHashIndex(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: w.hashTitleforID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	IDs = indexRes.IndexValues
	return
}

// GetLanguagesByIDByHashIndex find languages by given ID
func (w *Wiki) GetLanguagesByIDByHashIndex(offset, limit uint64) (languages []lang.Language, err *er.Error) {
	var indexReq = &gs.HashIndexGetValuesReq{
		IndexKey: w.hashIDforLanguage(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *gs.HashIndexGetValuesRes
	indexRes, err = gsdk.HashIndexGetValues(cluster, indexReq)
	languages = lang.Unsafe32ByteArraySliceToLanguagesSlice(indexRes.IndexValues)
	return
}

/*
	-- PRIMARY INDEXES --
*/

// HashIndexRecordIDForIDLanguage save RecordID chain for ID+Language
// Call in each update to the exiting record!
func (w *Wiki) HashIndexRecordIDForIDLanguage() {
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
	const field = "IDLanguage"
	var buf = make([]byte, 44+len(field)) // 8+32+4
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.ID[:])
	syllab.SetUInt32(buf, 40, uint32(w.Language))
	copy(buf[44:], field)
	return sha512.Sum512_256(buf[:])
}

/*
	-- SECONDARY INDEXES --
*/

// HashIndexIDForOrgID save ID chain for OrgID.
// Don't call in update to an exiting record!
func (w *Wiki) HashIndexIDForOrgID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   w.hashOrgIDforID(),
		IndexValue: w.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (w *Wiki) hashOrgIDforID() (hash [32]byte) {
	const field = "OrgID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], w.OrgID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForURI save ID chain for URI.
// Don't call in update to an exiting record!
func (w *Wiki) HashIndexIDForURI() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   w.hashURIForID(),
		IndexValue: w.ID,
	}
	var err = gsdk.HashIndexSetValue(cluster, &indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (w *Wiki) hashURIForID() (hash [32]byte) {
	const field = "URI"
	var buf = make([]byte, 8+len(field)+len(w.URI))
	syllab.SetUInt64(buf, 0, wikiStructureID)
	copy(buf[8:], field)
	copy(buf[8+len(field):], w.URI)
	return sha512.Sum512_256(buf)
}

// HashIndexIDForTitle save ID chain for Title.
func (w *Wiki) HashIndexIDForTitle() {
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

// HashListLanguageForID list Language chain for ID.
// Don't call in update to an exiting record!
func (w *Wiki) HashListLanguageForID() {
	var indexRequest = gs.HashIndexSetValueReq{
		Type:     gs.RequestTypeBroadcast,
		IndexKey: w.hashIDforLanguage(),
	}
	syllab.SetUInt32(indexRequest.IndexValue[:], 0, uint32(w.Language))
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
	w.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(w.OwnerAppID[:], buf[56:])

	copy(w.AppInstanceID[:], buf[88:])
	copy(w.UserConnectionID[:], buf[120:])
	copy(w.ID[:], buf[152:])
	copy(w.OrgID[:], buf[184:])

	w.Language = lang.Language(syllab.GetUInt32(buf, 216))
	w.URI = syllab.UnsafeGetString(buf, 220)
	w.Title = syllab.UnsafeGetString(buf, 228)
	w.Text = syllab.UnsafeGetString(buf, 236)
	w.Pictures = syllab.UnsafeGet32ByteArraySlice(buf, 244)
	w.Status = WikiStatus(syllab.GetUInt8(buf, 252))
	return
}

func (w *Wiki) syllabEncoder() (buf []byte) {
	buf = make([]byte, w.syllabLen())
	var hsi uint32 = w.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], w.RecordID[:])
	syllab.SetUInt64(buf, 32, w.RecordStructureID)
	syllab.SetUInt64(buf, 40, w.RecordSize)
	syllab.SetInt64(buf, 48, int64(w.WriteTime))
	copy(buf[56:], w.OwnerAppID[:])

	copy(buf[88:], w.AppInstanceID[:])
	copy(buf[120:], w.UserConnectionID[:])
	copy(buf[152:], w.ID[:])
	copy(buf[184:], w.OrgID[:])

	syllab.SetUInt32(buf, 216, uint32(w.Language))
	hsi = syllab.SetString(buf, w.URI, 220, hsi)
	hsi = syllab.SetString(buf, w.Title, 228, hsi)
	hsi = syllab.SetString(buf, w.Text, 236, hsi)
	hsi = syllab.Set32ByteArrayArray(buf, w.Pictures, 244, hsi)
	syllab.SetUInt8(buf, 252, uint8(w.Status))
	return
}

func (w *Wiki) syllabStackLen() (ln uint32) {
	return 253
}

func (w *Wiki) syllabHeapLen() (ln uint32) {
	ln += uint32(len(w.URI))
	ln += uint32(len(w.Title))
	ln += uint32(len(w.Text))
	ln += uint32(len(w.Pictures) * 32)
	return
}

func (w *Wiki) syllabLen() (ln uint64) {
	return uint64(w.syllabStackLen() + w.syllabHeapLen())
}

/*
	-- Record types --
*/

// WikiStatus indicate Wiki record status
type WikiStatus uint8

// Wiki status
const (
	WikiStatusUnset = iota
	WikiStatusRegister
	WikiStatusSuggestion
	WikiStatusBlocked
)
