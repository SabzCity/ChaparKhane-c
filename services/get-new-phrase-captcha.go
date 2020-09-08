/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"bytes"
	"encoding/base64"
	"strconv"

	"../libgo/achaemenid"
	"../libgo/captcha"
	"../libgo/http"
	"../libgo/syllab"
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

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.SetValue(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
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
	res = &getNewPhraseCaptchaRes{
		CaptchaID: pc.ID,
		Image:     pc.Image,
	}
	return
}

func (req *getNewPhraseCaptchaReq) validator() (err error) {
	return
}

func (req *getNewPhraseCaptchaReq) syllabDecoder(buf []byte) (err error) {
	if len(buf) < 2 {
		err = syllab.ErrSyllabDecodingFailedSmallSlice
		return
	}

	req.Language = captcha.Language(buf[0])
	req.ImageFormat = captcha.ImageFormat(buf[1])
	return
}

// jsonDecoder decode minifed version as {"Language":0,"ImageFormat":0}
func (req *getNewPhraseCaptchaReq) jsonDecoder(buf []byte) (err error) {
	var end bool
	var comaL, colonL int // Coma & Colon location
	var num uint64
	for !end {
		buf = buf[comaL+2:] // remove >>	'{"' 	&& 		',"'	due to don't need them
		colonL = bytes.IndexByte(buf, ':')
		comaL = bytes.IndexByte(buf, ',') // can check coma location if string type exist e.g. {"Language": "0,1,2","ImageFormat": 0}
		if comaL < 0 {
			// Reach last item and trailing comma not allowed!
			comaL = len(buf) - 1
			end = true
		}
		switch buf[0] { // Just check first letter first!
		case 'L':
			num, err = strconv.ParseUint(string(buf[colonL+1:comaL]), 10, 8)
			if err != nil {
				return
			}
			req.Language = captcha.Language(num)
		case 'I':
			num, err = strconv.ParseUint(string(buf[colonL+1:comaL]), 10, 8)
			if err != nil {
				return
			}
			req.ImageFormat = captcha.ImageFormat(num)
		}
	}
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *getNewPhraseCaptchaRes) syllabEncoder(offset int) (buf []byte) {
	var hsi int = 24                  // Heap start index || Stack size!
	var ln int = hsi + len(res.Image) // len of strings, slices, maps, ...
	buf = make([]byte, ln+offset)
	var b = buf[offset:]

	copy(b[0:], res.CaptchaID[:])
	b[16] = byte(hsi) // Heap start index
	b[17] = byte(hsi >> 8)
	b[18] = byte(hsi >> 16)
	b[19] = byte(hsi >> 24)
	ln = len(res.Image)
	b[20] = byte(ln)
	b[21] = byte(ln >> 8)
	b[22] = byte(ln >> 16)
	b[23] = byte(ln >> 24)
	copy(b[hsi:], res.Image)
	return
}

func (res *getNewPhraseCaptchaRes) jsonEncoder() (buf []byte) {
	var fixedSizeData = 90 // 14+((16*3)+15)+11+...+2
	var imageBase64Len = base64.StdEncoding.EncodedLen(len(res.Image))
	buf = make([]byte, 0, fixedSizeData+imageBase64Len)
	var ln int // temp var to indicate len of anything through logic

	buf = append(buf, `{"CaptchaID":[`...)
	for i := 0; i < 16; i++ {
		buf = strconv.AppendUint(buf, uint64(res.CaptchaID[i]), 10)
		buf = append(buf, ',')
	}
	buf = buf[:len(buf)-1] // remove trailing comma

	buf = append(buf, `],"Image":"`...)
	ln = len(buf)
	buf = buf[:ln+imageBase64Len]
	base64.StdEncoding.Encode(buf[ln:], res.Image)

	buf = append(buf, `"}`...)
	return
}
