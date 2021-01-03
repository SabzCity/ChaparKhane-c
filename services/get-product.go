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

var getProductService = achaemenid.Service{
	ID:                1989697652,
	URI:               "", // API services can set like "/apis?1989697652" but it is not efficient, find services by ID.
	IssueDate:         1608124902,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDRead,
		UserType: authorization.UserTypeOwner,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Get Product",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Product",
	},

	SRPCHandler: GetProductSRPC,
	HTTPHandler: GetProductHTTP,
}

// GetProductSRPC is sRPC handler of GetProduct service.
func GetProductSRPC(st *achaemenid.Stream) {
	var req = &getProductReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *getProductRes
	res, st.Err = getProduct(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// GetProductHTTP is HTTP handler of GetProduct service.
func GetProductHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &getProductReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *getProductRes
	res, st.Err = getProduct(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type getProductReq struct {
	ID [32]byte `json:",string"`
}

type getProductRes struct {
	WriteTime etime.Time

	AppInstanceID    [32]byte `json:",string"`
	UserConnectionID [32]byte `json:",string"`
	OwnerID          [32]byte `json:",string"`
	SellerID         [32]byte `json:",string"`
	QuiddityID           [32]byte `json:",string"`
	ProductionID     [32]byte `json:",string"`
	DCID             [32]byte `json:",string"`
	ProductAuctionID [32]byte `json:",string"`
	Status           datastore.ProductStatus
}

func getProduct(st *achaemenid.Stream, req *getProductReq) (res *getProductRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var p = datastore.Product{
		ID: req.ID,
	}
	err = p.GetLastByID()
	if err != nil {
		return
	}

	if st.Connection.UserID != p.OwnerID {
		err = authorization.ErrUserNotAllow
		return
	}

	res = &getProductRes{
		WriteTime: p.WriteTime,

		AppInstanceID:    p.AppInstanceID,
		UserConnectionID: p.UserConnectionID,
		OwnerID:          p.OwnerID,
		SellerID:         p.SellerID,
		QuiddityID:           p.QuiddityID,
		ProductionID:     p.ProductionID,
		DCID:             p.DCID,
		ProductAuctionID: p.ProductAuctionID,
		Status:           p.Status,
	}

	return
}

/*
	Request Encoders & Decoders
*/

func (req *getProductReq) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *getProductReq) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (req *getProductReq) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (req *getProductReq) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *getProductReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *getProductReq) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

func (req *getProductReq) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(req)
	return
}

func (req *getProductReq) jsonLen() (ln int) {
	return
}

/*
	Response Encoders & Decoders
*/

func (res *getProductRes) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (res *getProductRes) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (res *getProductRes) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (res *getProductRes) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (res *getProductRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *getProductRes) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, res)
	return
}

func (res *getProductRes) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(res)
	return
}

func (res *getProductRes) jsonLen() (ln int) {
	return
}
