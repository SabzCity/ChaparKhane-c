/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/error/error.js'
import '../libjs/gs1/urn.js'
import '../libjs/widget-barcode/camera-scanner.js'
import '../libjs/widget-barcode/barcode-reader.js'
import '../libjs/widget-notification/force-leave-page.js'
import '../libjs/sdk/sabz.city/find-quiddity-by-title.js'
import '../libjs/sdk/sabz.city/find-quiddity-by-uri.js'
import '../libjs/sdk/sabz.city/get-quiddity.js'
import '../libjs/sdk/sabz.city/get-person-number-status.js'
import '../libjs/sdk/sabz.city/get-organization.js'
import '../libjs/sdk/sabz.city/register-person.js'
import '../libjs/sdk/sabz.city/register-product-invoice.js'

const invoicePage = {
    ID: "invoice",
    Conditions: {
        id: "",
    },
    State: "",
    Robots: "none",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "note",
    Related: ["store", "quiddities"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "row": (pr) => ``
    },
}
pages.RegisterPage(invoicePage)

invoicePage.ConnectedCallback = function () {
    // TODO::: Do any logic before page render
    this.EnableNew()
    // TODO::: Do any logic after page render
}

invoicePage.DisconnectedCallback = async function () {
    if (this.newInvoiceList) {
        var forceLeave = await forceLeavePageWidget.ConnectedCallback()
    }
    if (forceLeave) {
        this.newInvoiceList = null
        barcodeCameraScannerWidget.DisconnectedCallback()
    }
    return forceLeave
}

invoicePage.ToggleFindInput = function () {
    const findInputElement = document.getElementById("findInput")
    if (findInputElement.hidden) {
        findInputElement.hidden = false
    } else {
        findInputElement.hidden = true
    }
}

invoicePage.FindInputKeyDownEvent = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        if (element.value !== "") {
            if (element.value.includes(":")) {
                pages.Router({}, "/quiddity?uri=" + element.value)
            } else {
                pages.Router({}, "/quiddity?title=" + element.value)
            }
        }
    } else {

    }
}

invoicePage.FindInputSearchEvent = async function (element) {

}

invoicePage.EnableNew = async function () {
    if (this.newInvoiceList) {
        if (this.newInvoiceList.State === InvoiceStateRegistered) {

        }
        let forceLeave = await forceLeavePageWidget.ConnectedCallback()
        if (!forceLeave) {
            return
        }
    }

    window.document.body.innerHTML = this.HTML()

    this.newInvoiceList = {
        PosID: localStorage.getItem('LastPosIDInInvoiceTransferPage') || "",
        poolByID: {},
        poolByURI: {},
        poolByTitle: {},
        RowNumber: 0,
        SuggestPrice: 0,
        PayablePrice: 0,
        ProductNumber: 0,
        State: InvoiceStateNew,
    }
    this.tableBodyElement = document.getElementById('tableBody')
    this.tableFooterElement = document.getElementById('tableFooter')
}

invoicePage.addProductToListByURIInput = function (element) {
    // Tell user that some key down recognized!
    HTMLAudioElement.Beep(50, 300, 3)

    if (event.keyCode === 13) { // event.key === 'Enter'
        const uri = element.value
        switch (uri.length) {
            case 8:
            case 12:
            case 13:
            case 14:
                this.addProductToListByURI(GTINToURN(uri))
                element.value = ""
                element.parentElement.toggle()
                return
            case 16:
                this.addProductToListByURI("urn:epc:id:iran:" + uri)
                return
            default:
                element.warnValidity()
                return
        }
    }
}

invoicePage.addProductToListByTitleInput = function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        const title = element.value
        this.addProductToListByTitle(title)
        element.value = ""
        element.parentElement.toggle()
    }
}

