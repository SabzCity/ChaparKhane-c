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

var registerQuiddityService = achaemenid.Service{
	ID:                1804195349,
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
		lang.LanguageEnglish: "Register Quiddity",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: RegisterQuidditySRPC,
	HTTPHandler: RegisterQuiddityHTTP,
}

// RegisterQuidditySRPC is sRPC handler of RegisterQuiddity service.
func RegisterQuidditySRPC(st *achaemenid.Stream) {
	var req = &registerQuiddityReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerQuiddityRes
	res, st.Err = registerQuiddity(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterQuiddityHTTP is HTTP handler of RegisterQuiddity service.
func RegisterQuiddityHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerQuiddityReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerQuiddityRes
	res, st.Err = registerQuiddity(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerQuiddityReq struct {
	Language lang.Language
	URI      string `valid:"text[0:100]"`
	Title    string `valid:"text[0:100]"`
}

type registerQuiddityRes struct {
	ID [32]byte `json:",string"`
}

func registerQuiddity(st *achaemenid.Stream, req *registerQuiddityReq) (res *registerQuiddityRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	err = checkQuiddityURI(st, req.URI)
	if err != nil {
		return
	}

	// err = checkOrgName(st, req.Name)
	// if err != nil {
	// 	return
	// }
	// err = checkOrgDomain(st, req.Domain)
	// if err != nil {
	// 	return
	// }

	var q = datastore.Quiddity{
		AppInstanceID:    achaemenid.Server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		ID:               uuid.Random32Byte(),
		OrgID:            st.Connection.UserID,

		Language: req.Language,
		URI:      req.URI,
		Title:    req.Title,
		Status:   datastore.QuiddityStatusRegister,
	}
	err = q.SaveNew()
	if err != nil {
		return
	}

	res = &registerQuiddityRes{
		ID: q.ID,
	}

	return
}

func (req *registerQuiddityReq) validator() (err *er.Error) {
	// Title must not include ':'(use in URI)
	err = validators.ValidateText(req.Title, 0, 100)
	err = validators.ValidateText(req.URI, 0, 100)
	return
}

func checkQuiddityURI(st *achaemenid.Stream, uri string) (err *er.Error) {
	var findQuiddityByURIReq = findQuiddityByURIReq{
		URI:    uri,
		Offset: 18446744073709551615,
		Limit:  1,
	}
	var findQuiddityByURIRes *findQuiddityByURIRes
	findQuiddityByURIRes, err = findQuiddityByURI(st, &findQuiddityByURIReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return
	}

	var getQuiddityReq = getQuiddityReq{
		ID: findQuiddityByURIRes.IDs[0],
	}
	var getQuiddityRes *getQuiddityRes
	getQuiddityRes, err = getQuiddity(st, &getQuiddityReq)
	if err.Equal(ganjine.ErrRecordNotFound) {
		// TODO::: how it is possible???
		return nil
	}
	if err != nil {
		return
	}
	if getQuiddityRes.URI == uri {
		return ErrQuiddityURIRegistered
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerQuiddityReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Language = lang.Language(syllab.GetUInt32(buf, 0))
	req.URI = syllab.UnsafeGetString(buf, 4)
	req.Title = syllab.UnsafeGetString(buf, 12)
	return
}

func (req *registerQuiddityReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	syllab.SetUInt32(buf, 0, uint32(req.Language))
	hsi = syllab.SetString(buf, req.URI, 4, hsi)
	hsi = syllab.SetString(buf, req.Title, 12, hsi)
	return
}

func (req *registerQuiddityReq) syllabStackLen() (ln uint32) {
	return 20
}

func (req *registerQuiddityReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.URI))
	ln += uint32(len(req.Title))
	return
}

func (req *registerQuiddityReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerQuiddityReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
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

func (req *registerQuiddityReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Language":`)
	encoder.EncodeUInt32(uint32(req.Language))

	encoder.EncodeString(`,"URI":"`)
	encoder.EncodeString(req.URI)

	encoder.EncodeString(`","Title":"`)
	encoder.EncodeString(req.Title)

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *registerQuiddityReq) jsonLen() (ln int) {
	ln = len(req.URI) + len(req.Title)
	ln += 43
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerQuiddityRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerQuiddityRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerQuiddityRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerQuiddityRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerQuiddityRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerQuiddityRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *registerQuiddityRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerQuiddityRes) jsonLen() (ln int) {
	ln = 52
	return
}
