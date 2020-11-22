/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/validators"
)

var updateWikiService = achaemenid.Service{
	ID:                2878200942,
	IssueDate:         1605102728,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDUpdate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Update Wiki",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: UpdateWikiSRPC,
	HTTPHandler: UpdateWikiHTTP,
}

// UpdateWikiSRPC is sRPC handler of UpdateWiki service.
func UpdateWikiSRPC(st *achaemenid.Stream) {
	var req = &updateWikiReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateWiki(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateWikiHTTP is HTTP handler of UpdateWiki service.
func UpdateWikiHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateWikiReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateWiki(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateWikiReq struct {
	ID       [32]byte `json:",string"`
	Language lang.Language
	URI      string     `valid:"text[0:100]"`
	Title    string     `valid:"text[0:100]"`
	Text     string     `valid:"text[0:0]"`
	Pictures [][32]byte `json:",string"`
}

func updateWiki(st *achaemenid.Stream, req *updateWikiReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	var w = datastore.Wiki{
		ID:       req.ID,
		Language: req.Language,
	}
	err = w.GetLastByIDLang()
	if err != nil {
		return
	}

	if w.OrgID != st.Connection.UserID {
		err = authorization.ErrUserNotAllow
		return
	}
	if w.Status == datastore.WikiStatusBlocked {
		err = ErrBlockedByJustice
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

	var oldTitle = w.Title
	var oldURI = w.URI
	w = datastore.Wiki{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               req.ID,
		OrgID:            st.Connection.UserID,

		Language: req.Language,
		URI:      req.URI,
		Title:    req.Title,
		Text:     req.Text,
		Pictures: req.Pictures,
		Status:   datastore.WikiStatusRegister,
	}
	err = w.Set()
	if err != nil {
		return
	}
	w.HashIndexRecordIDForIDLanguage()
	if req.Title != oldTitle && req.Title != "" {
		w.HashIndexIDForTitle()
	}
	if req.URI != oldURI && req.URI != "" {
		w.HashIndexIDForURI()
	}

	return
}

func (req *updateWikiReq) validator() (err *er.Error) {
	// Title must not include ':'(use in URI)
	err = validators.ValidateText(req.Title, 0, 100)
	err = validators.ValidateText(req.URI, 0, 100)
	err = validators.ValidateText(req.Text, 0, 0)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateWikiReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))
	req.URI = syllab.UnsafeGetString(buf, 36)
	req.Title = syllab.UnsafeGetString(buf, 44)
	req.Text = syllab.UnsafeGetString(buf, 52)
	req.Pictures = syllab.UnsafeGet32ByteArraySlice(buf, 60)
	return
}

func (req *updateWikiReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))
	hsi = syllab.SetString(buf, req.URI, 36, hsi)
	hsi = syllab.SetString(buf, req.Title, 44, hsi)
	hsi = syllab.SetString(buf, req.Text, 52, hsi)
	syllab.Set32ByteArrayArray(buf, req.Pictures, 60, hsi)
	return
}

func (req *updateWikiReq) syllabStackLen() (ln uint32) {
	return 68
}

func (req *updateWikiReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.URI))
	ln += uint32(len(req.Title))
	ln += uint32(len(req.Text))
	ln += uint32(len(req.Pictures) * 32)
	return
}

func (req *updateWikiReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateWikiReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
			if err != nil {
				return
			}
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

func (req *updateWikiReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Language":`)
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

func (req *updateWikiReq) jsonLen() (ln int) {
	ln = len(req.URI) + len(req.Title) + len(req.Text)
	ln += len(req.Pictures) * 46
	ln += 126
	return
}