invoicePage.addProductToListByURI = async function (uri) {
    let productInfo = this.newInvoiceList.poolByURI[uri]
    if (!productInfo) {
        // let GetQuiddityRes
        // let GetProductPriceRes
        // let GetProductAuctionRes
        try {
            const FindQuiddityByURIReq = {
                "URI": uri,
                "Offset": 1844674407370955,
                "Limit": 1,
            }
            const FindQuiddityByURIRes = await FindQuiddityByURI(FindQuiddityByURIReq)

            const GetQuiddityReq = {
                "ID": FindQuiddityByURIRes.IDs[0],
                "Language": users.active.ContentPreferences.Language.ID,
            }
            var GetQuiddityRes = await GetQuiddity(GetQuiddityReq)

            const GetProductPriceReq = {
                "QuiddityID": FindQuiddityByURIRes.IDs[0],
            }
            var GetProductPriceRes = await GetProductPrice(GetProductPriceReq)

            // TODO::: improve to check other auction not just default one.
            const FindProductAuctionByQuiddityIDReq = {
                "QuiddityID": FindQuiddityByURIRes.IDs[0],
                "Offset": 0,
                "Limit": 1,
            }
            const FindProductAuctionByQuiddityIDRes = await FindProductAuctionByQuiddityID(FindProductAuctionByQuiddityIDReq)

            // GetProductAuctionReq is the request structure of GetProductAuction()
            const GetProductAuctionReq = {
                "ID": FindProductAuctionByQuiddityIDRes.IDs[0],
            }
            var GetProductAuctionRes = await GetProductAuction(GetProductAuctionReq)

        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }

        productInfo = {
            Quiddity: GetQuiddityRes,
            ProductPrice: GetProductPriceRes,
            ProductAuction: GetProductAuctionRes,
            ProductNumber: 1,
        }
    }
    this.addProductToList(productInfo)
}

invoicePage.addProductToListByTitle = async function (title) {
    alert("Sorry! Not implemented yet!")
    return
    let productInfo = this.newInvoiceList.poolByTitle[title]
    if (!productInfo) {
        let GetQuiddityRes
        // try {
        //     // FindQuiddityByTitleReq is the request structure of FindQuiddityByTitle()
        //     const FindQuiddityByTitleReq = {
        //         "Title": title,
        //         "Offset": 1844674407370955,
        //         "Limit": 1,
        //     }
        //     let FindQuiddityByTitleRes = await FindQuiddityByTitle(FindQuiddityByTitleReq)

        //     try {
        //         // GetQuiddityReq is the request structure of GetQuiddity()
        //         const GetQuiddityReq = {
        //             "ID": FindQuiddityByTitleRes.IDs[0],
        //             "Language": users.active.ContentPreferences.Language.ID,
        //         }
        //         GetQuiddityRes = await GetQuiddity(GetQuiddityReq)
        //     } catch (err) {
        //         PersiaError.NotifyError(err)
        //         return
        //     }
        // } catch (err) {
        //     PersiaError.NotifyError(err)
        //     return
        // }

        productInfo = {
            Quiddity: GetQuiddityRes,
            ProductPrice: GetProductPriceRes,
            ProductAuction: GetProductAuctionRes,
            ProductNumber: 1,
        }
    }
    this.addProductToList(productInfo)
}

invoicePage.addProductToList = async function (productInfo) {
    if (productInfo.ID) {
        productInfo.ProductNumber++
        productInfo.Element.children[4].innerText = productInfo.ProductNumber

        this.newInvoiceList.SuggestPrice += productInfo.SuggestPrice
        this.newInvoiceList.PayablePrice += productInfo.PayablePrice
        this.newInvoiceList.ProductNumber++
        this.updateTableFooter()
    } else {
        this.newInvoiceList.RowNumber++
        productInfo.ID = this.newInvoiceList.RowNumber
        productInfo.Title = productInfo.Quiddity.Title
        productInfo.SuggestPrice = productInfo.ProductPrice.Price
        productInfo.PayablePrice = Math.ceil(productInfo.ProductPrice.Price - math.PerMyriad.Calculate(productInfo.ProductPrice.Price, productInfo.ProductAuction.Discount))

        this.tableBodyElement.insertAdjacentHTML('beforeend', this.Templates["row"](productInfo))
        productInfo.Element = this.tableBodyElement.lastChild

        this.newInvoiceList.poolByID[productInfo.ID] = productInfo
        this.newInvoiceList.poolByURI[productInfo.Quiddity.URI] = productInfo
        this.newInvoiceList.poolByTitle[productInfo.Title] = productInfo

        this.newInvoiceList.SuggestPrice += productInfo.SuggestPrice
        this.newInvoiceList.PayablePrice += productInfo.PayablePrice
        this.newInvoiceList.ProductNumber++
        this.updateTableFooter()
    }
}

