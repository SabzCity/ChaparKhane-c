/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/errors.js'
import '../libjs/cookie.js'
import '../libjs/price/currency.js'
import '../libjs/math/per-myriad.js'
import '../libjs/math/per-cent.js'
import '../../sdk-js/get-product-auction.js'
import '../../sdk-js/register-default-product-auction.js'
import '../../sdk-js/register-custom-product-auction.js'
import '../../sdk-js/update-product-auction.js'

const productAuctionPage = {
    ID: "product-auction",
    Conditions: {
        id: "",
    },
    State: "",
    Robots: "all",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "assessment",
    Related: ["wiki", "orgs"],
    HTML: (productAuction) => ``,
    CSS: '',
    Templates: {},
}
pages.RegisterPage(productAuctionPage)

productAuctionPage.ConnectedCallback = async function () {
    let productAuction
    if (this.Conditions.id) {
        try {
            // GetProductAuctionReq is the request structure of GetProductAuction()
            const GetProductAuctionReq = {
                "ID": this.Conditions.id,
            }
            productAuction = await GetProductAuction(GetProductAuctionReq)
            productAuction.ID = this.Conditions.id
        } catch (err) {
            productAuction = this.NotFoundPage()
            errors.HandleError(err)
        }
    } else {
        productAuction = this.NotFoundPage()
        errors.HandleError(errors.poolByID[1685872164])
    }

    window.document.body.innerHTML = this.HTML(productAuction)

    if (cookie.GetByName(HTTPCookieNameDelegateUserID)) {
        const toolbarElement = document.getElementById("toolbar")
        toolbarElement.hidden = false
    }
}

productAuctionPage.DisconnectedCallback = function () {
    // TODO::: Warn about not save changes
}

productAuctionPage.NotFoundPage = function () {
    return productAuction = {
        "WriteTime": 0, // int64
        "AppInstanceID": "", // [32]byte that encode||decode as base64
        "UserConnectionID": "", // [32]byte that encode||decode as base64
        "OrgID": "", // [32]byte that encode||decode as base64
        "ID": "",
        "WikiID": "", // [32]byte that encode||decode as base64

        "Currency": 7337, // Persia Derik
        "SuggestPrice": 0, // uint64
        "DistributionCenterCommission": 0, // uint16
        "SellerCommission": 0, // uint16
        "Discount": 0, // uint16
        "PayablePrice": 0, // uint64

        "DistributionCenterID": "", // [32]byte that encode||decode as base64
        "GroupID": "", // [32]byte that encode||decode as base64
        "MinNumBuy": 0, // uint64
        "StockNumber": 0, // uint64
        "LiveUntil": 0, // etime.Time
        "AllowWeekdays": 0, // uint8
        "AllowDayhours": 0, // uint8

        "Description": ``,
        "Type": 0, // uint8
        "Status": 0, // uint8
    }
}

productAuctionPage.calculatePrices = function (element) {
    const suggestPriceInputElement = document.getElementById("suggestPriceInput")
    const discountInputElement = document.getElementById("discountInput")
    const payablePriceInputElement = document.getElementById("payablePriceInput")

    const suggestPrice = suggestPriceInputElement.value
    const discount = math.PerCent.GetAsPerMyriad(discountInputElement.value)
    const payablePrice = payablePriceInputElement.value

    switch (element.id) {
        case "suggestPriceInput":
        case "discountInput":
            payablePriceInputElement.value = Math.round(suggestPrice - math.PerMyriad.Calculate(suggestPrice, discount))
            return
        case "payablePriceInput":
            discountInputElement.value = math.PerCent.GetReverse(payablePrice, suggestPrice)
            return
    }
}

productAuctionPage.EnableDefaultNew = function () {
    const newDefaultProductAuctionElement = document.getElementById("newDefaultProductAuction")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")
    const defaultNewInputsElements = document.getElementsByClassName("defaultNewInputs")

    newDefaultProductAuctionElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false
    for (let element of defaultNewInputsElements) element.hidden = false
}

productAuctionPage.EnableCustomNew = function () {
    const newCustomProductAuctionElement = document.getElementById("newCustomProductAuction")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")
    const customNewInputsElements = document.getElementsByClassName("customNewInputs")

    newCustomProductAuctionElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false
    for (let element of customNewInputsElements) element.hidden = false
}

productAuctionPage.EnableEdit = function () {
    const editProductAuctionElement = document.getElementById("editProductAuction")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")
    const editInputsElements = document.getElementsByClassName("editInputs")

    editProductAuctionElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false
    for (let element of editInputsElements) element.hidden = false

}

