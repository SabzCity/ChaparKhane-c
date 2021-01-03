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
	"../libgo/uuid"
)

var registerProductInvoiceService = achaemenid.Service{
	ID:                2959609272,
	URI:               "", // API services can set like "/apis?2959609272" but it is not efficient, find services by ID.
	IssueDate:         1607014006,
	ExpiryDate:        0,
	ExpireInFavorOf:   "", // English name of favor service just to show off!
	ExpireInFavorOfID: 0,
	Status:            achaemenid.ServiceStatePreAlpha,

	Authorization: authorization.Service{
		CRUD:     authorization.CRUDCreate,
		UserType: authorization.UserTypeAll ^ authorization.UserTypeGuest,
	},

	Name: map[lang.Language]string{
		lang.LanguageEnglish: "Register Product Invoice",
	},
	Description: map[lang.Language]string{
		lang.LanguageEnglish: "",
	},
	TAGS: []string{
		"",
	},

	SRPCHandler: RegisterProductInvoiceSRPC,
	HTTPHandler: RegisterProductInvoiceHTTP,
}

// RegisterProductInvoiceSRPC is sRPC handler of RegisterProductInvoice service.
func RegisterProductInvoiceSRPC(st *achaemenid.Stream) {
	var req = &registerProductInvoiceReq{}
	st.Err = req.syllabDecoder(srpc.GetPayload(st.IncomePayload))
	if st.Err != nil {
		return
	}

	var res *registerProductInvoiceRes
	res, st.Err = registerProductInvoice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		return
	}

	st.OutcomePayload = make([]byte, res.syllabLen()+4)
	res.syllabEncoder(srpc.GetPayload(st.OutcomePayload))
}

// RegisterProductInvoiceHTTP is HTTP handler of RegisterProductInvoice service.
func RegisterProductInvoiceHTTP(st *achaemenid.Stream, httpReq *http.Request, httpRes *http.Response) {
	var req = &registerProductInvoiceReq{}
	st.Err = req.jsonDecoder(httpReq.Body)
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	var res *registerProductInvoiceRes
	res, st.Err = registerProductInvoice(st, req)
	// Check if any error occur in bussiness logic
	if st.Err != nil {
		httpRes.SetStatus(http.StatusBadRequestCode, http.StatusBadRequestPhrase)
		return
	}

	httpRes.SetStatus(http.StatusOKCode, http.StatusOKPhrase)
	httpRes.Header.Set(http.HeaderKeyContentType, "application/json")
	httpRes.Body = res.jsonEncoder()
}

type registerProductInvoiceReq struct {
	UserID            [32]byte `json:",string"` // Due to a seller can register invoice for any user
	UserOTP           uint64
	UserTransactionID [32]byte `json:",string"`
	Products          []registerProductInvoiceDetail
}

type registerProductInvoiceRes struct {
	NotRegistred []registerProductInvoiceDetail
}

type registerProductInvoiceDetail struct {
	QuiddityID           [32]byte            `json:",string"`
	getProductPriceRes   *getProductPriceRes //`json:"-" syllab:"-"`
	ProductAuctionID     [32]byte            `json:",string"`
	getProductAuctionRes *getProductAuctionRes
	DistributionCenterID [32]byte `json:",string"`
	Number               uint64
	Status               uint8
}

