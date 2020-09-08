/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"crypto/rand"

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
	PhoneNumber uint64 `valid:"PhoneNumber"`
	PhoneOTP    uint64
	CaptchaID   [16]byte

	PasswordHash  [32]byte
	OTPAdditional int32
}

type registerNewPersonRes struct {
	PersonID    [16]byte // UUID of registered user
	OTPPattern  [32]byte
	SecurityKey [32]byte
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
	err = phraseCaptchas.Check(req.CaptchaID)
	if err != nil {
		return
	}

	var OTPPattern = make([]byte, 32)
	_, err = rand.Read(OTPPattern)
	// Note that err == nil only if we read len(OTPPattern) bytes.
	if err != nil {
		// err =
		return
	}

	var SecurityKey = make([]byte, 32)
	_, err = rand.Read(SecurityKey)
	// Note that err == nil only if we read len(SecurityKey) bytes.
	if err != nil {
		// err =
		return
	}

	var pa = datastore.PersonAuthentication{
		AppInstanceID:    server.Nodes.LocalNode.InstanceID,
		UserConnectionID: st.Connection.ID,
		PersonID:         uuid.NewV4(), // Make new User UUID.
		ReferentPersonID: st.Connection.UserID,
		Status:           datastore.PersonAuthenticationNotForceUse2Factor,
		PasswordHash:     req.PasswordHash,
		OTPAdditional:    req.OTPAdditional,
	}
	copy(pa.OTPPattern[:], OTPPattern[:])
	copy(pa.SecurityKey[:], SecurityKey[:])

	// Store pa to datastore!
	err = pa.Set()
	if err != nil {
		// TODO:::
	}

	// Index desire data
	pa.IndexPersonID()
	pa.IndexRegisterTime()
	if st.Connection.UserType != 0 {
		pa.IndexReferentPersonID()
	}

	res = &registerNewPersonRes{
		PersonID:    pa.PersonID,
		OTPPattern:  pa.OTPPattern,
		SecurityKey: pa.SecurityKey,
	}

	return
}

func (req *registerNewPersonReq) validator() (err error) {
	return
}

func (req *registerNewPersonReq) syllabDecoder(buf []byte) (err error) {
	req.PhoneNumber = uint64(buf[0]) | uint64(buf[1])<<8 | uint64(buf[2])<<16 | uint64(buf[3])<<24 | uint64(buf[4])<<32 | uint64(buf[5])<<40 | uint64(buf[6])<<48 | uint64(buf[7])<<56
	req.PhoneOTP = uint64(buf[8]) | uint64(buf[9])<<8 | uint64(buf[10])<<16 | uint64(buf[11])<<24 | uint64(buf[12])<<32 | uint64(buf[13])<<40 | uint64(buf[14])<<48 | uint64(buf[15])<<56
	copy(req.CaptchaID[:], buf[16:])
	copy(req.PasswordHash[:], buf[32:])
	req.OTPAdditional = int32(buf[64]) | int32(buf[65])<<8 | int32(buf[66])<<16 | int32(buf[67])<<24

	return
}

func (req *registerNewPersonReq) jsonDecoder(buf []byte) (err error) {
	// TODO::: Help to complete json generator package to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

// offset add free space by given number at begging of return slice that almost just use in sRPC protocol! It can be 0!!
func (res *registerNewPersonRes) syllabEncoder(offset int) (buf []byte) {
	var hsi int = 80 // Heap start index || Stack size!
	var ln int = hsi      // len of strings, slices, maps, ...
	buf = make([]byte, ln+offset)
	var b = buf[offset:]

	copy(b[0:], res.PersonID[:])
	copy(b[16:], res.OTPPattern[:])
	copy(b[48:], res.SecurityKey[:])

	return
}

func (res *registerNewPersonRes) jsonEncoder() (buf []byte, err error) {
	// TODO::: Help to complete json generator package to have better performance!
	buf, err = json.Marshal(res)
	return
}
