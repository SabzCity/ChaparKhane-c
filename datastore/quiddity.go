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
	"../libgo/matn"
	"../libgo/pehrest"
	psdk "../libgo/pehrest-sdk"
	"../libgo/syllab"
)

const (
	quiddityStructureID uint64 = 1479548559340177913
)

var quiddityStructure = ganjine.DataStructure{
	ID:                1479548559340177913,
	IssueDate:         1599455751,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // Other structure name
	ExpireInFavorOfID: 0,  // Other StructureID! Handy ID or Hash of ExpireInFavorOf!
	Status:            ganjine.DataStructureStatePreAlpha,
	Structure:         Quiddity{},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Quiddity",
		lang.LanguagePersian: "ماهیت",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: `Quiddity is the essence of an object and with name or title of it in human languages and recently in machine languages`,
		lang.LanguagePersian: `ماهیت یعنی ذات و چیستی یک شئ که دارای نام یا عنوان مشخص در زبان های انسانی و جدیدا در زبان های ماشین می باشد`,
	},
	TAGS: []string{
		"",
	},
}

// Quiddity ---Read locale description in quiddityStructure---
type Quiddity struct {
	/* Common header data */
	RecordID          [32]byte
	RecordStructureID uint64
	RecordSize        uint64
	WriteTime         etime.Time
	OwnerAppID        [32]byte

	/* Unique data */
	AppInstanceID    [32]byte // Store to remember which app instance set||chanaged this record!
	UserConnectionID [32]byte // Store to remember which user connection set||chanaged this record!
	ID               [32]byte `index-hash:"RecordID[pair,Language],Language"` // Unique content ID in all languages!
	OrgID            [32]byte `index-hash:"ID"`

	Language lang.Language
	URI      string `index-hash:"ID"` // Locale name in the Computer world!!	https://en.quidditypedia.org/quiddity/Uniform_Resource_Identifier && https://en.quidditypedia.org/quiddity/Uniform_Resource_Name && https://en.quidditypedia.org/quiddity/Electronic_Product_Code
	Title    string `index-text:"ID"` // Locale name in the Human world!!		It can be not unique in all quiddity content.
	Status   QuiddityStatus
}

// SaveNew method set some data and write entire Quiddity record with all indexes!
func (q *Quiddity) SaveNew() (err *er.Error) {
	err = q.Set()
	if err != nil {
		return
	}
	q.IndexRecordIDForIDLanguage()
	q.IndexIDForOrgID()
	q.IndexIDForURI()
	q.IndexIDForTitle()
	q.ListLanguageForID()
	return
}

// Set method set some data and write entire Quiddity record!
func (q *Quiddity) Set() (err *er.Error) {
	q.RecordStructureID = quiddityStructureID
	q.RecordSize = q.syllabLen()
	q.WriteTime = etime.Now()
	q.OwnerAppID = achaemenid.Server.AppID

	var req = gs.SetRecordReq{
		Type:   gs.RequestTypeBroadcast,
		Record: q.syllabEncoder(),
	}
	q.RecordID = sha512.Sum512_256(req.Record[32:])
	copy(req.Record[0:], q.RecordID[:])

	err = gsdk.SetRecord(&req)
	if err != nil {
		// TODO::: Handle error situation
	}

	return
}

// GetByRecordID method read all existing record data by given RecordID!
func (q *Quiddity) GetByRecordID() (err *er.Error) {
	var req = gs.GetRecordReq{
		RecordID:          q.RecordID,
		RecordStructureID: quiddityStructureID,
	}
	var res *gs.GetRecordRes
	res, err = gsdk.GetRecord(&req)
	if err != nil {
		return
	}

	err = q.syllabDecoder(res.Record)
	if err != nil {
		return
	}

	if q.RecordStructureID != quiddityStructureID {
		err = ganjine.ErrMisMatchedStructureID
	}
	return
}

// GetLastByIDLang method find and read last version of record by given ID+Lang
func (q *Quiddity) GetLastByIDLang() (err *er.Error) {
	var RecordsID [][32]byte
	RecordsID, err = q.FindRecordsIDByIDLanguage(18446744073709551615, 1)
	if err != nil {
		return
	}

	q.RecordID = RecordsID[0]
	err = q.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", quiddityStructureID)
	}
	return
}

