/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/errors.js'
import '../libjs/widget-barcode/camera-scanner.js'
import '../libjs/widget-barcode/barcode-reader.js'
import '../libjs/widget-notification/force-leave-page.js'
import '../../sdk-js/find-wiki-by-title.js'
import '../../sdk-js/find-wiki-by-uri.js'
import '../../sdk-js/get-wiki-by-id.js'
import '../../sdk-js/get-person-number-status.js'
import '../../sdk-js/get-organization-by-name.js'
import '../../sdk-js/get-organization-by-id.js'
import '../../sdk-js/register-new-person.js'

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
    Related: ["store", "wikis"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "new": () => ``,
        "row": (pr) => ``
    },
}
pages.RegisterPage(invoicePage)

invoicePage.ConnectedCallback = function () {
    // TODO::: Do any logic before page render
    window.document.body.innerHTML = this.HTML()
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
                pages.Router({}, "/wiki?uri=" + element.value)
            } else {
                pages.Router({}, "/wiki?title=" + element.value)
            }
        }
    } else {

    }
}

invoicePage.FindInputSearchEvent = async function (element) {

}

invoicePage.EnableNew = async function () {
    if (this.newInvoiceList) {
        let forceLeave = await forceLeavePageWidget.ConnectedCallback()
        if (!forceLeave) {
            return
        }
        window.document.body.innerHTML = this.HTML()
    }

    const mainElement = document.getElementById('main')
    mainElement.insertAdjacentHTML('beforeend', this.Templates["new"]())

    this.newInvoiceList = {
        poolByID: {},
        poolByURI: {},
        poolByTitle: {},
        RowNumber: 0,
        SuggestPrice: 0,
        PayablePrice: 0,
        ProductNumber: 0,
    }
    this.tableBodyElement = document.getElementById('tableBody')
    this.tableFooterElement = document.getElementById('tableFooter')
}

invoicePage.addProductToListByURIInput = function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        const uri = element.value
        if (uri.length === 13) {
            this.addProductToListByURI("urn:gs1:ean13:" + uri)
        } else {
            element.setCustomValidity("Not valid barcode!")
            element.reportValidity()
            return
        }
        element.value = ""
        element.parentElement.toggle()
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
        // let GetWikiByIDRes
        // try {
        //     // FindWikiByURIReq is the request structure of FindWikiByURI()
        //     const FindWikiByURIReq = {
        //         "URI": uri, // string
        //         "Offset": 1844674407370955,
        //         "Limit": 1,
        //     }
        //     let FindWikiByURIRes = FindWikiByURI(FindWikiByURIReq)

        //     try {
        //         // GetWikiByIDReq is the request structure of GetWikiByID()
        //         const GetWikiByIDReq = {
        //             "ID": FindWikiByURIRes.IDs[0],
        //             "Language": users.active.ContentPreferences.Language.id,
        //         }
        //         GetWikiByIDRes = await GetWikiByID(GetWikiByIDReq)
        //         GetWikiByIDRes.ID = GetWikiByIDReq.ID
        //     } catch (err) {
        //         errors.HandleError(err)
        //         return
        //     }
        // } catch (err) {
        //     errors.HandleError(err)
        //     return
        // }

        productInfo = {
            // WikiID: GetWikiByIDRes.ID,
            // URI: GetWikiByIDRes.URI,
            // Title: GetWikiByIDRes.Title,
            URI: uri,
            Title: uri,
            SuggestPrice: 12500,
            PayablePrice: 10000,
            ProductNumber: 1,
            Actions: "",
        }
    }

    // Tell user that product recognized!
    barcodeCameraScannerWidget.Beep(200, 700, 5)

    this.addProductToList(productInfo)
}

