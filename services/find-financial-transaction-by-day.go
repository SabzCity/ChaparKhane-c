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
)

var findFinancialTransactionByDayService = achaemenid.Service{
	ID:                1089718716,
	IssueDate:         1606376521,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Find Financial Transaction By Day",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"FinancialTransaction",
	},

	SRPCHandler: FindFinancialTransactionByDaySRPC,
	HTTPHandler: FindFinancialTransactionByDayHTTP,
}

// FindFinancialTransactionByDaySRPC is sRPC handler of FindFinancialTransactionByDay service.
func FindFinancialTransactionByDaySRPC(st *achaemenid.Stream) {
	var req = &findFinancialTransactionByDayReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *findFinancialTransactionByDayRes
	res, st.Err = findFinancialTransactionByDay(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// FindFinancialTransactionByDayHTTP is HTTP handler of FindFinancialTransactionByDay service.
func FindFinancialTransactionByDayHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &findFinancialTransactionByDayReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *findFinancialTransactionByDayRes
	res, st.Err = findFinancialTransactionByDay(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type findFinancialTransactionByDayReq struct {
	WriteTime etime.Time
	Offset    uint64
	Limit     uint64 `valid:"Limit[1:100]"`
}

type findFinancialTransactionByDayRes struct {
	IDs [][32]byte `json:",string"`
}

func findFinancialTransactionByDay(st *achaemenid.Stream, req *findFinancialTransactionByDayReq) (res *findFinancialTransactionByDayRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}
	// Validate data here due to service use internally by other services!
	err = req.validator()
	if err != nil {
		return
	}

	res = &findFinancialTransactionByDayRes{}

	var ft = datastore.FinancialTransaction{
		WriteTime: req.WriteTime,
		UserID:    st.Connection.UserID,
	}
	res.IDs, err = ft.FindRecordIDsByUserIDWriteTime(req.Offset, req.Limit)
	return
}

func (req *findFinancialTransactionByDayReq) validator() (err *er.Error) {
	return
}

/*
	Request Encoders & Decoders
*/

func (req *findFinancialTransactionByDayReq) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *findFinancialTransactionByDayReq) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (req *findFinancialTransactionByDayReq) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (req *findFinancialTransactionByDayReq) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *findFinancialTransactionByDayReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *findFinancialTransactionByDayReq) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

func (req *findFinancialTransactionByDayReq) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(req)
	return
}

func (req *findFinancialTransactionByDayReq) jsonLen() (ln int) {
	return
}

/*
	Response Encoders & Decoders
*/

func (res *findFinancialTransactionByDayRes) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (res *findFinancialTransactionByDayRes) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (res *findFinancialTransactionByDayRes) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (res *findFinancialTransactionByDayRes) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (res *findFinancialTransactionByDayRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *findFinancialTransactionByDayRes) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, res)
	return
}

func (res *findFinancialTransactionByDayRes) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(res)
	return
}

func (res *findFinancialTransactionByDayRes) jsonLen() (ln int) {
	return
}
