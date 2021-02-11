/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/error/error.js'
import '../libjs/language/language.js'
import '../libjs/cookie.js'
import '../libjs/gs1/urn.js'
import '../libjs/sdk/sabz.city/get-quiddity.js'
import '../libjs/sdk/sabz.city/find-quiddity-by-title.js'
import '../libjs/sdk/sabz.city/find-quiddity-by-uri.js'
import '../libjs/sdk/sabz.city/get-quiddity-languages.js'
import '../libjs/sdk/sabz.city/register-quiddity.js'
import '../libjs/sdk/sabz.city/update-quiddity.js'
import '../libjs/sdk/sabz.city/register-quiddity-new-language.js'

const quiddityPage = {
    ID: "quiddity",
    Conditions: {
        id: "",
        uri: "",
        lang: "NaN",
        new: "",
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
    Icon: "translate",
    Related: ["orgs"],
    HTML: (quiddity) => ``,
    CSS: '',
    Templates: {},
}
pages.RegisterPage(quiddityPage)

quiddityPage.ConnectedCallback = async function () {
    if (isNaN(this.Conditions.lang)) {
        this.Conditions.lang = users.active.ContentPreferences.Language.ID
    } else {
        this.Conditions.lang = Number(this.Conditions.lang)
    }

    if (this.Conditions.new === "true") {
        const quiddity = this.NotFound
        window.document.body.innerHTML = this.HTML(quiddity)
        this.EnableNew()
        return
    } else if (this.Conditions.id) {
        this.showByID()
        return
    } else if (this.Conditions.uri) {
        try {
            // FindQuiddityByURIReq is the request structure of FindQuiddityByURI()
            const FindQuiddityByURIReq = {
                "URI": this.Conditions.uri,
                "Offset": 1844674407370955,
                "Limit": 1,
            }
            const FindQuiddityByURIRes = await FindQuiddityByURI(FindQuiddityByURIReq)
            this.Conditions.id = FindQuiddityByURIRes.IDs[0]
            this.showByID()
            return
        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }
    }

    const quiddity = this.NotFound
    window.document.body.innerHTML = this.HTML(quiddity)
    PersiaError.NotifyError(PersiaError.GetByID(2275209226)) // "Record Not Found"
}

quiddityPage.DisconnectedCallback = function () {
    this.resetCondition()
}

quiddityPage.resetCondition = function () {
    this.Conditions = {
        id: "",
        title: "",
        uri: "",
        lang: "NaN",
        new: "",
    }
}

quiddityPage.NotFound = {
    "WriteTime": 0,  // int64
    "AppInstanceID": "",
    "UserConnectionID": "",

    "OrgID": cookie.GetByName(HTTPCookieNameDelegateUserID),
    "URI": "urn:epc:id:gtin:",
    "Title": "Not Found",
    "Text": "Not Found",
    "Language": users.active.ContentPreferences.Language.ID,
    "Pictures": [],
}

quiddityPage.showByID = async function () {
    try {
        // GetQuiddityReq is the request structure of GetQuiddity()
        const GetQuiddityReq = {
            "ID": this.Conditions.id,
            "Language": this.Conditions.lang,
        }
        const GetQuiddityRes = await GetQuiddity(GetQuiddityReq)
        window.document.body.innerHTML = this.HTML(GetQuiddityRes)
    } catch (err) {
        PersiaError.NotifyError(err)
        return
    }
}

quiddityPage.EnableNew = function () {
    const newQuiddityElement = document.getElementById("newQuiddity")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")

    const quiddityTitleInputElement = document.getElementById("quiddityTitleInput")
    const quiddityLanguageInputElement = document.getElementById("quiddityLanguageInput")
    const quiddityURIInputInputElement = document.getElementById("quiddityURIInput")

    newQuiddityElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false

    quiddityTitleInputElement.hidden = false
    quiddityLanguageInputElement.hidden = false
    quiddityURIInputInputElement.hidden = false
}

quiddityPage.EnableEdit = function () {
    const editQuiddityElement = document.getElementById("editQuiddity")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")

    const quiddityTitleInputElement = document.getElementById("quiddityTitleInput")
    const quiddityURIInputInputElement = document.getElementById("quiddityURIInput")

    editQuiddityElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false

    quiddityTitleInputElement.hidden = false
    quiddityURIInputInputElement.hidden = false
}

quiddityPage.AddNewLanguage = function () {
    const addNewLanguageElement = document.getElementById("addNewLanguage")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")

    const quiddityTitleInputElement = document.getElementById("quiddityTitleInput")
    const quiddityLanguageInputElement = document.getElementById("quiddityLanguageInput")

    addNewLanguageElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false

    quiddityTitleInputElement.hidden = false
    quiddityLanguageInputElement.hidden = false
}

quiddityPage.SaveEdit = async function () {
    const newQuiddityElement = document.getElementById("newQuiddity")
    const editQuiddityElement = document.getElementById("editQuiddity")
    const addNewLanguageElement = document.getElementById("addNewLanguage")
    // const saveChangesElement = document.getElementById("saveChanges")
    // const discardChangesElement = document.getElementById("discardChanges")

    // const quiddityTitleElement = document.getElementById("quiddityTitle")
    const quiddityTitleInputElement = document.getElementById("quiddityTitleInput")
    // const quiddityLanguageElement = document.getElementById("quiddityLanguage")
    const quiddityLanguageInputElement = document.getElementById("quiddityLanguageInput")
    // const quiddityURIElement = document.getElementById("quiddityURI")
    const quiddityURIInputInputElement = document.getElementById("quiddityURIInput")

    const lang = language.GetSupportedByNativeName(quiddityLanguageInputElement.value)
    if (!lang) {
        PersiaError.NotifyError(PersiaError.GetByID(61004700)) // Bad Language
        return
    }

    if (newQuiddityElement.disabled) {
        try {
            // RegisterQuiddityReq is the request structure of RegisterQuiddity()
            const RegisterQuiddityReq = {
                "Language": lang.ID,
                "URI": quiddityURIInputInputElement.value,
                "Title": quiddityTitleInputElement.value,
            }
            let res = await RegisterQuiddity(RegisterQuiddityReq)
            pages.Router({}, "/quiddity?id=" + res.ID + "&lang=" + lang.ID)
            return
        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }
    } else if (editQuiddityElement.disabled) {
        try {
            // UpdateQuiddityReq is the request structure of UpdateQuiddity()
            const UpdateQuiddityReq = {
                "ID": this.Conditions.id,
                "Language": lang.ID,
                "URI": quiddityURIInputInputElement.value,
                "Title": quiddityTitleInputElement.value,
            }
            await UpdateQuiddity(UpdateQuiddityReq)
            pages.Router({}, "/quiddity?id=" + this.Conditions.id + "&lang=" + lang.ID)
            return
        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }
    } else if (addNewLanguageElement.disabled) {
        try {
            // RegisterQuiddityNewLanguageReq is the request structure of RegisterQuiddityNewLanguage()
            const RegisterQuiddityNewLanguageReq = {
                "ID": this.Conditions.id,
                "Language": lang.ID,
                "URI": quiddityURIInputInputElement.value,
                "Title": quiddityTitleInputElement.value,
            }
            await RegisterQuiddityNewLanguage(RegisterQuiddityNewLanguageReq)
            pages.Router({}, "/quiddity?id=" + this.Conditions.id + "&lang=" + lang.ID)
            return
        } catch (err) {
            PersiaError.NotifyError(err)
            return
        }
    } else {
        // TODO::: warn user can't do anything!
    }
}

quiddityPage.DiscardChanges = function () {
    alert("Sorry! Not implemented yet!")
}

/**
 * 
 * @param {Symbol[]} results 
 */
quiddityPage.FindByBarcode = async function (results) {
    const quiddityURIInputInputElement = document.getElementById("quiddityURIInput")
    switch (results[0].type) {
        case ZBar.SymbolTypes.EAN13:
            HTMLAudioElement.Beep(200, 700, 5)
            quiddityURIInputInputElement.value = GTINToURN(results[0].decode())
            break
    }
    // Check desire record exist and warn user.
    if (quiddityURIInputInputElement.value !== "") {
        try {
            // FindQuiddityByURIReq is the request structure of FindQuiddityByURI()
            const FindQuiddityByURIReq = {
                "URI": quiddityURIInputInputElement.value,
                "Offset": 1844674407370955,
                "Limit": 1,
            }
            const FindQuiddityByURIRes = await FindQuiddityByURI(FindQuiddityByURIReq)

            try {
                // GetQuiddityReq is the request structure of GetQuiddity()
                const GetQuiddityReq = {
                    "ID": FindQuiddityByURIRes.IDs[0],
                    "Language": quiddityPage.Conditions.lang,
                }
                const GetQuiddityRes = await GetQuiddity(GetQuiddityReq)
                if (GetQuiddityRes.Status === 4) PersiaError.NotifyError(103532924) // 4=QuiddityStatusBlocked
                else if (GetQuiddityRes.Status !== 2) PersiaError.NotifyError(2051920677) // 2=QuiddityStatusUnRegister
            } catch (err) {
                if (err != 2275209226) PersiaError.NotifyError(err) // "Ganjine","Record Not Found"
                return
            }
        } catch (err) {
            if (err != 2275209226) PersiaError.NotifyError(err) // "Ganjine","Record Not Found"
            return
        }
    }
}

quiddityPage.BarcodeScannerOptions = {
    CallBackResults: quiddityPage.FindByBarcode,
}