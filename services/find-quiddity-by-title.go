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
	"../libgo/matn"
	"../libgo/srpc"
	"../libgo/syllab"
)

var findQuiddityByTitleService = achaemenid.Service{
	ID:                1840032628,
	IssueDate:         1605026748,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Quiddity By Title",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "Find Quiddity IDs by given title",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: FindQuiddityByTitleSRPC,
	HTTPHandler: FindQuiddityByTitleHTTP,
}

// FindQuiddityByTitleSRPC is sRPC handler of FindQuiddityByTitle service.
func FindQuiddityByTitleSRPC(st *achaemenid.Stream) {
	var req = &findQuiddityByTitleReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *matn.IndexTextFindRes
	res, st.Err = findQuiddityByTitle(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.SyllabLen()+4)
	res.SyllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindQuiddityByTitleHTTP is HTTP handler of FindQuiddityByTitle service.
func FindQuiddityByTitleHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findQuiddityByTitleReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *matn.IndexTextFindRes
	res, st.Err = findQuiddityByTitle(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.JSONEncoder()
}

type findQuiddityByTitleReq struct {
	Title      string
	PageNumber uint64
}

func findQuiddityByTitle(st *achaemenid.Stream, req *findQuiddityByTitleReq) (res *matn.IndexTextFindRes, err *er.Error) {
	var q = datastore.Quiddity{
		Title: req.Title,
	}
	res, err = q.FindIDsByTitle(req.PageNumber)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findQuiddityByTitleReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Title = syllab.UnsafeGetString(buf, 0)
	req.PageNumber = syllab.GetUInt64(buf, 8)
	return
}

func (req *findQuiddityByTitleReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	hsi = syllab.SetString(buf, req.Title, 0, hsi)
	syllab.SetUInt64(buf, 8, req.PageNumber)
	return
}

func (req *findQuiddityByTitleReq) syllabStackLen() (ln uint32) {
	return 16
}

func (req *findQuiddityByTitleReq) syllabHeapLen() (ln uint32) {
	ln = uint32(len(req.Title))
	return
}

func (req *findQuiddityByTitleReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findQuiddityByTitleReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "Title":
			req.Title, err = decoder.DecodeString()
		case "PageNumber":
			req.PageNumber, err = decoder.DecodeUInt64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *findQuiddityByTitleReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Title":"`)
	encoder.EncodeString(req.Title)

	encoder.EncodeString(`","PageNumber":`)
	encoder.EncodeUInt64(req.PageNumber)

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *findQuiddityByTitleReq) jsonLen() (ln int) {
	ln = len(req.Title)
	ln += 50
	return
}
