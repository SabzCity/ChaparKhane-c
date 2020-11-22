/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"strconv"

	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getWikiLanguagesService = achaemenid.Service{
	ID:                183294015,
	IssueDate:         1605106959,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "Get Wiki Languages",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: GetWikiLanguagesSRPC,
	HTTPHandler: GetWikiLanguagesHTTP,
}

// GetWikiLanguagesSRPC is sRPC handler of GetWikiLanguages service.
func GetWikiLanguagesSRPC(st *achaemenid.Stream) {
	var req = &getWikiLanguagesReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getWikiLanguagesRes
	res, st.Err = getWikiLanguages(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetWikiLanguagesHTTP is HTTP handler of GetWikiLanguages service.
func GetWikiLanguagesHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getWikiLanguagesReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getWikiLanguagesRes
	res, st.Err = getWikiLanguages(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getWikiLanguagesReq struct {
	ID [32]byte `json:",string"`
}

type getWikiLanguagesRes struct {
	Languages []lang.Language
}

func getWikiLanguages(st *achaemenid.Stream, req *getWikiLanguagesReq) (res *getWikiLanguagesRes, err *er.Error) {
	var w = datastore.Wiki{
		ID: req.ID,
	}
	res = &getWikiLanguagesRes{}
	res.Languages, err = w.GetLanguagesByIDByHashIndex(0, 100)
	return
}

func (req *getWikiLanguagesReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getWikiLanguagesReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getWikiLanguagesReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	return
}

func (req *getWikiLanguagesReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getWikiLanguagesReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getWikiLanguagesReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getWikiLanguagesReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getWikiLanguagesReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getWikiLanguagesReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getWikiLanguagesRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	var add uint32 = syllab.GetUInt32(buf, 0)
	var ln uint32 = syllab.GetUInt32(buf, 0+4)
	res.Languages = lang.UnsafeByteSliceToLanguagesSlice(buf[add : add+(ln*4)])
	return
}

func (res *getWikiLanguagesRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	var ln = uint32(len(res.Languages))
	syllab.SetUInt32(buf, 0, hsi)
	syllab.SetUInt32(buf, 0+4, ln)
	copy(buf[hsi:], lang.UnsafeLanguagesSliceToByteSlice(res.Languages))
	// hsi = hsi + (ln * 4)
	return
}

func (res *getWikiLanguagesRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getWikiLanguagesRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Languages) * 4)
	return
}

func (res *getWikiLanguagesRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getWikiLanguagesRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'L':
			decoder.SetFounded()
			decoder.Offset(12)
			var num uint32
			res.Languages = make([]lang.Language, 0, 8) // TODO::: Is cap efficient enough?
			for !decoder.CheckToken(']') {
				num, err = decoder.DecodeUInt32()
				if err != nil {
					return
				}
				res.Languages = append(res.Languages, lang.Language(num))
				decoder.Offset(1)
			}
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getWikiLanguagesRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"Languages":[`)
	var ln = len(res.Languages)
	for i := 0; i < ln; i++ {
		encoder.Buf = strconv.AppendUint(encoder.Buf, uint64(res.Languages[i]), 10)
		encoder.Buf = append(encoder.Buf, ',')
	}
	encoder.RemoveTrailingComma()

	encoder.EncodeString(`]}`)
	return encoder.Buf
}

func (res *getWikiLanguagesRes) jsonLen() (ln int) {
	ln = (len(res.Languages) * 11)
	ln += 16
	return
}