func registerProductInvoice(st *achaemenid.Stream, req *registerProductInvoiceReq) (res *registerProductInvoiceRes, err *er.Error) {
	err = st.Authorize()
	if err != nil {
		return
	}

	if req.UserID != [32]byte{} && (req.UserOTP == 0 || req.UserTransactionID == [32]byte{}) {
		err = ErrProductInvoiceDelegate
		return
	}

	var (
		sellerID [32]byte

		totalPriceAmount      price.Amount
		totalDCCommission     price.Amount
		totalSellerCommission price.Amount

		ft      datastore.FinancialTransaction
		product datastore.Product

		// notRegisteredPriceAmount price.Amount
	)

	for _, pro := range req.Products {
		var getProductPriceReq = getProductPriceReq{
			QuiddityID: pro.QuiddityID,
		}
		pro.getProductPriceRes, err = getProductPrice(st, &getProductPriceReq)

		var getProductAuctionReq = getProductAuctionReq{
			ID: pro.ProductAuctionID,
		}
		pro.getProductAuctionRes, err = getProductAuction(st, &getProductAuctionReq)
		// TODO::: authorize product auction e.g. allowUserID, ...

		totalPriceAmount += pro.getProductPriceRes.Price - pro.getProductPriceRes.Price.PerMyriad(pro.getProductAuctionRes.Discount)
		totalDCCommission += pro.getProductPriceRes.Price.PerMyriad(pro.getProductAuctionRes.DCCommission)
		totalSellerCommission += pro.getProductPriceRes.Price.PerMyriad(pro.getProductAuctionRes.SellerCommission)
	}

	if req.UserID == [32]byte{} {
		req.UserID = st.Connection.UserID
	} else {
		sellerID = st.Connection.UserID

		if req.UserTransactionID != [32]byte{} {
			ft = datastore.FinancialTransaction{
				RecordID: req.UserTransactionID,
			}
			err = ft.GetByRecordID()
			if err != nil {
				return
			}
			if ft.UserID != req.UserID || ft.Amount < totalPriceAmount {
				err = ErrProductInvoiceDelegate
				return
			}
		} else {
			// TODO::: check given OTP
		}
	}

	ft = datastore.FinancialTransaction{
		UserID: req.UserID,
	}
	err = ft.Lock()
	if err != nil {
		return
	}
	if ft.Balance < totalPriceAmount {
		err = ErrFinancialTransactionBalance
		return
	}
	ft = datastore.FinancialTransaction{
		AppInstanceID: achaemenid.Server.Nodes.LocalNode.InstanceID,
		// UserConnectionID:      st.Connection.ID, can't uncomment this line due to HTTP use connectionID as authentication proccess!
		UserID:                req.UserID,
		ReferenceType:         datastore.FinancialTransactionProductAuctionPrice,
		PreviousTransactionID: ft.RecordID,
		Amount:                -totalPriceAmount,
		Balance:               ft.Balance - totalPriceAmount,
	}
	err = ft.UnLock()
	if err != nil {
		return
	}

	for _, pro := range req.Products {
		product = datastore.Product{
			AppInstanceID: achaemenid.Server.Nodes.LocalNode.InstanceID,
			// UserConnectionID:      st.Connection.ID, can't uncomment this line due to HTTP use connectionID as authentication proccess!
			ID:         uuid.Random32Byte(),
			OwnerID:    req.UserID,
			SellerID:   sellerID,
			QuiddityID: pro.QuiddityID,
			// ProductionID       :
			DCID:             pro.DistributionCenterID,
			ProductAuctionID: pro.ProductAuctionID,
			Status:           datastore.ProductChangeOwner,
		}
		product.SaveNew()
	}

	res = &registerProductInvoiceRes{}

	return
}

// CalculatePrices method set prices by given price data
// func (pa *ProductAuction) CalculatePrices() {
// 	pa.PayablePrice =
// 	pa.DCCommissionPrice = pa.BasePrice.PerMyriad(pa.DCCommission)
// 	pa.SellerCommissionPrice = pa.BasePrice.PerMyriad(pa.SellerCommission)
// }

/*
	Request Encoders & Decoders
*/

func (req *registerProductInvoiceReq) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *registerProductInvoiceReq) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (req *registerProductInvoiceReq) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (req *registerProductInvoiceReq) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (req *registerProductInvoiceReq) syllabLen() (ln int) {
	return int(req.syllabStackLen() + req.syllabHeapLen())
}

func (req *registerProductInvoiceReq) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, req)
	return
}

func (req *registerProductInvoiceReq) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(req)
	return
}

func (req *registerProductInvoiceReq) jsonLen() (ln int) {
	return
}

/*
	Response Encoders & Decoders
*/

func (res *registerProductInvoiceRes) syllabDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (res *registerProductInvoiceRes) syllabEncoder(buf []byte) {
	// TODO::: Use syllab generator to generate needed codes
}

func (res *registerProductInvoiceRes) syllabStackLen() (ln uint32) {
	return 0 // fixed size data + variables data add&&len
}

func (res *registerProductInvoiceRes) syllabHeapLen() (ln uint32) {
	// TODO::: Use syllab generator to generate needed codes
	return
}

func (res *registerProductInvoiceRes) syllabLen() (ln int) {
	return int(res.syllabStackLen() + res.syllabHeapLen())
}

func (res *registerProductInvoiceRes) jsonDecoder(buf []byte) (err *er.Error) {
	// TODO::: Use json generator to have better performance!
	err = json.UnMarshal(buf, res)
	return
}

func (res *registerProductInvoiceRes) jsonEncoder() (buf []byte) {
	// TODO::: Use json generator to have better performance!
	buf, _ = json.Marshal(res)
	return
}

func (res *registerProductInvoiceRes) jsonLen() (ln int) {
	return
}
