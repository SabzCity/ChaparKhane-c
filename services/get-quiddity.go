/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	etime "../libgo/earth-time"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getQuiddityService = achaemenid.Service{
	ID:                2527290225,
	IssueDate:         1605026701,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Quiddity",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Quiddity",
	},

	SRPCHandler: GetQuidditySRPC,
	HTTPHandler: GetQuiddityHTTP,
}

// GetQuidditySRPC is sRPC handler of GetQuiddity service.
func GetQuidditySRPC(st *achaemenid.Stream) {
	var req = &getQuiddityReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getQuiddityRes
	res, st.Err = getQuiddity(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetQuiddityHTTP is HTTP handler of GetQuiddity service.
func GetQuiddityHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getQuiddityReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getQuiddityRes
	res, st.Err = getQuiddity(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getQuiddityReq struct {
	ID       [32]byte `json:",string"`
	Language lang.Language
}

type getQuiddityRes struct {
	WriteTime etime.Time

	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	OrgID            [32]byte `json:",string"`

	URI    string
	Title  string
	Status datastore.QuiddityStatus
}

func getQuiddity(st *achaemenid.Stream, req *getQuiddityReq) (res *getQuiddityRes, err *er.Error) {
	var w = datastore.Quiddity{
		ID:       req.ID,
		Language: req.Language,
	}
	err = w.GetLastByIDLang()
	if err != nil {
		return
	}

	res = &getQuiddityRes{
		WriteTime: w.WriteTime,

		AppInstanceID:    w.AppInstanceID,
		UserConnectionID: w.UserConnectionID,
		OrgID:            w.OrgID,

		URI:    w.URI,
		Title:  w.Title,
		Status: w.Status,
	}

	return
}

/*
	Request Encoders & Decoders
*/

func (req *getQuiddityReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))
	return
}

func (req *getQuiddityReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))
	return
}

func (req *getQuiddityReq) syllabStackLen() (ln uint32) {
	return 36
}

func (req *getQuiddityReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getQuiddityReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getQuiddityReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		case "Language":
			var num uint32
			num, err = decoder.DecodeUInt32()
			req.Language = lang.Language(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *getQuiddityReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt32(uint32(req.Language))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *getQuiddityReq) jsonLen() (ln int) {
	ln = 74
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getQuiddityRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	res.WriteTime = etime.Time(syllab.GetInt64(buf, 0))
	copy(res.AppInstanceID[:], buf[8:])
	copy(res.UserConnectionID[:], buf[40:])
	copy(res.OrgID[:], buf[72:])

	res.URI = syllab.UnsafeGetString(buf, 104)
	res.Title = syllab.UnsafeGetString(buf, 112)
	res.Status = datastore.QuiddityStatus(syllab.GetUInt8(buf, 120))
	return
}

func (res *getQuiddityRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	copy(buf[72:], res.OrgID[:])
	hsi = syllab.SetString(buf, res.URI, 104, hsi)
	hsi = syllab.SetString(buf, res.Title, 112, hsi)
	syllab.SetUInt8(buf, 120, uint8(res.Status))
	return
}

func (res *getQuiddityRes) syllabStackLen() (ln uint32) {
	return 121
}

func (res *getQuiddityRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.URI))
	ln += uint32(len(res.Title))
	return
}

func (res *getQuiddityRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getQuiddityRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "WriteTime":
			var num int64
			num, err = decoder.DecodeInt64()
			res.WriteTime = etime.Time(num)
		case "AppInstanceID":
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
		case "UserConnectionID":
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
		case "OrgID":
			err = decoder.DecodeByteArrayAsBase64(res.OrgID[:])
		case "URI":
			res.URI, err = decoder.DecodeString()
		case "Title":
			res.Title, err = decoder.DecodeString()
		case "Status":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.Status = datastore.QuiddityStatus(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getQuiddityRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"WriteTime":`)
	encoder.EncodeInt64(int64(res.WriteTime))

	encoder.EncodeString(`,"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","OrgID":"`)
	encoder.EncodeByteSliceAsBase64(res.OrgID[:])

	encoder.EncodeString(`","URI":"`)
	encoder.EncodeString(res.URI)

	encoder.EncodeString(`","Title":"`)
	encoder.EncodeString(res.Title)

	encoder.EncodeString(`","Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getQuiddityRes) jsonLen() (ln int) {
	ln = len(res.URI) + len(res.Title)
	ln += 248
	return
}
