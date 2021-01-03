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
	sep "../libgo/sdk/sep.ir"
	"../libgo/srpc"
	"../libgo/syllab"
	"../libgo/uuid"
	"../libgo/validators"
)

var registerFinancialTransactionService = achaemenid.Service{
	ID:                3071145687,
	IssueDate:         1606376549,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Financial Transaction",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"FinancialTransaction",
	},

	SRPCHandler: RegisterFinancialTransactionSRPC,
	HTTPHandler: RegisterFinancialTransactionHTTP,
}

// RegisterFinancialTransactionSRPC is sRPC handler of RegisterFinancialTransaction service.
func RegisterFinancialTransactionSRPC(st *achaemenid.Stream) {
	var req = &registerFinancialTransactionReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerFinancialTransactionRes
	res, st.Err = registerFinancialTransaction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterFinancialTransactionHTTP is HTTP handler of RegisterFinancialTransaction service.
func RegisterFinancialTransactionHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerFinancialTransactionReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerFinancialTransactionRes
	res, st.Err = registerFinancialTransaction(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerFinancialTransactionReq struct {
	FromSocietyID uint32
	FromUserID    [32]byte `json:",string"`
	PosID         string   // POS == Point of sale

	Amount      price.Amount
	Description string `valid:"text[0:150]" json:",optional"`

	ToSocietyID uint32
	ToUserID    [32]byte `json:",string"`
}

type registerFinancialTransactionRes struct {
	ID [32]byte `json:",string"`
}

func registerFinancialTransaction(st *achaemenid.Stream, req *registerFinancialTransactionReq) (res *registerFinancialTransactionRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	if req.FromSocietyID != achaemenid.Server.Manifest.SocietyID && req.ToSocietyID != achaemenid.Server.Manifest.SocietyID {
		err = ErrFinancialTransactionBadSociety
		return
	}

	if req.FromUserID != st.Connection.UserID && req.ToUserID != st.Connection.UserID {
		err = ErrFinancialTransactionBadUser
		return
	}

	var ft datastore.FinancialTransaction

	if req.FromSocietyID == achaemenid.Server.Manifest.SocietyID && req.FromSocietyID == req.ToSocietyID {
		if req.PosID != "" {
			switch req.PosID[0] {
			case '0': // offline devices
				switch req.PosID[1] {
				case '0': // sep.ir
					var posSendOrderReq = sep.POSSendOrderReq{
						TerminalID:      req.PosID[2:],
						Amount:          req.Amount.String(0),
						AccountType:     sep.POSAccountTypeSingle,
						TransactionType: sep.POSTransactionPurchase,
						ResNum:          uuid.V4().String(),
					}
					var posSendOrderRes *sep.POSSendOrderRes
					posSendOrderRes, err = sepPOS.POSSendOrder(&posSendOrderReq)
					if err != nil {
						return
					}

					ft = datastore.FinancialTransaction{
						UserID: req.ToUserID,
					}
					err = ft.Lock()
					if err != nil {
						return
					}

					ft = datastore.FinancialTransaction{
						AppInstanceID:         achaemenid.Server.Nodes.LocalNode.InstanceID,
						UserConnectionID:      st.Connection.ID,
						UserID:                req.ToUserID,
						ReferenceType:         datastore.FinancialTransactionPOSTransfer,
						PreviousTransactionID: ft.RecordID,
						Amount:                req.Amount,
						Balance:               ft.Balance + req.Amount,
					}
					copy(ft.ReferenceID[:], posSendOrderRes.TraceNumber)
					copy(ft.ReferenceID[len(posSendOrderRes.TraceNumber):], "/")
					copy(ft.ReferenceID[len(posSendOrderRes.TraceNumber)+1:], posSendOrderRes.RRN)
					err = ft.UnLock()
					if err != nil {
						return
					}
				}
			case '1': // web interface
				// TODO:::
				err = ErrNotImplemented
				return
			default:
				err = price.ErrInvalidTerminalID
				return
			}
		} else {
			ft = datastore.FinancialTransaction{
				UserID: req.FromUserID,
			}
			err = ft.Lock()
			if err != nil {
				return
			}

			if ft.Balance < req.Amount {
				err = ErrFinancialTransactionBalance
				return
			}

			ft = datastore.FinancialTransaction{
				AppInstanceID:         achaemenid.Server.Nodes.LocalNode.InstanceID,
				UserConnectionID:      st.Connection.ID,
				UserID:                req.FromUserID,
				ReferenceType:         datastore.FinancialTransactionDonate,
				PreviousTransactionID: ft.RecordID,
				Amount:                -req.Amount,
				Balance:               ft.Balance - req.Amount,
			}
			err = ft.UnLock()
			if err != nil {
				return
			}
			res = &registerFinancialTransactionRes{
				ID: ft.RecordID,
			}

			ft = datastore.FinancialTransaction{
				UserID: req.ToUserID,
			}
			err = ft.Lock()
			if err != nil {
				// TODO::: reverse fromUserID balance
				return
			}

			ft = datastore.FinancialTransaction{
				AppInstanceID:         achaemenid.Server.Nodes.LocalNode.InstanceID,
				UserConnectionID:      st.Connection.ID,
				UserID:                req.ToUserID,
				ReferenceType:         datastore.FinancialTransactionDonate,
				PreviousTransactionID: ft.RecordID,
				Amount:                req.Amount,
				Balance:               ft.Balance + req.Amount,
			}
			err = ft.UnLock()
			return
		}
	} else if req.FromSocietyID != achaemenid.Server.Manifest.SocietyID {
		// TODO::: send request to other society
		err = ErrNotImplemented
		return
	} else if req.ToSocietyID != achaemenid.Server.Manifest.SocietyID {
		// TODO::: send request to other society
		err = ErrNotImplemented
		return
	}

	res = &registerFinancialTransactionRes{
		ID: ft.RecordID,
	}

	return
}

func (req *registerFinancialTransactionReq) validator() (err *er.Error) {
	err = validators.ValidateText(req.Description, 0, 150)
	return
}

/*
	Request Encoders & Decoders
*/

func (req *registerFinancialTransactionReq) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < req.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	req.FromSocietyID = syllab.GetUInt32(buf, 0)
	copy(req.FromUserID[:], buf[4:])
	req.PosID = syllab.UnsafeGetString(buf, 36)
	req.Amount = price.Amount(syllab.GetInt64(buf, 44))
	req.Description = syllab.UnsafeGetString(buf, 52)
	req.ToSocietyID = syllab.GetUInt32(buf, 60)
	copy(req.ToUserID[:], buf[64:])
	return
}

func (req *registerFinancialTransactionReq) syllabEncoder(buf []byte) {
	var hsi uint32 = req.syllabStackLen() // Heap start index || Stack size!

	syllab.SetUInt32(buf, 0, req.FromSocietyID)
	copy(buf[4:], req.FromUserID[:])
	hsi = syllab.SetString(buf, req.PosID, 36, hsi)
	syllab.SetInt64(buf, 44, int64(req.Amount))
	hsi = syllab.SetString(buf, req.Description, 52, hsi)
	syllab.SetUInt32(buf, 60, req.ToSocietyID)
	copy(buf[64:], req.ToUserID[:])
	return
}

func (req *registerFinancialTransactionReq) syllabStackLen() (ln uint32) {
	return 96
}

func (req *registerFinancialTransactionReq) syllabHeapLen() (ln uint32) {
	ln += uint32(len(req.PosID))
	ln += uint32(len(req.Description))
	return
}

func (req *registerFinancialTransactionReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerFinancialTransactionReq) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "FromSocietyID":
			req.FromSocietyID, err = decoder.DecodeUInt32()
		case "FromUserID":
			err = decoder.DecodeByteArrayAsBase64(req.FromUserID[:])
		case "PosID":
			req.PosID, err = decoder.DecodeString()
		case "Amount":
			var num int64
			num, err = decoder.DecodeInt64()
			req.Amount = price.Amount(num)
		case "Description":
			req.Description, err = decoder.DecodeString()
		case "ToSocietyID":
			req.ToSocietyID, err = decoder.DecodeUInt32()
		case "ToUserID":
			err = decoder.DecodeByteArrayAsBase64(req.ToUserID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (req *registerFinancialTransactionReq) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, req.jsonLen()),
	}

	encoder.EncodeString(`{"FromSocietyID":`)
	encoder.EncodeUInt32(req.FromSocietyID)

	encoder.EncodeString(`,"FromUserID":"`)
	encoder.EncodeByteSliceAsBase64(req.FromUserID[:])

	encoder.EncodeString(`","PosID":"`)
	encoder.EncodeString(req.PosID)

	encoder.EncodeString(`","Amount":`)
	encoder.EncodeInt64(int64(req.Amount))

	encoder.EncodeString(`,"Description":"`)
	encoder.EncodeString(req.Description)

	encoder.EncodeString(`","ToSocietyID":`)
	encoder.EncodeUInt32(req.ToSocietyID)

	encoder.EncodeString(`,"ToUserID":"`)
	encoder.EncodeByteSliceAsBase64(req.ToUserID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (req *registerFinancialTransactionReq) jsonLen() (ln int) {
	ln = len(req.PosID) + len(req.Description)
	ln += 247
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerFinancialTransactionRes) syllabDecoder(buf []byte) (err *er.Error) {
	if uint32(len(buf)) < res.syllabStackLen() {
		err = syllab.ErrSyllabDecodeSmallSlice
		return
	}

	copy(res.ID[:], buf[0:])
	return
}

func (res *registerFinancialTransactionRes) syllabEncoder(buf []byte) {
	copy(buf[0:], res.ID[:])
	return
}

func (res *registerFinancialTransactionRes) syllabStackLen() (ln uint32) {
	return 32
}

func (res *registerFinancialTransactionRes) syllabHeapLen() (ln uint32) {
	return
}

func (res *registerFinancialTransactionRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerFinancialTransactionRes) jsonDecoder(buf []byte) (err *er.Error) {
	var decoder = json.DecoderUnsafeMinifed{
		Buf: buf,
	}
	for err == nil {
		var keyName = decoder.DecodeKey()
		switch keyName {
		case "ID":
			err = decoder.DecodeByteArrayAsBase64(res.ID[:])
		default:
			err = decoder.NotFoundKeyStrict()
		}

		if len(decoder.Buf) < 3 {
			return
		}
	}
	return
}

func (res *registerFinancialTransactionRes) jsonEncoder() (buf []byte) {
	var encoder = json.Encoder{
		Buf: make([]byte, 0, res.jsonLen()),
	}

	encoder.EncodeString(`{"ID":"`)
	encoder.EncodeByteSliceAsBase64(res.ID[:])

	encoder.EncodeString(`"}`)
	return encoder.Buf
}

func (res *registerFinancialTransactionRes) jsonLen() (ln int) {
	ln = 52
	return
}