productAuctionPage.SaveEdit = async function () {
    const newDefaultProductAuctionElement = document.getElementById("newDefaultProductAuction")
    const newCustomProductAuctionElement = document.getElementById("newCustomProductAuction")
    const editProductAuctionElement = document.getElementById("editProductAuction")

    const wikiIDInputElement = document.getElementById("wikiIDInput")

    const currencyInputElement = document.getElementById("currencyInput")
    const suggestPriceInputElement = document.getElementById("suggestPriceInput")
    const distributionCenterCommissionInputElement = document.getElementById("distributionCenterCommissionInput")
    const sellerCommissionInputElement = document.getElementById("sellerCommissionInput")
    const discountInputElement = document.getElementById("discountInput")

    const distributionCenterIDInputElement = document.getElementById("distributionCenterIDInput")
    const groupIDInputElement = document.getElementById("groupIDInput")
    const minNumBuyInputElement = document.getElementById("minNumBuyInput")
    const stockNumberInputElement = document.getElementById("stockNumberInput")
    const liveUntilInputElement = document.getElementById("liveUntilInput")
    const allowWeekdaysInputElement = document.getElementById("allowWeekdaysInput")
    const allowDayhoursInputElement = document.getElementById("allowDayhoursInput")

    const descriptionInputElement = document.getElementById("descriptionInput")
    const typeInputElement = document.getElementById("typeInput")

    const cur = currency.poolByNativeName[currencyInputElement.value]
    if (!cur) {
        errors.HandleError(errors.poolByID[166114569])
        return
    }

    if (newDefaultProductAuctionElement.disabled) {
        try {
            // RegisterDefaultProductAuctionReq is the request structure of RegisterDefaultProductAuction()
            const RegisterDefaultProductAuctionReq = {
                "WikiID": wikiIDInputElement.value,
                "Language": 0,  // lang.Language that just use to check wiki exist and belong to requested org

                "Currency": cur.iso4217_num,
                "SuggestPrice": Math.round(suggestPriceInputElement.value),
                "DistributionCenterCommission": math.PerCent.GetAsPerMyriad(distributionCenterCommissionInputElement.value),
                "SellerCommission": math.PerCent.GetAsPerMyriad(sellerCommissionInputElement.value),
                "Discount": math.PerCent.GetAsPerMyriad(discountInputElement.value),

                "Description": descriptionInputElement.value,
                "Type": Number(typeInputElement.value),
            }
            const res = await RegisterDefaultProductAuction(RegisterDefaultProductAuctionReq)
            pages.Router({}, "/product-auction?id=" + res.ID)
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else if (newCustomProductAuctionElement.disabled) {
        try {
            // RegisterCustomProductAuctionReq is the request structure of RegisterCustomProductAuction()
            const RegisterCustomProductAuctionReq = {
                "ID": this.Conditions.id,

                "SuggestPrice": Math.round(suggestPriceInputElement.value),
                "DistributionCenterCommission": math.PerCent.GetAsPerMyriad(distributionCenterCommissionInputElement.value),
                "SellerCommission": math.PerCent.GetAsPerMyriad(sellerCommissionInputElement.value),
                "Discount": math.PerCent.GetAsPerMyriad(discountInputElement.value),

                "DistributionCenterID": distributionCenterIDInputElement.value,
                "GroupID": groupIDInputElement.value,
                "MinNumBuy": Number(minNumBuyInputElement.value),
                "StockNumber": Number(stockNumberInputElement.value),
                "LiveUntil": Number(liveUntilInputElement.value),
                "AllowWeekdays": Number(allowWeekdaysInputElement.value),
                "AllowDayhours": Number(allowDayhoursInputElement.value),

                "Description": descriptionInputElement.value,
                "Type": Number(typeInputElement.value),
            }
            const res = await RegisterCustomProductAuction(RegisterCustomProductAuctionReq)
            pages.Router({}, "/product-auction?id=" + res.ID)
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else if (editProductAuctionElement.disabled) {
        try {
            // UpdateProductAuctionReq is the request structure of UpdateProductAuction()
            const UpdateProductAuctionReq = {
                "ID": this.Conditions.id,

                "SuggestPrice": Math.round(suggestPriceInputElement.value),
                "DistributionCenterCommission": math.PerCent.GetAsPerMyriad(distributionCenterCommissionInputElement.value),
                "SellerCommission": math.PerCent.GetAsPerMyriad(sellerCommissionInputElement.value),
                "Discount": math.PerCent.GetAsPerMyriad(discountInputElement.value),

                "MinNumBuy": Number(minNumBuyInputElement.value),
                "StockNumber": Number(stockNumberInputElement.value),
                "LiveUntil": Number(liveUntilInputElement.value),
                "AllowWeekdays": Number(allowWeekdaysInputElement.value),
                "AllowDayhours": Number(allowDayhoursInputElement.value),

                "Description": descriptionInputElement.value,
                "Type": Number(typeInputElement.value),
            }
            await UpdateProductAuction(UpdateProductAuctionReq)
            pages.Router({}, "/product-auction?id=" + this.Conditions.id)
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else {
        // TODO::: How it is possible??? need to warn user can't do anything????
    }
}

productAuctionPage.DiscardChanges = function () {

}