invoicePage.updateTableFooter = function () {
    // Tell user that some changes to invoice list recognized and occur!
    HTMLAudioElement.Beep(200, 700, 5)

    this.tableFooterElement.children[0].innerText = this.newInvoiceList.RowNumber
    this.tableFooterElement.children[2].innerText = this.newInvoiceList.SuggestPrice.toLocaleString()
    this.tableFooterElement.children[3].innerText = this.newInvoiceList.PayablePrice.toLocaleString()
    this.tableFooterElement.children[4].innerText = this.newInvoiceList.ProductNumber

    window.scrollTo(0, document.body.scrollHeight);
}

invoicePage.increaseProductNumber = function (element) {
    element.blur()
    const rowElement = element.parentElement.parentElement
    const ID = rowElement.children[0].innerText
    const productInfo = this.newInvoiceList.poolByID[ID]

    productInfo.ProductNumber++
    rowElement.children[4].innerText = productInfo.ProductNumber

    this.newInvoiceList.SuggestPrice += productInfo.SuggestPrice
    this.newInvoiceList.PayablePrice += productInfo.PayablePrice
    this.newInvoiceList.ProductNumber++
    this.updateTableFooter()
}

invoicePage.decreaseProductNumber = function (element) {
    element.blur()
    const rowElement = element.parentElement.parentElement
    const ID = rowElement.children[0].innerText
    const productInfo = this.newInvoiceList.poolByID[ID]

    if (productInfo.ProductNumber === 0) return

    productInfo.ProductNumber--
    rowElement.children[4].innerText = productInfo.ProductNumber

    this.newInvoiceList.SuggestPrice -= productInfo.SuggestPrice
    this.newInvoiceList.PayablePrice -= productInfo.PayablePrice
    this.newInvoiceList.ProductNumber--
    this.updateTableFooter()
}

invoicePage.removeProductNumber = function (element) {
    element.blur()
    const rowElement = element.parentElement.parentElement
    const ID = rowElement.children[0].innerText
    const productInfo = this.newInvoiceList.poolByID[ID]

    delete this.newInvoiceList.poolByID[productInfo.ID]
    delete this.newInvoiceList.poolByTitle[productInfo.Title]
    delete this.newInvoiceList.poolByURI[productInfo.Quiddity.URI]
    rowElement.remove()

    this.newInvoiceList.RowNumber--
    this.newInvoiceList.SuggestPrice -= (productInfo.SuggestPrice * productInfo.ProductNumber)
    this.newInvoiceList.PayablePrice -= (productInfo.PayablePrice * productInfo.ProductNumber)
    this.newInvoiceList.ProductNumber -= productInfo.ProductNumber
    this.updateTableFooter()
}

invoicePage.checkDCNameElement = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        try {
            // FindOrganizationByNameReq is the request structure of FindOrganizationByName()
            const FindOrganizationByNameReq = {
                "Name": element.value,
            }
            let res = await FindOrganizationByName(FindOrganizationByNameReq, true)

            try {
                // GetOrganizationReq is the request structure of GetOrganization()
                const GetOrganizationReq = {
                    "ID": res.ID,
                }
                org = await GetOrganization(GetOrganizationReq, true)
                this.newInvoiceList.DC = org

                localStorage.setItem('LastDCNameInInvoicePage', org.Name)
                localStorage.setItem('LastSenderDCIDInInvoicePage', org.ID)
            } catch (err) {
                return PersiaError.NotifyError(err)
            }
        } catch (err) {
            return PersiaError.NotifyError(err)
        }
    }
}

