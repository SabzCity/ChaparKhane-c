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

var registerQuiddityNewLanguageService = achaemenid.Service{
	ID:                2771579507,
	IssueDate:         1605102782,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeOrg,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Quiddity New Language",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: RegisterQuiddityNewLanguageSRPC,
	HTTPHandler: RegisterQuiddityNewLanguageHTTP,
}

// RegisterQuiddityNewLanguageSRPC is sRPC handler of RegisterQuiddityNewLanguage service.
func RegisterQuiddityNewLanguageSRPC(st *achaemenid.Stream) {
	var req = &registerQuiddityNewLanguageReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = registerQuiddityNewLanguage(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// RegisterQuiddityNewLanguageHTTP is HTTP handler of RegisterQuiddityNewLanguage service.
func RegisterQuiddityNewLanguageHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerQuiddityNewLanguageReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = registerQuiddityNewLanguage(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type registerQuiddityNewLanguageReq struct {
	ID       [32]byte `json:",string"`
	Language lang.Language
	URI      string `valid:"text[0:100]"`
	Title    string `valid:"text[0:100]"`
}

func registerQuiddityNewLanguage(st *achaemenid.Stream, req *registerQuiddityNewLanguageReq) (err *er.Error) {
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
		ID: req.ID,
	}
	var languages []lang.Language
	languages, err = q.FindLanguagesByID(0, 1)
	if err != nil {
		return
	}
	q.Language = languages[0]
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
	q.ListLanguageForID()
	q.IndexIDForTitle()

	return
}

func (req *registerQuiddityNewLanguageReq) validator() (err *er.Error) {
	// Title must not include ':'(use in URI)
	err = validators.ValidateText(req.Title, 0, 100)
	err = validators.ValidateText(req.URI, 0, 100)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerQuiddityNewLanguageReq) syllabDecoder(buf []byte) (err *er.Error) {
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

func (req *registerQuiddityNewLanguageReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], req.ID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))
	hsi = syllab.SetString(buf, req.URI, 36, hsi)
	hsi = syllab.SetString(buf, req.Title, 44, hsi)
	return
}

func (req *registerQuiddityNewLanguageReq) syllabStackLen() (ln uint32) {
	return 52
}

func (req *registerQuiddityNewLanguageReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.URI))
	ln += uint32(len(req.Title))
	return
}

func (req *registerQuiddityNewLanguageReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerQuiddityNewLanguageReq) jsonDecoder(buf []byte) (err *er.Error) {
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

func (req *registerQuiddityNewLanguageReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt32(uint32(req.Language))

	encoder.EncodeString(`,"URI":"`)
	encoder.EncodeString(req.URI)

	encoder.EncodeString(`","Title":"`)
	encoder.EncodeString(req.Title)

	encoder.EncodeString(`]}`)

	return encoder.Buf
}

func (req *registerQuiddityNewLanguageReq) jsonLen() (ln int) {
	ln = len(req.URI) + len(req.Title)
	ln += 102
	return
}
