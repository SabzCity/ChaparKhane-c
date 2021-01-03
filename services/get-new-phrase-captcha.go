/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/authorization"
	"../libgo/captcha"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getNewPhraseCaptchaService = achaemenid.Service{
	ID:                2701809032,
	IssueDate:         1593013452,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "GetNewPhraseCaptcha",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "returns new phrase captcha challenge that expire in 4 minute",
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: GetNewPhraseCaptchaSRPC,
	HTTPHandler: GetNewPhraseCaptchaHTTP,
}

// GetNewPhraseCaptchaSRPC is sRPC handler of GetNewPhraseCaptcha service.
func GetNewPhraseCaptchaSRPC(st *achaemenid.Stream) {
	var req = &getNewPhraseCaptchaReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res = &getNewPhraseCaptchaRes{}
	res, st.Err = getNewPhraseCaptcha(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetNewPhraseCaptchaHTTP is HTTP handler of GetNewPhraseCaptcha service.
func GetNewPhraseCaptchaHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getNewPhraseCaptchaReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res = &getNewPhraseCaptchaRes{}
	res, st.Err = getNewPhraseCaptcha(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getNewPhraseCaptchaReq struct {
	Language    lang.Language
	ImageFormat captcha.ImageFormat
}

type getNewPhraseCaptchaRes struct {
	CaptchaID [16]byte `json:",string"`
	Image     []byte   `json:",string"`
}

func getNewPhraseCaptcha(st *achaemenid.Stream, req *getNewPhraseCaptchaReq) (res *getNewPhraseCaptchaRes, err *er.Error) {
	var pc *captcha.PhraseCaptcha = phraseCaptchas.NewImage(req.Language, req.ImageFormat)
	res = &getNewPhraseCaptchaRes{
		CaptchaID: pc.ID,
		Image:     pc.Image,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getNewPhraseCaptchaReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.Language = lang.Language(syllab.GetUInt8(buf, 0))
	req.ImageFormat = captcha.ImageFormat(syllab.GetUInt8(buf, 1))
	return
}

func (req *getNewPhraseCaptchaReq) syllabEncoder(buf []byte) {
	syllab.SetUInt32(buf, 0, uint32(req.Language))
	syllab.SetUInt8(buf, 4, uint8(req.ImageFormat))
	return
}

func (req *getNewPhraseCaptchaReq) syllabStackLen() (ln uint32) {
	return 5
}

func (req *getNewPhraseCaptchaReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getNewPhraseCaptchaReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

// jsonDecoder decode minifed version as {"Language":0,"ImageFormat":0}
func (req *getNewPhraseCaptchaReq) jsonDecoder(buf []byte) (err *er.Error) {
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
		case "ImageFormat":
			var num uint8
			num, err = decoder.DecodeUInt8()
			req.ImageFormat = captcha.ImageFormat(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}
func (req *getNewPhraseCaptchaReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"Language":`)
	encoder.EncodeUInt8(uint8(req.Language))

	encoder.EncodeString(`,"ImageFormat":`)
	encoder.EncodeUInt8(uint8(req.ImageFormat))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *getNewPhraseCaptchaReq) jsonLen() (ln int) {
	ln = 34
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getNewPhraseCaptchaRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.CaptchaID[:], buf[0:])
	res.Image = syllab.UnsafeGetByteArray(buf, 16)
	return
}

func (res *getNewPhraseCaptchaRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	copy(buf[0:], res.CaptchaID[:])
	hsi = syllab.SetByteArray(buf, res.Image, 16, hsi)
	return
}

func (res *getNewPhraseCaptchaRes) syllabStackLen() (ln uint32) {
	return 24
}

func (res *getNewPhraseCaptchaRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Image))
	return
}

func (res *getNewPhraseCaptchaRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getNewPhraseCaptchaRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "CaptchaID":
			err = decoder.DecodeByteArrayAsBase64(res.CaptchaID[:])
		case "Image":
			res.Image, err = decoder.DecodeByteSliceAsBase64()
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getNewPhraseCaptchaRes) jsonEncoder() []byte {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"CaptchaID":"`)
	encoder.EncodeByteSliceAsBase64(res.CaptchaID[:])

	encoder.EncodeString(`","Image":"`)
	encoder.EncodeByteSliceAsBase64(res.Image)

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *getNewPhraseCaptchaRes) jsonLen() (ln int) {
	ln = ((len(res.Image)*8 + 5) / 6)
	ln += 49
	return
}
