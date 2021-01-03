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

var updateProductOwnerService = achaemenid.Service{
	ID:                3481673445,
	URI:               "", // API services can set like "/apis?3481673445" but it is not efficient, find services by ID.
	IssueDate:         1608124535,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDNone,
		UserType: authorization.UserTypeAll,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Update Product Owner",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"",
	},

	SRPCHandler: UpdateProductOwnerSRPC,
	HTTPHandler: UpdateProductOwnerHTTP,
}

// UpdateProductOwnerSRPC is sRPC handler of UpdateProductOwner service.
func UpdateProductOwnerSRPC(st *achaemenid.Stream) {
	var req = &updateProductOwnerReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	st.Err = updateProductOwner(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, 4)
}

// UpdateProductOwnerHTTP is HTTP handler of UpdateProductOwner service.
func UpdateProductOwnerHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &updateProductOwnerReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	st.Err = updateProductOwner(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
}

type updateProductOwnerReq struct {
	ID               [32]byte `json:",string"`
	DCID             [32]byte `json:",string"`
	SellerID         [32]byte `json:",string"`
	ProductAuctionID [32]byte `json:",string"`
}

func updateProductOwner(st *achaemenid.Stream, req *updateProductOwnerReq) (err *er.Error) {
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

func (req *updateProductOwnerReq) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *updateProductOwnerReq) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (req *updateProductOwnerReq) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (req *updateProductOwnerReq) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *updateProductOwnerReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *updateProductOwnerReq) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

func (req *updateProductOwnerReq) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(req)
	return
}

func (req *updateProductOwnerReq) jsonLen() (ln int) {
	return
}
