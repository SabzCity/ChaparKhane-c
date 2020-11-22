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

var getWikiByIDService = achaemenid.Service{
	ID:                535138582,
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
		lang.EnglishLanguage: "Get Wiki By ID",
	},
	Description: map[lang.Language]string{
		lang.EnglishLanguage: "",
	},
	TAGS: []string{
		"Wiki",
	},

	SRPCHandler: GetWikiByIDSRPC,
	HTTPHandler: GetWikiByIDHTTP,
}

// GetWikiByIDSRPC is sRPC handler of GetWikiByID service.
func GetWikiByIDSRPC(st *achaemenid.Stream) {
	var req = &getWikiByIDReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getWikiByIDRes
	res, st.Err = getWikiByID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetWikiByIDHTTP is HTTP handler of GetWikiByID service.
func GetWikiByIDHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getWikiByIDReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getWikiByIDRes
	res, st.Err = getWikiByID(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getWikiByIDReq struct {
	ID       [32]byte `json:",string"`
	Language lang.Language
}

type getWikiByIDRes struct {
	WriteTime etime.Time

	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	OrgID            [32]byte `json:",string"`

	URI      string
	Title    string
	Text     string
	Pictures [][32]byte `json:",string"`
	Status   datastore.WikiStatus
}

func getWikiByID(st *achaemenid.Stream, req *getWikiByIDReq) (res *getWikiByIDRes, err *er.Error) {
	var w = datastore.Wiki{
		ID:       req.ID,
		Language: req.Language,
	}
	err = w.GetLastByIDLang()
	if err != nil {
		return
	}

	res = &getWikiByIDRes{
		WriteTime: w.WriteTime,

		AppInstanceID:    w.AppInstanceID,
		UserConnectionID: w.UserConnectionID,
		OrgID:            w.OrgID,

		URI:      w.URI,
		Title:    w.Title,
		Text:     w.Text,
		Pictures: w.Pictures,
		Status:   w.Status,
	}

	return
}

/*
	Request Encoders & Decoders
*/

func (req *getWikiByIDReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	req.Language = lang.Language(syllab.GetUInt32(buf, 32))
	return
}

func (req *getWikiByIDReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	syllab.SetUInt32(buf, 32, uint32(req.Language))
	return
}

func (req *getWikiByIDReq) syllabStackLen() (ln uint32) {
	return 36
}

func (req *getWikiByIDReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getWikiByIDReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getWikiByIDReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'I':
			decoder.SetFounded()
			decoder.Offset(5)
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
			if err != nil {
				return
			}
		case 'L':
			decoder.SetFounded()
			decoder.Offset(10)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			req.Language = lang.Language(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (req *getWikiByIDReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`","Language":`)
	encoder.EncodeUInt64(uint64(req.Language))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (req *getWikiByIDReq) jsonLen() (ln int) {
	ln = 84
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getWikiByIDRes) syllabDecoder(buf []byte) (err *er.Error) {
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
	res.Text = syllab.UnsafeGetString(buf, 120)
	res.Pictures = syllab.UnsafeGet32ByteArraySlice(buf, 128)
	res.Status = datastore.WikiStatus(syllab.GetUInt8(buf, 136))
	return
}

func (res *getWikiByIDRes) syllabEncoder(buf []byte) {
	var hsi uint32 = res.syllabStackLen() // Heap start index || Stack size!

	syllab.SetInt64(buf, 0, int64(res.WriteTime))
	copy(buf[8:], res.AppInstanceID[:])
	copy(buf[40:], res.UserConnectionID[:])
	copy(buf[72:], res.OrgID[:])
	hsi = syllab.SetString(buf, res.URI, 104, hsi)
	hsi = syllab.SetString(buf, res.Title, 112, hsi)
	hsi = syllab.SetString(buf, res.Text, 120, hsi)
	syllab.Set32ByteArrayArray(buf, res.Pictures, 128, hsi)
	syllab.SetUInt8(buf, 136, uint8(res.Status))
	return
}

func (res *getWikiByIDRes) syllabStackLen() (ln uint32) {
	return 137
}

func (res *getWikiByIDRes) syllabHeapLen() (ln uint32) {
	ln += uint32(len(res.URI))
	ln += uint32(len(res.Title))
	ln += uint32(len(res.Text))
	ln += uint32(len(res.Pictures) * 32)
	return
}

func (res *getWikiByIDRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getWikiByIDRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for len(decoder.Buf) > 2 {
		decoder.Offset(2)
		switch decoder.Buf[0] {
		case 'W':
			decoder.SetFounded()
			decoder.Offset(11)
			var num int64
			num, err = decoder.DecodeInt64()
			if err != nil {
				return
			}
			res.WriteTime = etime.Time(num)
		case 'A':
			decoder.SetFounded()
			decoder.Offset(16)
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
			if err != nil {
				return
			}
		case 'U':
			switch decoder.Buf[1] {
			case 's':
				decoder.SetFounded()
				decoder.Offset(19)
				err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
				if err != nil {
					return
				}
			case 'R':
				decoder.SetFounded()
				decoder.Offset(6)
				res.URI = decoder.DecodeString()
			}
		case 'O':
			decoder.SetFounded()
			decoder.Offset(8)
			err = decoder.DecodeByteArrayAsBase64(res.OrgID[:])
			if err != nil {
				return
			}
		case 'T':
			switch decoder.Buf[1] {
			case 'i':
				decoder.SetFounded()
				decoder.Offset(8)
				res.Title = decoder.DecodeString()
			case 'e':
				decoder.SetFounded()
				decoder.Offset(7)
				res.Text = decoder.DecodeString()
			}
		case 'P':
			decoder.SetFounded()
			decoder.Offset(11)
			res.Pictures, err = decoder.Decode32ByteArraySliceAsBase64()
			if err != nil {
				return
			}
		case 'S':
			decoder.SetFounded()
			decoder.Offset(8)
			var num uint64
			num, err = decoder.DecodeUInt64()
			if err != nil {
				return
			}
			res.Status = datastore.WikiStatus(num)
		}

		err = decoder.IterationCheck()
		if err != nil {
			return
		}
	}
	return
}

func (res *getWikiByIDRes) jsonEncoder() (buf []byte) {
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

	encoder.EncodeString(`","Text":"`)
	encoder.EncodeString(res.Text)

	encoder.EncodeString(`","Pictures":[`)
	encoder.Encode32ByteArraySliceAsBase64(res.Pictures)

	encoder.EncodeString(`],"Status":`)
	encoder.EncodeUInt8(uint8(res.Status))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getWikiByIDRes) jsonLen() (ln int) {
	ln = len(res.URI) + len(res.Title) + len(res.Text)
	ln += len(res.Pictures) * 46
	ln += 287
	return
}
