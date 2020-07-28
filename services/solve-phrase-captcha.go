/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../libgo/achaemenid"
	"../libgo/captcha"
	"../libgo/http"
	"../libgo/json"
)

var solvePhraseCaptchaService = achaemenid.Service{
	ID:                2251404010,
	URI:               "", // API services can set like "/apis?2251404010" but it is not efficient, find services by ID.
	Name:              "SolvePhraseCaptcha",
	IssueDate:         1593013968,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		`Solve the number captcha by give specific ID and answer
		and use captcha ID for any request need it until captcha expire in 2 minute`,
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: SolvePhraseCaptchaSRPC,
	HTTPHandler: SolvePhraseCaptchaHTTP,
}

// SolvePhraseCaptchaSRPC is sRPC handler of SolvePhraseCaptcha service.
func SolvePhraseCaptchaSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &solvePhraseCaptchaReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *solvePhraseCaptchaRes
	res, st.ReqRes.Err = solvePhraseCaptcha(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// SolvePhraseCaptchaHTTP is HTTP handler of SolvePhraseCaptcha service.
func SolvePhraseCaptchaHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &solvePhraseCaptchaReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *solvePhraseCaptchaRes
	res, st.ReqRes.Err = solvePhraseCaptcha(st, req)
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

type solvePhraseCaptchaReq struct {
	CaptchaID [16]byte
	Answer    string
}

type solvePhraseCaptchaRes struct {
	CaptchaState captcha.State
}

func solvePhraseCaptcha(st *achaemenid.Stream, req *solvePhraseCaptchaReq) (res *solvePhraseCaptchaRes, err error) {
	res = &solvePhraseCaptchaRes{
		CaptchaState: phraseCaptchas.Solve(req.CaptchaID, req.Answer),
	}

	return
}

func (req *solvePhraseCaptchaReq) validator() (err error) {
	return
}

func (req *solvePhraseCaptchaReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *solvePhraseCaptchaReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *solvePhraseCaptchaRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *solvePhraseCaptchaRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
