/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/errors.js'
import '../libjs/widget-barcode/camera-scanner.js'
import '../../sdk-js/find-wiki-by-org-id.js'
import '../../sdk-js/get-wiki-by-id.js'

const wikisPage = {
    ID: "wikis",
    Conditions: {
        org: "",
        lang: "NaN",
        offset: "NaN",
        limit: "NaN",
    },
    State: "",
    Robots: "noindex, follow, noarchive, noimageindex",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "translate",
    Related: ["", ""],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "wiki-card": (wiki) => ``,
    },
}
pages.RegisterPage(wikisPage)

wikisPage.ConnectedCallback = async function () {
    window.document.body.innerHTML = this.HTML()

    if (isNaN(this.Conditions.lang)) {
        // TODO::: write user lang
        this.Conditions.lang = 0
    }
    if (isNaN(this.Conditions.offset)) {
        this.Conditions.offset = 1844674407370955161
    } else {
        this.Conditions.offset = Number(this.Conditions.offset)
    }
    if (isNaN(this.Conditions.limit)) {
        this.Conditions.limit = 9
    } else {
        this.Conditions.limit = Number(this.Conditions.limit)
    }

    if (this.Conditions.org) {
        try {
            // GetWikiIDsByOrgIDReq is the request structure of GetWikiIDsByOrgID()
            const GetWikiIDsByOrgIDReq = {
                "OrgID": this.Conditions.org,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const GetWikiIDsByOrgIDRes = await GetWikiIDsByOrgID(GetWikiIDsByOrgIDReq)
            this.showByIDs(GetWikiIDsByOrgIDRes.IDs)
        } catch (err) {
            errors.HandleError(err)
            return
        }
    }
}

wikisPage.DisconnectedCallback = function () {
    this.resetCondition()
    barcodeCameraScannerWidget.DisconnectedCallback()
}

wikisPage.resetCondition = function () {
    this.Conditions = {
        org: "",
        lang: "NaN",
        offset: "NaN",
        limit: "NaN",
    }
}

wikisPage.showByIDs = async function (IDs) {
    const mainElement = document.getElementById('main')
    for (let id of IDs) {
        try {
            // GetWikiByIDReq is the request structure of GetWikiByID()
            const GetWikiByIDReq = {
                "ID": id,
                "Language": this.Conditions.lang,
            }
            const GetWikiByIDRes = await GetWikiByID(GetWikiByIDReq)
            GetWikiByIDRes.ID = id
            GetWikiByIDRes.Language = this.Conditions.lang
            mainElement.insertAdjacentHTML('beforeend', this.Templates["wiki-card"](GetWikiByIDRes))
        } catch (err) {
            errors.HandleError(err)
        }
    }
}

wikisPage.ToggleFindInput = function () {
    const findWikiInputElement = document.getElementById("findWikiInput")
    const mainElement = document.getElementById('main')
    mainElement.innerText = ""

    if (findWikiInputElement.hidden) {
        findWikiInputElement.hidden = false
    } else {
        findWikiInputElement.hidden = true
    }
}

wikisPage.FindWikiInputKeyDownEvent = async function (element) {
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

wikisPage.FindWikiInputSearchEvent = async function (element) {

}

/**
 * 
 * @param {Symbol[]} results 
 */
wikisPage.FindByBarcode = async function (results) {
    const mainElement = document.getElementById('main')
    for (let res of results) {
        mainElement.innerText += res.decode() + "\n"
        switch (res.type) {
            case ZBar.SymbolTypes.EAN13:
                // TODO::: first check desire record exist.
                barcodeCameraScannerWidget.Beep(200, 500, 5)
                // pages.Router({}, "/wiki?uri=urn:gs1:ean13:" + res.decode())
                errors.HandleError(errors.poolByID[1685872164])
                return
        }
    }
}

wikisPage.BarcodeScannerOptions = {
    CallBackResults: wikisPage.FindByBarcode,
}