invoicePage.posIDInput = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        this.newInvoiceList.PosID = element.value
        localStorage.setItem('LastPosIDInInvoiceTransferPage', element.value)
        element.focusNext()
    }
}

invoicePage.checkBuyerTelElement = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        let tel = element.value
        if (!tel) {
            element.warnValidity()
            return
        }
        if (tel.startsWith("0")) tel = tel.substring(1)
        const phoneNumber = "98" + tel
        this.newInvoiceList.BuyerTelNumber = Number(phoneNumber)

        // check phone number registered before this request!
        try {
            const GetPersonNumberStatusReq = {
                PhoneNumber: this.newInvoiceList.BuyerTelNumber
            }
            let res = await GetPersonNumberStatus(GetPersonNumberStatusReq)
            if (res.Status == 1 || res.Status == 3) {
                this.newInvoiceList.BuyerUserID = res.PersonID
                element.focusNext()
            } else {
                document.getElementById('suggestRegisterUserNumber').innerText = this.newInvoiceList.BuyerTelNumber
                document.getElementById('suggestRegisterUser').toggle()
            }
        } catch (err) {
            return PersiaError.NotifyError(err)
        }

        // element.parentElement.toggle()
    }
}

invoicePage.registerUser = async function () {
    try {
        const RegisterPersonReq = {
            "PhoneNumber": this.newInvoiceList.BuyerTelNumber,
        }
        let res = await RegisterPerson(RegisterPersonReq)
        this.newInvoiceList.BuyerUserID = res.PersonID
        this.sendTransfer()
        document.getElementById('suggestRegisterUser').toggle()
    } catch (err) {
        PersiaError.NotifyError(err)
    }
}

invoicePage.sendTransfer = async function (buttonElement) {
    if (!this.newInvoiceList.BuyerUserID) {
        PersiaError.NotifyError("خریدار جهت ثبت فاکتور ثبت نشده است")
        return
    }

    buttonElement.disabled = true
    this.newInvoiceList.State = InvoiceStateSendToPOS
    const invoiceStateContainerElement = document.getElementById("invoiceStateContainer")
    invoiceStateContainerElement.innerText = "در حال پرداخت توسط خریدار ..."
    try {
        const RegisterFinancialTransactionReq = {
            "FromSocietyID": 2, // TODO:::
            "FromUserID": users.active.UserID,
            "PosID": "00" + this.newInvoiceList.PosID, // TODO::: 00 is hack user must choose way

            "Amount": Math.ceil(this.newInvoiceList.PayablePrice), // Math.round(this.newInvoiceList.PayablePrice),

            "ToSocietyID": 2, // TODO:::
            "ToUserID": this.newInvoiceList.BuyerUserID,
        }
        const RegisterFinancialTransactionRes = await RegisterFinancialTransaction(RegisterFinancialTransactionReq)
        localStorage.setItem('LastTransferPaymentIDInInvoicePage', RegisterFinancialTransactionRes.ID)
        invoiceStateContainerElement.innerText = "پرداخت توسط خریدار انجام شده"
        this.newInvoiceList.State = InvoiceStatePayed

        try {
            const res = await this.checkoutInvoice()

            // TODO::: close and show dialog about action.
            alert("Transaction complete with this ID: " + RegisterFinancialTransactionRes.ID + "\nTransaction complete with this not registered items:", res.NotRegistered)
            this.EnableNew()
        } catch (err) {
            PersiaError.NotifyError(err)
        }
    } catch (err) {
        PersiaError.NotifyError(err)
        invoiceStateContainerElement.innerText = "خطا در پرداخت توسط خریدار ..."
    }
    buttonElement.disabled = false
}

