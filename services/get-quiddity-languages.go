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

var getQuiddityLanguagesService = achaemenid.Service{
	ID:                3158400636,
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
		lang.LanguageEnglish: "Get Quiddity Languages",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: GetQuiddityLanguagesSRPC,
	HTTPHandler: GetQuiddityLanguagesHTTP,
}

// GetQuiddityLanguagesSRPC is sRPC handler of GetQuiddityLanguages service.
func GetQuiddityLanguagesSRPC(st *achaemenid.Stream) {
	var req = &getQuiddityLanguagesReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getQuiddityLanguagesRes
	res, st.Err = getQuiddityLanguages(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetQuiddityLanguagesHTTP is HTTP handler of GetQuiddityLanguages service.
func GetQuiddityLanguagesHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getQuiddityLanguagesReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getQuiddityLanguagesRes
	res, st.Err = getQuiddityLanguages(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getQuiddityLanguagesReq struct {
	ID [32]byte `json:",string"`
}

type getQuiddityLanguagesRes struct {
	Languages []lang.Language
}

func getQuiddityLanguages(st *achaemenid.Stream, req *getQuiddityLanguagesReq) (res *getQuiddityLanguagesRes, err *er.Error) {
	var w = datastore.Quiddity{
		ID: req.ID,
	}
	res = &getQuiddityLanguagesRes{}
	res.Languages, err = w.FindLanguagesByID(0, 100)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getQuiddityLanguagesReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getQuiddityLanguagesReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	return
}

func (req *getQuiddityLanguagesReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getQuiddityLanguagesReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getQuiddityLanguagesReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getQuiddityLanguagesReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *getQuiddityLanguagesReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getQuiddityLanguagesReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getQuiddityLanguagesRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	var add uint32 = syllab.GetUInt32(buf, 0)
	var ln uint32 = syllab.GetUInt32(buf, 0+4)
	res.Languages = lang.UnsafeByteSliceToLanguagesSlice(buf[add : add+(ln*4)])
	return
}

func (res *getQuiddityLanguagesRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	var ln = uint32(len(res.Languages))
	syllab.SetUInt32(buf, 0, hsi)
	syllab.SetUInt32(buf, 0+4, ln)
	copy(buf[hsi:], lang.UnsafeLanguagesSliceToByteSlice(res.Languages))
	// hsi = hsi + (ln * 4)
	return
}

func (res *getQuiddityLanguagesRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getQuiddityLanguagesRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Languages) * 4)
	return
}

func (res *getQuiddityLanguagesRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getQuiddityLanguagesRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "Languages":
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

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getQuiddityLanguagesRes) jsonEncoder() (buf []byte) {
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

func (res *getQuiddityLanguagesRes) jsonLen() (ln int) {
	ln = (len(res.Languages) * 11)
	ln += 16
	return
}
