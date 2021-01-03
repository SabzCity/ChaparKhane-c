/* For license and copyright information please see LEGAL file in repository */

package services

import (
	"../datastore"
	"../libgo/achaemenid"
	"../libgo/authorization"
	er "../libgo/error"
	"../libgo/http"
	"../libgo/json"
	lang "../libgo/language"
	"../libgo/price"
	"../libgo/srpc"
	"../libgo/syllab"
)

var getFinancialTransactionService = achaemenid.Service{
	ID:                344701073,
	IssueDate:         1606376463,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeOwner,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Financial Transaction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"FinancialTransaction",
	},

	SRPCHandler: GetFinancialTransactionSRPC,
	HTTPHandler: GetFinancialTransactionHTTP,
}

// GetFinancialTransactionSRPC is sRPC handler of GetFinancialTransaction service.
func GetFinancialTransactionSRPC(st *achaemenid.Stream) {
	var req = &getFinancialTransactionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getFinancialTransactionRes
	res, st.Err = getFinancialTransaction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetFinancialTransactionHTTP is HTTP handler of GetFinancialTransaction service.
func GetFinancialTransactionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getFinancialTransactionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getFinancialTransactionRes
	res, st.Err = getFinancialTransaction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getFinancialTransactionReq struct {
	ID [32]byte `json:",string"`
}

type getFinancialTransactionRes struct {
	AppInstanceID         [32]byte `json:",string"`
	UserConnectionID      [32]byte `json:",string"`
	ReferenceID           [32]byte `json:",string"`
	ReferenceType         datastore.FinancialTransactionType
	PreviousTransactionID [32]byte `json:",string"`
	Amount                price.Amount
	Balance               price.Amount
}

func getFinancialTransaction(st *achaemenid.Stream, req *getFinancialTransactionReq) (res *getFinancialTransactionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var ft = datastore.FinancialTransaction{
		RecordID: req.ID,
	}
	err = ft.GetByRecordID()
	if err != nil {
		return
	}

	if st.Connection.UserID != ft.UserID {
		err = authorization.ErrUserNotOwnRecord
		return
	}

	res = &getFinancialTransactionRes{
		AppInstanceID:         ft.AppInstanceID,
		UserConnectionID:      ft.UserConnectionID,
		ReferenceID:           ft.ReferenceID,
		ReferenceType:         ft.ReferenceType,
		PreviousTransactionID: ft.PreviousTransactionID,
		Amount:                ft.Amount,
		Balance:               ft.Balance,
	}
	return
}

/*
	Request Encoders & Decoders
*/

func (req *getFinancialTransactionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(req.ID[:], buf[0:])
	return
}

func (req *getFinancialTransactionReq) syllabEncoder(buf []byte) {
	copy(buf[0:], req.ID[:])
	return
}

func (req *getFinancialTransactionReq) syllabStackLen() (ln uint32) {
	return 32
}

func (req *getFinancialTransactionReq) syllabHeapLen() (ln uint32) {
	return
}

func (req *getFinancialTransactionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getFinancialTransactionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(req.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *getFinancialTransactionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(req.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *getFinancialTransactionReq) jsonLen() (ln int) {
	ln = 52
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getFinancialTransactionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.AppInstanceID[:], buf[0:])
	copy(res.UserConnectionID[:], buf[32:])
	copy(res.ReferenceID[:], buf[64:])
	res.ReferenceType = datastore.FinancialTransactionType(syllab.GetUInt8(buf, 96))
	copy(res.PreviousTransactionID[:], buf[97:])
	res.Amount = price.Amount(syllab.GetInt64(buf, 129))
	res.Balance = price.Amount(syllab.GetInt64(buf, 137))
	return
}

func (res *getFinancialTransactionRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.AppInstanceID[:])
	copy(buf[32:], res.UserConnectionID[:])
	copy(buf[64:], res.ReferenceID[:])
	syllab.SetUInt8(buf, 96, uint8(res.ReferenceType))
	copy(buf[97:], res.PreviousTransactionID[:])
	syllab.SetInt64(buf, 129, int64(res.Amount))
	syllab.SetInt64(buf, 137, int64(res.Balance))
	return
}

func (res *getFinancialTransactionRes) syllabStackLen() (ln uint32) {
	return 145
}

func (res *getFinancialTransactionRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *getFinancialTransactionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getFinancialTransactionRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "AppInstanceID":
			err = decoder.DecodeByteArrayAsBase64(res.AppInstanceID[:])
		case "UserConnectionID":
			err = decoder.DecodeByteArrayAsBase64(res.UserConnectionID[:])
		case "ReferenceID":
			err = decoder.DecodeByteArrayAsBase64(res.ReferenceID[:])
		case "ReferenceType":
			var num uint8
			num, err = decoder.DecodeUInt8()
			res.ReferenceType = datastore.FinancialTransactionType(num)
		case "PreviousTransactionID":
			err = decoder.DecodeByteArrayAsBase64(res.PreviousTransactionID[:])
		case "Amount":
			var num int64
			num, err = decoder.DecodeInt64()
			res.Amount = price.Amount(num)
		case "Balance":
			var num int64
			num, err = decoder.DecodeInt64()
			res.Balance = price.Amount(num)
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *getFinancialTransactionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"AppInstanceID":"`)
	encoder.EncodeByteSliceAsBase64(res.AppInstanceID[:])

	encoder.EncodeString(`","UserConnectionID":"`)
	encoder.EncodeByteSliceAsBase64(res.UserConnectionID[:])

	encoder.EncodeString(`","ReferenceID":"`)
	encoder.EncodeByteSliceAsBase64(res.ReferenceID[:])

	encoder.EncodeString(`","ReferenceType":`)
	encoder.EncodeUInt8(uint8(res.ReferenceType))

	encoder.EncodeString(`,"PreviousTransactionID":"`)
	encoder.EncodeByteSliceAsBase64(res.PreviousTransactionID[:])

	encoder.EncodeString(`","Amount":`)
	encoder.EncodeInt64(int64(res.Amount))

	encoder.EncodeString(`,"Balance":`)
	encoder.EncodeInt64(int64(res.Balance))

	encoder.EncodeByte('}')
	return encoder.Buf
}

func (res *getFinancialTransactionRes) jsonLen() (ln int) {
	ln = 356
	return
}