invoicePage.lastCheckoutCheck = async function () {
    if (!this.newInvoiceList.BuyerUserID) {
        this.newInvoiceList.BuyerUserID = users.active.UserID
        this.newInvoiceList.BuyerTelNumber = users.active.UserNumber
    } else {
        const personOTPInput = document.getElementById('personOTPInput')
        personOTPInput.hidden = false
    }

    // Check user account balance

    document.getElementById('getLastApproved').toggle()
    document.getElementById('lastApproveBuyerNumber').innerText = this.newInvoiceList.BuyerTelNumber
    document.getElementById('lastApprovePayablePrice').innerText = this.newInvoiceList.PayablePrice
}

invoicePage.approveCheckout = async function () {
    document.getElementById('getLastApproved').toggle()
    try {
        const res = await this.checkoutInvoice()
        // TODO::: close and show dialog about action.
        alert("Transaction complete with this not registered items:", res.NotRegistered)
        this.EnableNew()
    } catch (err) {
        PersiaError.NotifyError(err)
    }
}

invoicePage.checkoutInvoice = async function () {
    const personOTPInput = document.getElementById('personOTPInput')
    const invoiceStateContainerElement = document.getElementById("invoiceStateContainer")
    const RegisterProductInvoiceReq = {
        "UserID": this.newInvoiceList.BuyerUserID,
        "UserOTP": Number(personOTPInput.value),
        "UserTransactionID": localStorage.getItem('LastTransferPaymentIDInInvoicePage') || "",
        "SenderDCID": "cfiiAF6pxrG15E50WcmqWGOJj7eutxifP04LOwPIEEg", // localStorage.getItem('LastSenderDCIDInInvoicePage'),
        // "ReceiverDCID": localStorage.getItem('LastReceiverDCIDInInvoicePage'),
        "Language": users.active.ContentPreferences.Language.ID,
        "Products": [],
    }
    for (const pr in this.newInvoiceList.poolByID) {
        RegisterProductInvoiceReq.Products.push({
            "QuiddityID": this.newInvoiceList.poolByID[pr].Quiddity.ID,
            "ProductAuctionID": this.newInvoiceList.poolByID[pr].ProductAuction.ID,
            "Number": this.newInvoiceList.poolByID[pr].ProductNumber,
        })
    }
    try {
        let RegisterProductInvoiceRes = await RegisterProductInvoice(RegisterProductInvoiceReq)
        this.newInvoiceList.State = InvoiceStateRegistered
        invoiceStateContainerElement.innerText = "فاکتور خریدار ثبت نهایی شده است"
        return RegisterProductInvoiceRes
    } catch (err) {
        invoiceStateContainerElement.innerText = "خطا در  ثبت فاکتور ..."
        throw err
    }
}

/**
 * 
 * @param {Symbol[]} results 
 */
invoicePage.handleBarcodeCameraScanner = async function (results) {
    for (let res of results) {
        switch (res.type) {
            case ZBar.SymbolTypes.EAN8:
            case ZBar.SymbolTypes.EAN13:
                invoicePage.addProductToListByURI(GTINToURN(res.decode()))
                break
            default:
                invoicePage.addProductToListByURI("urn:epc:id:iran:" + res.decode())
                break
        }
    }
}

invoicePage.barcodeCameraScannerWidgetOptions = {
    CallBackResults: invoicePage.handleBarcodeCameraScanner,
}

/**
 * 
 * @param {string} barcodeString 
 */
invoicePage.handleBarcodeReader = async function (barcodeString) {
    switch (barcodeString.length) {
        case 8:
        case 12:
        case 13:
        case 14:
            invoicePage.addProductToListByURI(GTINToURN(barcodeString))
            break
        case 16:
            invoicePage.addProductToListByURI("urn:epc:id:iran:" + barcodeString)
            break
    }
}

invoicePage.barcodeReaderWidgetOptions = {
    CallBackResults: invoicePage.handleBarcodeReader,
}

const InvoiceStateNew = 0
const InvoiceStateSendToPOS = 1
const InvoiceStatePayed = 2
const InvoiceStateRegistered = 3
