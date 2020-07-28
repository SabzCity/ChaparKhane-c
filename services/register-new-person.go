/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/http"
	"../libgo/json"
	"../libgo/uuid"
)

var registerNewPersonService = achaemenid.Service{
	ID:                956555232,
	URI:               "", // API services can set like "/apis?956555232" but it is not efficient, find services by ID.
	Name:              "RegisterNewPerson",
	IssueDate:         1592316187,
	ExpiryDate:        0,
	ExpireInFavorOf:   "",
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,
	Description: []string{
		"register a new user in SabzCity platform.",
	},
	TAGS:        []string{"Authentication"},
	SRPCHandler: RegisterNewPersonSRPC,
	HTTPHandler: RegisterNewPersonHTTP,
}

// RegisterNewPersonSRPC is sRPC handler of RegisterNewPerson service.
func RegisterNewPersonSRPC(s *achaemenid.Server, st *achaemenid.Stream) {
	var req = &registerNewPersonReq{}
	st.ReqRes.Err = req.syllabDecoder(st.Payload[4:])
	if st.ReqRes.Err != nil {
		return
	}

	var res *registerNewPersonRes
	res, st.ReqRes.Err = registerNewPerson(st, req)
	// Check if any error occur in bussiness logic
	if st.ReqRes.Err != nil {
		return
	}

	st.ReqRes.Payload = res.syllabEncoder(4)
}

// RegisterNewPersonHTTP is HTTP handler of RegisterNewPerson service.
func RegisterNewPersonHTTP(s *achaemenid.Server, st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerNewPersonReq{}
	st.ReqRes.Err = req.jsonDecoder(httpReq.Body)
	if st.ReqRes.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerNewPersonRes
	res, st.ReqRes.Err = registerNewPerson(st, req)
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

type registerNewPersonReq struct {
	PhoneNumber  uint64 `valid:"PhoneNumber"`
	PhoneOTP     uint64
	PasswordHash [32]byte
	CaptchaID    [16]byte
}

type registerNewPersonRes struct {
	PersonID [16]byte // UUID of registered user
}

func registerNewPerson(st *achaemenid.Stream, req *registerNewPersonReq) (res *registerNewPersonRes, err error) {
	// TODO::: Authenticate request first by service policy.

	err = st.Authorize()
	if err != nil {
		return
	}

	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	// Prevent DDos attack by do some easy process for user e.g. captcha is not good way!

	var pa = datastore.PersonAuthentication{
		PersonID:         uuid.NewV4(), // Make new User UUID.
		ReferentPersonID: st.Connection.UserID,
		Status:           datastore.PersonAuthenticationNotForceUse2Factor,
		PasswordHash:     req.PasswordHash,
		OTPPattern:       [32]byte{},
		RecoveryCode:     [128]byte{},
		SecurityQuestion: 0,
		SecurityAnswer:   "",
	}

	// Store pa to datastore!
	// err = pa.Set()

	res = &registerNewPersonRes{
		PersonID: pa.PersonID,
	}

	return
}

func (req *registerNewPersonReq) validator() (err error) {
	return
}

func (req *registerNewPersonReq) syllabDecoder(buf []byte) (err error) {
	return
}

func (req *registerNewPersonReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *registerNewPersonRes) syllabEncoder(offset int) (buf []byte) {
	return
}

func (res *registerNewPersonRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