invoicePage.addProductToListByTitle = async function (title) {
    let productInfo = this.newInvoiceList.poolByTitle[title]
    if (!productInfo) {
        // let GetWikiByIDRes
        // try {
        //     // FindWikiByTitleReq is the request structure of FindWikiByTitle()
        //     const FindWikiByTitleReq = {
        //         "Title": title,
        //         "Offset": 1844674407370955,
        //         "Limit": 1,
        //     }
        //     let FindWikiByTitleRes = await FindWikiByTitle(FindWikiByTitleReq)

        //     try {
        //         // GetWikiByIDReq is the request structure of GetWikiByID()
        //         const GetWikiByIDReq = {
        //             "ID": FindWikiByTitleRes.IDs[0],
        //             "Language": users.active.ContentPreferences.Language.id,
        //         }
        //         GetWikiByIDRes = await GetWikiByID(GetWikiByIDReq)
        //         GetWikiByIDRes.ID = GetWikiByIDReq.ID
        //     } catch (err) {
        //         errors.HandleError(err)
        //         return
        //     }
        // } catch (err) {
        //     errors.HandleError(err)
        //     return
        // }

        productInfo = {
            // WikiID: GetWikiByIDRes.ID,
            // URI: GetWikiByIDRes.URI,
            // Title: GetWikiByIDRes.Title,
            Title: title,
            SuggestPrice: 12500,
            PayablePrice: 10000,
            ProductNumber: 1,
            Actions: "",
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

        this.tableBodyElement.insertAdjacentHTML('beforeend', this.Templates["row"](productInfo))
        productInfo.Element = this.tableBodyElement.lastChild

        this.newInvoiceList.poolByID[productInfo.ID] = productInfo
        this.newInvoiceList.poolByURI[productInfo.URI] = productInfo
        this.newInvoiceList.poolByTitle[productInfo.Title] = productInfo

        this.newInvoiceList.SuggestPrice += productInfo.SuggestPrice
        this.newInvoiceList.PayablePrice += productInfo.PayablePrice
        this.newInvoiceList.ProductNumber++
        this.updateTableFooter()
    }
}

invoicePage.updateTableFooter = function () {
    this.tableFooterElement.children[0].innerText = this.newInvoiceList.RowNumber
    this.tableFooterElement.children[2].innerText = this.newInvoiceList.SuggestPrice
    this.tableFooterElement.children[3].innerText = this.newInvoiceList.PayablePrice
    this.tableFooterElement.children[4].innerText = this.newInvoiceList.ProductNumber
}

invoicePage.increaseProductNumber = function (element) {
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
    const rowElement = element.parentElement.parentElement
    const ID = rowElement.children[0].innerText
    const productInfo = this.newInvoiceList.poolByID[ID]

    delete this.newInvoiceList.poolByID[productInfo.ID]
    delete this.newInvoiceList.poolByTitle[productInfo.Title]
    delete this.newInvoiceList.poolByURI[productInfo.URI]
    rowElement.remove()

    this.newInvoiceList.RowNumber--
    this.newInvoiceList.SuggestPrice -= (productInfo.PayablePrice * productInfo.ProductNumber)
    this.newInvoiceList.PayablePrice -= (productInfo.PayablePrice * productInfo.ProductNumber)
    this.newInvoiceList.ProductNumber -= productInfo.ProductNumber
    this.updateTableFooter()
}

invoicePage.checkDCNameElement = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        try {
            // GetOrganizationByNameReq is the request structure of GetOrganizationByName()
            const GetOrganizationByNameReq = {
                "Name": element.value,
            }
            let res = await GetOrganizationByName(GetOrganizationByNameReq, true)

            try {
                // GetOrganizationByIDReq is the request structure of GetOrganizationByID()
                const GetOrganizationByIDReq = {
                    "ID": res.ID,
                }
                org = await GetOrganizationByID(GetOrganizationByIDReq, true)
                this.newInvoiceList.DC = org

                localStorage.setItem('LastDCNameInInvoicePage', org.Name)
                localStorage.setItem('LastDCIDInInvoicePage', org.ID)
            } catch (err) {
                return errors.HandleError(err)
            }
        } catch (err) {
            return errors.HandleError(err)
        }
    }
}

invoicePage.checkDCIDElement = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        try {
            // GetOrganizationByIDReq is the request structure of GetOrganizationByID()
            const GetOrganizationByIDReq = {
                "ID": element.value,
            }
            let org = await GetOrganizationByID(GetOrganizationByIDReq, true)
            this.newInvoiceList.DC = org

            localStorage.setItem('LastDCNameInInvoicePage', org.Name)
            localStorage.setItem('LastDCIDInInvoicePage', org.ID)
        } catch (err) {
            errors.HandleError(err)
        }
    }
}

invoicePage.checkBuyerTelElement = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        let tel = element.value

        // TODO::: Need to validate phone number here?
        if (!tel) {
            element.setCustomValidity("Number Not Valid")
            element.reportValidity()
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
            } else {
                document.getElementById('suggestRegisterUser').toggle()
            }
        } catch (err) {
            return errors.HandleError(err)
        }

        element.parentElement.toggle()
    }
}

invoicePage.registerUser = async function () {
    try {
        const RegisterNewPersonReq = {
            "PhoneNumber": this.newInvoiceList.BuyerTelNumber,
        }
        let res = await RegisterNewPerson(RegisterNewPersonReq)
        this.newInvoiceList.BuyerUserID = res.PersonID
        document.getElementById('suggestRegisterUser').toggle()
    } catch (err) {
        errors.HandleError(err)
    }
}

invoicePage.lastCheckoutCheck = async function () {
    if (!this.newInvoiceList.BuyerUserID || this.newInvoiceList.BuyerUserID) {
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

invoicePage.checkoutInvoice = async function () {
    document.getElementById('getLastApproved').toggle()

    const personOTPInput = document.getElementById('personOTPInput')
    if (personOTPInput.value) {

    }

    alert("Sorry! Not implemented yet!")
    // this.newInvoiceList = null
}

/**
 * 
 * @param {Symbol[]} results 
 */
invoicePage.handleBarcodeCameraScanner = async function (results) {
    for (let res of results) {
        switch (res.type) {
            case ZBar.SymbolTypes.EAN13:
                invoicePage.addProductToListByURI("urn:gs1:ean13:" + res.decode())
                return
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
    if (barcodeString.length === 13) {
        invoicePage.addProductToListByURI("urn:gs1:ean13:" + barcodeString)
    }
}

invoicePage.barcodeReaderWidgetOptions = {
    CallBackResults: invoicePage.handleBarcodeReader,
}
