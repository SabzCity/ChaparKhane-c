/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/error/error.js'
import '../libjs/widget-barcode/camera-scanner.js'
import '../libjs/sdk/sabz.city/find-quiddity-by-org-id.js'
import '../libjs/sdk/sabz.city/get-quiddity.js'

const quidditiesPage = {
    ID: "quiddities",
    Conditions: {
        title: "",
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
    Related: [],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "quiddity-card": (quiddity) => ``,
    },
}
pages.RegisterPage(quidditiesPage)

quidditiesPage.ConnectedCallback = async function () {
    window.document.body.innerHTML = this.HTML()

    if (isNaN(this.Conditions.lang)) {
        this.Conditions.lang = users.active.ContentPreferences.Language.ID
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

    if (this.Conditions.title) {
        try {
            // FindQuiddityByTitleReq is the request structure of FindQuiddityByTitle()
            const FindQuiddityByTitleReq = {
                "Title": this.Conditions.title,
                "PageNumber": 0,
            }
            const FindQuiddityByTitleRes = await FindQuiddityByTitle(FindQuiddityByTitleReq)
            let IDs = []
            for (let token of FindQuiddityByTitleRes.Tokens) {
                if (token.RecordPrimaryKey !== "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA") IDs.push(token.RecordPrimaryKey)
            }
            this.showByIDs(IDs)
        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }
    } else if (this.Conditions.org) {
        try {
            // FindQuiddityByOrgIDReq is the request structure of FindQuiddityByOrgID()
            const FindQuiddityByOrgIDReq = {
                "OrgID": this.Conditions.org,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const FindQuiddityByOrgIDRes = await FindQuiddityByOrgID(FindQuiddityByOrgIDReq)
            this.showByIDs(FindQuiddityByOrgIDRes.IDs)
        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }
    }
}

quidditiesPage.DisconnectedCallback = function () {
    this.resetCondition()
    barcodeCameraScannerWidget.DisconnectedCallback()
}

quidditiesPage.resetCondition = function () {
    this.Conditions = {
        org: "",
        lang: "NaN",
        offset: "NaN",
        limit: "NaN",
    }
}

quidditiesPage.showByIDs = async function (IDs) {
    const mainElement = document.getElementById('main')
    for (let id of IDs) {
        try {
            // GetQuiddityReq is the request structure of GetQuiddity()
            const GetQuiddityReq = {
                "ID": id,
                "Language": Number(this.Conditions.lang) || users.active.ContentPreferences.Language.ID,
            }
            const GetQuiddityRes = await GetQuiddity(GetQuiddityReq)
            mainElement.insertAdjacentHTML('beforeend', this.Templates["quiddity-card"](GetQuiddityRes))
        } catch (err) {
            PersiaError.NotifyError(err)
        }
    }
}

quidditiesPage.ToggleFindInput = function () {
    const findQuiddityInputContainerElement = document.getElementById("findQuiddityInputContainer")
    const mainElement = document.getElementById('main')
    mainElement.innerText = ""

    if (findQuiddityInputContainerElement.hidden) {
        findQuiddityInputContainerElement.hidden = false
    } else {
        findQuiddityInputContainerElement.hidden = true
    }
}

quidditiesPage.FindQuiddityInputKeyDownEvent = async function (element) {
    if (event.keyCode === 13) { // event.key === 'Enter'
        if (element.value !== "") {
            if (!isNaN(element.value)) {
                pages.Router({}, "/quiddity?uri=" + GTINToURN(element.value))
            } else if (element.value.includes(":") || element.value.includes(".")) {
                pages.Router({}, "/quiddity?uri=" + element.value)
            } else {
                pages.Router({}, "/quiddities?title=" + element.value)
            }
        }
    } else {

    }
}

quidditiesPage.FindQuiddityInputSearchEvent = async function (element) {

}

/**
 * 
 * @param {Symbol[]} results 
 */
quidditiesPage.FindByBarcode = async function (results) {
    const mainElement = document.getElementById('main')
    for (let res of results) {
        mainElement.innerText += res.decode() + "\n"
        switch (res.type) {
            case ZBar.SymbolTypes.EAN13:
                // TODO::: first check desire record exist.
                HTMLAudioElement.Beep(200, 700, 5)
                pages.Router({}, "/quiddity?uri=" + GTINToURN(res.decode()))
                return
        }
    }
}

quidditiesPage.BarcodeScannerOptions = {
    CallBackResults: quidditiesPage.FindByBarcode,
}
