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
	"../libgo/srpc"
)

var updateProductDcService = achaemenid.Service{
	ID:                892367742,
	URI:               "", // API services can set like "/apis?892367742" but it is not efficient, find services by ID.
	IssueDate:         1608124777,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDNone,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Update Product Dc",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"Product",
	},

	SRPCHandler: UpdateProductDcSRPC,
	HTTPHandler: UpdateProductDcHTTP,
}

// UpdateProductDcSRPC is sRPC handler of UpdateProductDc service.
func UpdateProductDcSRPC(st *achaemenid.Stream) {
	var req = &updateProductDcReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateProductDc(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateProductDcHTTP is HTTP handler of UpdateProductDc service.
func UpdateProductDcHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateProductDcReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateProductDc(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateProductDcReq struct {
	ID               [32]byte `json:",string"`
	DCID             [32]byte `json:",string"`
	SellerID         [32]byte `json:",string"`
	ProductAuctionID [32]byte `json:",string"`
}

func updateProductDc(st *achaemenid.Stream, req *updateProductDcReq) (err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	var p = datastore.Product{
		ID: req.ID,
	}
	p.GetLastByID()

	// TODO::: Service logic: Authenticate & Authorizing request first by service policy.

	// TODO::: Remove me and write some code e.g. save a record and related indexes if you need!

	return
}

/*
	Request Encoders & Decoders
*/

func (req *updateProductDcReq) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *updateProductDcReq) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (req *updateProductDcReq) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (req *updateProductDcReq) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *updateProductDcReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateProductDcReq) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

func (req *updateProductDcReq) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(req)
	return
}

func (req *updateProductDcReq) jsonLen() (ln int) {
	return
}
