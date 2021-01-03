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

var updateQuiddityService = achaemenid.Service{
	ID:                4085673992,
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
		lang.LanguageEnglish: "Update Quiddity",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: UpdateQuidditySRPC,
	HTTPHandler: UpdateQuiddityHTTP,
}

// UpdateQuidditySRPC is sRPC handler of UpdateQuiddity service.
func UpdateQuidditySRPC(st *achaemenid.Stream) {
	var req = &updateQuiddityReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateQuiddity(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateQuiddityHTTP is HTTP handler of UpdateQuiddity service.
func UpdateQuiddityHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateQuiddityReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateQuiddity(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateQuiddityReq struct {
	ID       [32]byte `json:",string"`
	Language lang.Language
	URI      string `valid:"text[0:100]"`
	Title    string `valid:"text[0:100]"`
}

func updateQuiddity(st *achaemenid.Stream, req *updateQuiddityReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	var q = datastore.Quiddity{
		ID:       req.ID,
		Language: req.Language,
	}
	err = q.GetLastByIDLang()
	if err != nil {
		return
	}

	if q.OrgID != st.Connection.UserID {
		err = authorization.ErrUserNotAllow
		return
	}
	if q.Status == datastore.QuiddityStatusBlocked {
		err = ErrBlockedByJustice
		return
	}

	err = checkQuiddityURI(st, req.URI)
	if err != nil {
		return
	}

	var oldTitle = q.Title
	var oldURI = q.URI
	q = datastore.Quiddity{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               req.ID,
		OrgID:            st.Connection.UserID,

		Language: req.Language,
		URI:      req.URI,
		Title:    req.Title,
		Status:   datastore.QuiddityStatusRegister,
	}
	err = q.Set()
	if err != nil {
		return
	}
	q.IndexRecordIDForIDLanguage()
	if req.Title != oldTitle && req.Title != "" {
		q.IndexIDForTitle()
	}
	if req.URI != oldURI && req.URI != "" {
		q.IndexIDForURI()
	}

	return
}

func (req *updateQuiddityReq) validator() (err *er.Error) {
	// Title must not include ':'(use in URI)
	err = validators.ValidateText(req.Title, 0, 100)
	err = validators.ValidateText(req.URI, 0, 100)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateQuiddityReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))
	req.URI = syllab.UnsafeGetString(buf, 36)
	req.Title = syllab.UnsafeGetString(buf, 44)
	return
}

func (req *updateQuiddityReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))
	hsi = syllab.SetString(buf, req.URI, 36, hsi)
	hsi = syllab.SetString(buf, req.Title, 44, hsi)
	return
}

func (req *updateQuiddityReq) syllabStackLen() (ln uint32) {
	return 52
}

func (req *updateQuiddityReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.URI))
	ln += uint32(len(req.Title))
	return
}

func (req *updateQuiddityReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateQuiddityReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		case "Language":
			var num uint32
			num, err = decoder.DecodeUInt32()
			req.Language = lang.Language(num)
		case "URI":
			req.URI, err = decoder.DecodeString()
		case "Title":
			req.Title, err = decoder.DecodeString()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *updateQuiddityReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt8(uint8(req.Language))

	encoder.EncodeString(`,"URI":"`)
	encoder.EncodeString(req.URI)

	encoder.EncodeString(`","Title":"`)
	encoder.EncodeString(req.Title)

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *updateQuiddityReq) jsonLen() (ln int) {
	ln = len(req.URI) + len(req.Title)
	ln += 85
	return
}
