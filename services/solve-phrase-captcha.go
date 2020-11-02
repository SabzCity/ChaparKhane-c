/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var solvePhraseCaptchaService = achaemenid.Service{
	ID:                2251404010,
	URI:               "", // API services can set like "/apis?2251404010" but it is not efficient, find services by ID.
	IssueDate:         1593013968,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Name: map[lang.Language]string{
		lang.EnglishLanguage: "SolvePhraseCaptcha",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: `Solve the number captcha by give specific ID and answer
and use captcha ID for any request need it until captcha expire in 2 minute`,
	},
	TAGS: []string{
		"Authentication",
	},

	SRPCHandler: SolvePhraseCaptchaSRPC,
	HTTPHandler: SolvePhraseCaptchaHTTP,
}

// SolvePhraseCaptchaSRPC is sRPC handler of SolvePhraseCaptcha service.
func SolvePhraseCaptchaSRPC(st *achaemenid.Stream) {
	var req = &solvePhraseCaptchaReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = solvePhraseCaptcha(st, req)
}

// SolvePhraseCaptchaHTTP is HTTP handler of SolvePhraseCaptcha service.
func SolvePhraseCaptchaHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &solvePhraseCaptchaReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = solvePhraseCaptcha(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type solvePhraseCaptchaReq struct {
	CaptchaID [16]byte `json:",string"`
	Answer    string
}

func solvePhraseCaptcha(st *achaemenid.Stream, req *solvePhraseCaptchaReq) (err *er.Error) {
	err = phraseCaptchas.Solve(req.CaptchaID, req.Answer)
	return
}

func (req *solvePhraseCaptchaReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.CaptchaID[:], buf[0:])
	req.Answer = syllab.UnsafeGetString(buf, 16)
	return
}

func (req *solvePhraseCaptchaReq) syllabStackLen() (ln uint32) {
	return 24
}

func (req *solvePhraseCaptchaReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'C':
			decoder.SetFounded()
			decoder.Offset(9 + 3)
			err = decoder.DecodeByteArrayAsBase64(req.CaptchaID[:])
			if err != nil {
				return
			}
		case 'A':
			decoder.SetFounded()
			decoder.Offset(6 + 3)
			req.Answer = decoder.DecodeString()
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}
