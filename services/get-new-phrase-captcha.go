/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"encoding/base64"

	"../libgo/achaemenid"
	"../libgo/captcha"
	"../libgo/http"
	"../libgo/json"
)

var getNewPhraseCaptchaService = achaemenid.Service{
	ID:                2701809032,
	URI:               "", // API services can set like "/apis?2701809032" but it is not efficient, find services by ID.
	Name:              "GetNewPhraseCaptcha",
	IssueDate:         1593013452,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"returns new phrase captcha challenge that expire in 4 minute",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: GetNewPhraseCaptchaSRPC,
	HTTPHandler: GetNewPhraseCaptchaHTTP,
}

// GetNewPhraseCaptchaSRPC is sRPC handler of GetNewPhraseCaptcha service.
func GetNewPhraseCaptchaSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &getNewPhraseCaptchaReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res = &getNewPhraseCaptchaRes{}
	res, st.ReqRes.Err = getNewPhraseCaptcha(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// GetNewPhraseCaptchaHTTP is HTTP handler of GetNewPhraseCaptcha service.
func GetNewPhraseCaptchaHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getNewPhraseCaptchaReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res = &getNewPhraseCaptchaRes{}
	res, st.ReqRes.Err = getNewPhraseCaptcha(st, req)
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

type getNewPhraseCaptchaReq struct {
	Language    captcha.Language
	ImageFormat captcha.ImageFormat
}

type getNewPhraseCaptchaRes struct {
	CaptchaID [16]byte
	Image     []byte
}

func getNewPhraseCaptcha(st *achaemenid.Stream, req *getNewPhraseCaptchaReq) (res *getNewPhraseCaptchaRes, err error) {
	var pc *captcha.PhraseCaptcha = phraseCaptchas.NewImage(req.Language, req.ImageFormat)

	// TODO::: remove base64 in server and do it in client! first attempt failed! need more working on it! json encode problem??
	res = &getNewPhraseCaptchaRes{
		CaptchaID: pc.ID,
		// Image:     pc.Image,
	}
	res.Image = make([]byte, base64.StdEncoding.EncodedLen(len(pc.Image)))
	base64.StdEncoding.Encode(res.Image, pc.Image)

	// remove created image, due to we don't need it here anymore
	pc.Image = nil

	return
}

func (req *getNewPhraseCaptchaReq) validator() (err error) {
	return
}

func (req *getNewPhraseCaptchaReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *getNewPhraseCaptchaReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *getNewPhraseCaptchaRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *getNewPhraseCaptchaRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
