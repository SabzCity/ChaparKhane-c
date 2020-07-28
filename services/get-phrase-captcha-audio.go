/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/captcha"
	"../libgo/http"
	"../libgo/json"
)

var getPhraseCaptchaAudioService = achaemenid.Service{
	ID:                3649760686,
	URI:               "", // API services can set like "/apis?3649760686" but it is not efficient, find services by ID.
	Name:              "GetPhraseCaptchaAudio",
	IssueDate:         1593085514,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"",
	},
	TAGS:        []string{""},
	SRPCHandler: GetPhraseCaptchaAudioSRPC,
	HTTPHandler: GetPhraseCaptchaAudioHTTP,
}

// GetPhraseCaptchaAudioSRPC is sRPC handler of GetPhraseCaptchaAudio service.
func GetPhraseCaptchaAudioSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &getPhraseCaptchaAudioReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *getPhraseCaptchaAudioRes
	res, st.ReqRes.Err = getPhraseCaptchaAudio(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// GetPhraseCaptchaAudioHTTP is HTTP handler of GetPhraseCaptchaAudio service.
func GetPhraseCaptchaAudioHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getPhraseCaptchaAudioReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getPhraseCaptchaAudioRes
	res, st.ReqRes.Err = getPhraseCaptchaAudio(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.Body, st.ReqRes.Err = res.jsonEncoder()
	// st.ReqRes.Err make occur on just memory full!

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.SetValue(http.HeaderKeyContentType, "application/json")
}

type getPhraseCaptchaAudioReq struct {
	CaptchaID   [16]byte
	Language    captcha.Language
	AudioFormat captcha.AudioFormat
}

type getPhraseCaptchaAudioRes struct {
	Audio []byte
}

func getPhraseCaptchaAudio(st *achaemenid.Stream, req *getPhraseCaptchaAudioReq) (res *getPhraseCaptchaAudioRes, err error) {
	var pc *captcha.PhraseCaptcha = phraseCaptchas.GetAudio(req.CaptchaID, req.Language, req.AudioFormat)

	res = &getPhraseCaptchaAudioRes{
		Audio: pc.Audio,
	}
	// remove created audio, due to we don't need it here anymore
	pc.Audio = nil

	return
}

func (req *getPhraseCaptchaAudioReq) validator() (err error) {
	return
}

func (req *getPhraseCaptchaAudioReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *getPhraseCaptchaAudioReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *getPhraseCaptchaAudioRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *getPhraseCaptchaAudioRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