// GetLastByURI find and read last version of record by given URI
func (q *Quiddity) GetLastByURI() (err *er.Error) {
	var RecordsID [][32]byte
	RecordsID, err = q.FindIDsByURI(18446744073709551615, 1)
	if err != nil {
		return
	}

	q.RecordID = RecordsID[0]
	err = q.GetByRecordID()
	if err.Equal(ganjine.ErrMisMatchedStructureID) {
		log.Warn("Platform collapsed!! HASH Collision Occurred on", quiddityStructureID)
	}
	return
}

/*
	-- Search Methods --
*/

// FindRecordsIDByIDLanguage find RecordsID by given ID
func (q *Quiddity) FindRecordsIDByIDLanguage(offset, limit uint64) (RecordsID [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: q.hashIDLanguageforRecordID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	RecordsID = indexRes.IndexValues
	return
}

// FindIDsByOrgID find IDs by given OrgID
func (q *Quiddity) FindIDsByOrgID(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: q.hashOrgIDForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByURI find IDs by given URI
func (q *Quiddity) FindIDsByURI(offset, limit uint64) (IDs [][32]byte, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: q.hashURIForID(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	IDs = indexRes.IndexValues
	return
}

// FindIDsByTitle find IDs by given Title
func (q *Quiddity) FindIDsByTitle(pageNumber uint64) (tokens *matn.IndexTextFindRes, err *er.Error) {
	var indexReq = matn.IndexTextFindReq{
		Term:            q.Title,
		RecordStructure: quiddityStructureID,
		PageNumber:      pageNumber,
	}
	tokens, err = matn.IndexTextFind(&indexReq)
	return
}

// FindLanguagesByID find languages by given ID
func (q *Quiddity) FindLanguagesByID(offset, limit uint64) (languages []lang.Language, err *er.Error) {
	var indexReq = &pehrest.HashGetValuesReq{
		IndexKey: q.hashIDForLanguage(),
		Offset:   offset,
		Limit:    limit,
	}
	var indexRes *pehrest.HashGetValuesRes
	indexRes, err = psdk.HashGetValues(indexReq)
	languages = lang.Unsafe32ByteArraySliceToLanguagesSlice(indexRes.IndexValues)
	return
}

/*
	-- PRIMARY INDEXES --
*/

// IndexRecordIDForIDLanguage save RecordID chain for ID+Language
// Call in each update to the exiting record!
func (q *Quiddity) IndexRecordIDForIDLanguage() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   q.hashIDLanguageforRecordID(),
		IndexValue: q.RecordID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (q *Quiddity) hashIDLanguageforRecordID() (hash [32]byte) {
	const field = "IDLanguage"
	var buf = make([]byte, 44+len(field)) // 8+32+4
	syllab.SetUInt64(buf, 0, quiddityStructureID)
	copy(buf[8:], q.ID[:])
	syllab.SetUInt32(buf, 40, uint32(q.Language))
	copy(buf[44:], field)
	return sha512.Sum512_256(buf[:])
}

/*
	-- SECONDARY INDEXES --
*/

// IndexIDForOrgID save ID chain for OrgID.
// Don't call in update to an exiting record!
func (q *Quiddity) IndexIDForOrgID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   q.hashOrgIDForID(),
		IndexValue: q.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (q *Quiddity) hashOrgIDForID() (hash [32]byte) {
	const field = "OrgID"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, quiddityStructureID)
	copy(buf[8:], q.OrgID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

// IndexIDForURI save ID chain for URI.
// Don't call in update to an exiting record!
func (q *Quiddity) IndexIDForURI() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:       gs.RequestTypeBroadcast,
		IndexKey:   q.hashURIForID(),
		IndexValue: q.ID,
	}
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (q *Quiddity) hashURIForID() (hash [32]byte) {
	const field = "URI"
	var buf = make([]byte, 8+len(field)+len(q.URI))
	syllab.SetUInt64(buf, 0, quiddityStructureID)
	copy(buf[8:], field)
	copy(buf[8+len(field):], q.URI)
	return sha512.Sum512_256(buf)
}

// IndexIDForTitle save ID chain for Title.
func (q *Quiddity) IndexIDForTitle() {
	var indexRequest = matn.TextIndexReq{
		RecordID:         q.RecordID,
		RecordStructure:  quiddityStructureID,
		RecordPrimaryKey: q.ID,
		// RecordSecondaryKey: q.,
		RecordOwnerID: q.OrgID,
		RecordFieldID: 11,
		Text:          q.Title,
	}
	var err = matn.TextIndex(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

/*
	-- LIST FIELDS --
*/

// ListLanguageForID list Language chain for ID.
// Don't call in update to an exiting record!
func (q *Quiddity) ListLanguageForID() {
	var indexRequest = pehrest.HashSetValueReq{
		Type:     gs.RequestTypeBroadcast,
		IndexKey: q.hashIDForLanguage(),
	}
	syllab.SetUInt32(indexRequest.IndexValue[:], 0, uint32(q.Language))
	var err = psdk.HashSetValue(&indexRequest)
	if err != nil {
		// TODO::: we must retry more due to record wrote successfully!
	}
}

func (q *Quiddity) hashIDForLanguage() (hash [32]byte) {
	const field = "ListLanguage"
	var buf = make([]byte, 40+len(field)) // 8+32
	syllab.SetUInt64(buf, 0, quiddityStructureID)
	copy(buf[8:], q.ID[:])
	copy(buf[40:], field)
	return sha512.Sum512_256(buf)
}

/*
	-- Syllab Encoder & Decoder --
*/

func (q *Quiddity) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < q.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(q.RecordID[:], buf[0:])
	q.RecordStructureID = syllab.GetUInt64(buf, 32)
	q.RecordSize = syllab.GetUInt64(buf, 40)
	q.WriteTime = etime.Time(syllab.GetInt64(buf, 48))
	copy(q.OwnerAppID[:], buf[56:])

	copy(q.AppInstanceID[:], buf[88:])
	copy(q.UserConnectionID[:], buf[120:])
	copy(q.ID[:], buf[152:])
	copy(q.OrgID[:], buf[184:])

	q.Language = lang.Language(syllab.GetUInt32(buf, 216))
	q.URI = syllab.UnsafeGetString(buf, 220)
	q.Title = syllab.UnsafeGetString(buf, 228)
	q.Status = QuiddityStatus(syllab.GetUInt8(buf, 236))
	return
}

func (q *Quiddity) syllabEncoder() (buf []byte) {
	buf = make([]byte, q.syllabLen())
	var hsi uint32 = q.syllabStackLen() // Heap start index || Stack size!

	// copy(buf[0:], q.RecordID[:])
	syllab.SetUInt64(buf, 32, q.RecordStructureID)
	syllab.SetUInt64(buf, 40, q.RecordSize)
	syllab.SetInt64(buf, 48, int64(q.WriteTime))
	copy(buf[56:], q.OwnerAppID[:])

	copy(buf[88:], q.AppInstanceID[:])
	copy(buf[120:], q.UserConnectionID[:])
	copy(buf[152:], q.ID[:])
	copy(buf[184:], q.OrgID[:])

	syllab.SetUInt32(buf, 216, uint32(q.Language))
	hsi = syllab.SetString(buf, q.URI, 220, hsi)
	hsi = syllab.SetString(buf, q.Title, 228, hsi)
	syllab.SetUInt8(buf, 236, uint8(q.Status))
	return
}

func (q *Quiddity) syllabStackLen() (ln uint32) {
	return 237
}

func (q *Quiddity) syllabHeapLen() (ln uint32) {
	ln += uint32(len(q.URI))
	ln += uint32(len(q.Title))
	return
}

func (q *Quiddity) syllabLen() (ln uint64) {
	return uint64(q.syllabStackLen() + q.syllabHeapLen())
}

/*
	-- Record types --
*/

// QuiddityStatus indicate Quiddity record status
type QuiddityStatus uint8

// Quiddity status
const (
	QuiddityStatusUnset = iota
	QuiddityStatusRegister
	QuiddityStatusSuggestion
	QuiddityStatusBlocked
)
