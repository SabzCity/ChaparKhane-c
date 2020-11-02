/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/captcha"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getPhraseCaptchaAudioService = achaemenid.Service{
	ID:                3649760686,
	URI:               "", // API services can set like "/apis?3649760686" but it is not efficient, find services by ID.
	IssueDate:         1593085514,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "GetPhraseCaptchaAudio",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: GetPhraseCaptchaAudioSRPC,
	HTTPHandler: GetPhraseCaptchaAudioHTTP,
}

// GetPhraseCaptchaAudioSRPC is sRPC handler of GetPhraseCaptchaAudio service.
func GetPhraseCaptchaAudioSRPC(st *achaemenid.Stream) {
	var req = &getPhraseCaptchaAudioReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getPhraseCaptchaAudioRes
	res, st.Err = getPhraseCaptchaAudio(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetPhraseCaptchaAudioHTTP is HTTP handler of GetPhraseCaptchaAudio service.
func GetPhraseCaptchaAudioHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getPhraseCaptchaAudioReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getPhraseCaptchaAudioRes
	res, st.Err = getPhraseCaptchaAudio(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getPhraseCaptchaAudioReq struct {
	CaptchaID   [16]byte `json:",string"`
	Language    lang.Language
	AudioFormat captcha.AudioFormat
}

type getPhraseCaptchaAudioRes struct {
	Audio []byte `json:",string"`
}

func getPhraseCaptchaAudio(st *achaemenid.Stream, req *getPhraseCaptchaAudioReq) (res *getPhraseCaptchaAudioRes, err *er.Error) {
	var pc *captcha.PhraseCaptcha = phraseCaptchas.GetAudio(req.CaptchaID, req.Language, req.AudioFormat)
	res = &getPhraseCaptchaAudioRes{
		Audio: pc.Audio,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getPhraseCaptchaAudioReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.CaptchaID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt8(buf, 16))
	req.AudioFormat = captcha.AudioFormat(syllab.GetUInt8(buf, 17))
	return
}

func (req *getPhraseCaptchaAudioReq) syllabStackLen() (ln uint32) {
	return 18
}

func (req *getPhraseCaptchaAudioReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'C':
			decoder.SetFounded()
			decoder.Offset(12)
			err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
			if err != nil {
				return
			}
		case 'L':
			decoder.SetFounded()
			decoder.Offset(10)
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			req.Language = lang.Language(num)
		case 'A':
			decoder.SetFounded()
			decoder.Offset(13)
			var num uint8
			num, err = decoder.DecodeUInt8()
			if err != nil {
				return
			}
			req.AudioFormat = captcha.AudioFormat(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getPhraseCaptchaAudioRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.Audio)
}

func (res *getPhraseCaptchaAudioRes) syllabStackLen() (ln uint32) {
	return 8
}

func (res *getPhraseCaptchaAudioRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.Audio))
	return
}

func (res *getPhraseCaptchaAudioRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getPhraseCaptchaAudioRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"Audio":"`)
	encoder.EncodeByteSliceAsBase64(res.Audio)

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *getPhraseCaptchaAudioRes) jsonLen() (ln int) {
	ln += (len(res.Audio)*8 + 5) / 6
	ln += 12
	return
}
