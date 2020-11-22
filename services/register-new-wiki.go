/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/ganjine"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/uuid"
	"../libgo/validators"
)

var registerNewWikiService = achaemenid.Service{
	ID:                252755339,
	IssueDate:         1604939795,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Register New Wiki",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: RegisterNewWikiSRPC,
	HTTPHandler: RegisterNewWikiHTTP,
}

// RegisterNewWikiSRPC is sRPC handler of RegisterNewWiki service.
func RegisterNewWikiSRPC(st *achaemenid.Stream) {
	var req = &registerNewWikiReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerNewWikiRes
	res, st.Err = registerNewWiki(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterNewWikiHTTP is HTTP handler of RegisterNewWiki service.
func RegisterNewWikiHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerNewWikiReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerNewWikiRes
	res, st.Err = registerNewWiki(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerNewWikiReq struct {
	Language lang.Language
	URI      string     `valid:"text[0:100]"`
	Title    string     `valid:"text[0:100]"`
	Text     string     `valid:"text[0:0]"`
	Pictures [][32]byte `json:",string"`
}

type registerNewWikiRes struct {
	ID [32]byte `json:",string"`
}

func registerNewWiki(st *achaemenid.Stream, req *registerNewWikiReq) (res *registerNewWikiRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	err = checkWikiTitle(st, req.Title, req.Language)
	if err != nil {
		return
	}
	err = checkWikiURI(st, req.URI)
	if err != nil {
		return
	}

	var w = datastore.Wiki{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               uuid.Random32Byte(),
		OrgID:            st.Connection.UserID,

		Language: req.Language,
		URI:      req.URI,
		Title:    req.Title,
		Text:     req.Text,
		Pictures: req.Pictures,
		Status:   datastore.WikiStatusRegister,
	}
	err = w.SaveNew()
	if err != nil {
		return
	}

	res = &registerNewWikiRes{
		ID: w.ID,
	}

	return
}

func (req *registerNewWikiReq) validator() (err *er.Error) {
	// Title must not include ':'(use in URI)
	err = validators.ValidateText(req.Title, 0, 100)
	err = validators.ValidateText(req.URI, 0, 100)
	err = validators.ValidateText(req.Text, 0, 0)
	return
}

func checkWikiTitle(st *achaemenid.Stream, title string, lang lang.Language) (err *er.Error) {
	var findWikiByTitleReq = findWikiByTitleReq{
		Title:  title,
		Offset: 18446744073709551615,
		Limit:  1,
	}
	var findWikiByTitleRes *findWikiByTitleRes
	findWikiByTitleRes, err = findWikiByTitle(st, &findWikiByTitleReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return
	}

	var getWikiByIDReq = getWikiByIDReq{
		ID:       findWikiByTitleRes.IDs[0],
		Language: lang,
	}
	var getWikiByIDRes *getWikiByIDRes
	getWikiByIDRes, err = getWikiByID(st, &getWikiByIDReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		// TODO::: how it is possible???
		return nil
	}
	if err != nil {
		return
	}
	if getWikiByIDRes.Title == title {
		return ErrWikiTitleRegistered
	}
	return
}

func checkWikiURI(st *achaemenid.Stream, uri string) (err *er.Error) {
	var findWikiByURIReq = findWikiByURIReq{
		URI:    uri,
		Offset: 18446744073709551615,
		Limit:  1,
	}
	var findWikiByURIRes *findWikiByURIRes
	findWikiByURIRes, err = findWikiByURI(st, &findWikiByURIReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return
	}

	var getWikiByIDReq = getWikiByIDReq{
		ID: findWikiByURIRes.IDs[0],
	}
	var getWikiByIDRes *getWikiByIDRes
	getWikiByIDRes, err = getWikiByID(st, &getWikiByIDReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		// TODO::: how it is possible???
		return nil
	}
	if err != nil {
		return
	}
	if getWikiByIDRes.URI == uri {
		return ErrWikiURIRegistered
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerNewWikiReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Language = lang.Language(syllab.GetUInt32(buf, 0))
	req.URI = syllab.UnsafeGetString(buf, 4)
	req.Title = syllab.UnsafeGetString(buf, 12)
	req.Text = syllab.UnsafeGetString(buf, 20)
	req.Pictures = syllab.UnsafeGet32ByteArraySlice(buf, 28)
	return
}

func (req *registerNewWikiReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	syllab.SetUInt32(buf, 0, uint32(req.Language))
	hsi = syllab.SetString(buf, req.URI, 4, hsi)
	hsi = syllab.SetString(buf, req.Title, 12, hsi)
	hsi = syllab.SetString(buf, req.Text, 20, hsi)
	syllab.Set32ByteArrayArray(buf, req.Pictures, 28, hsi)
	return
}

func (req *registerNewWikiReq) syllabStackLen() (ln uint32) {
	return 36
}

func (req *registerNewWikiReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.URI))
	ln += uint32(len(req.Title))
	ln += uint32(len(req.Text))
	ln += uint32(len(req.Pictures) * 32)
	return
}

func (req *registerNewWikiReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerNewWikiReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'L':
			decoder.SetFounded()
			decoder.Offset(10)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Language = lang.Language(num)
		case 'U':
			decoder.SetFounded()
			decoder.Offset(6)
			req.URI = decoder.DecodeString()
		case 'T':
			switch decoder.Buf[1] {
			case 'i':
				decoder.SetFounded()
				decoder.Offset(8)
				req.Title = decoder.DecodeString()
			case 'e':
				decoder.SetFounded()
				decoder.Offset(7)
				req.Text = decoder.DecodeString()
			}
		case 'P':
			decoder.SetFounded()
			decoder.Offset(11)
			req.Pictures, err = decoder.Decode32ByteArraySliceAsBase64()
			if err != nil {
				return
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *registerNewWikiReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Language":`)
	encoder.EncodeUInt64(uint64(req.Language))

	encoder.EncodeString(`,"URI":"`)
	encoder.EncodeString(req.URI)

	encoder.EncodeString(`","Title":"`)
	encoder.EncodeString(req.Title)

	encoder.EncodeString(`","Text":"`)
	encoder.EncodeString(req.Text)

	encoder.EncodeString(`","Pictures":[`)
	encoder.Encode32ByteArraySliceAsBase64(req.Pictures)

	encoder.EncodeString(`]}`)

	return encoder.Buf
}

func (req *registerNewWikiReq) jsonLen() (ln int) {
	ln = len(req.URI) + len(req.Title) + len(req.Text)
	ln += len(req.Pictures) * 46
	ln += 75
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerNewWikiRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerNewWikiRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerNewWikiRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerNewWikiRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerNewWikiRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerNewWikiRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
			if err != nil {
				return
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *registerNewWikiRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerNewWikiRes) jsonLen() (ln int) {
	ln = 52
	return
}
