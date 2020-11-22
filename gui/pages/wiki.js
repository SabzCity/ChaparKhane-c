/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/errors.js'
import '../libjs/language.js'
import '../libjs/cookie.js'
import '../../sdk-js/get-wiki-by-id.js'
import '../../sdk-js/get-wiki-by-title.js'
import '../../sdk-js/get-wiki-by-uri.js'
import '../../sdk-js/get-wiki-languages.js'
import '../../sdk-js/register-new-wiki.js'
import '../../sdk-js/update-wiki.js'
import '../../sdk-js/register-wiki-new-language.js'

const wikiPage = {
    ID: "wiki",
    Conditions: {
        id: "",
        title: "",
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
    HTML: (wiki) => ``,
    CSS: '',
    Templates: {},
}
pages.RegisterPage(wikiPage)

wikiPage.ConnectedCallback = async function () {
    if (isNaN(this.Conditions.lang)) {
        this.Conditions.lang = users.active.ContentPreferences.Language.id
    }

    this.Conditions.lang = Number(this.Conditions.lang)

    if (this.Conditions.new === "true") {
        const wiki = this.NotFound()
        window.document.body.innerHTML = this.HTML(wiki)
        this.EnableNew()
        return
    } else if (this.Conditions.id) {
        this.showByID()
        return
    } else if (this.Conditions.title) {
        try {
            // FindWikiByTitleReq is the request structure of FindWikiByTitle()
            const FindWikiByTitleReq = {
                "Title": this.Conditions.title,
                "Offset": 1844674407370955,
                "Limit": 1,
            }
            const FindWikiByTitleRes = await FindWikiByTitle(FindWikiByTitleReq)
            this.Conditions.id = FindWikiByTitleRes.IDs[0]
            this.showByID()
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else if (this.Conditions.uri) {
        try {
            // FindWikiByURIReq is the request structure of FindWikiByURI()
            const FindWikiByURIReq = {
                "URI": this.Conditions.uri,
                "Offset": 1844674407370955,
                "Limit": 1,
            }
            const FindWikiByURIRes = await FindWikiByURI(FindWikiByURIReq)
            this.Conditions.id = FindWikiByURIRes.IDss[0]
            this.showByID()
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    }

    const wiki = this.NotFound()
    window.document.body.innerHTML = this.HTML(wiki)
    errors.HandleError(errors.poolByID[1685872164])
}

wikiPage.DisconnectedCallback = function () {
    this.resetCondition()
}

wikiPage.resetCondition = function () {
    this.Conditions = {
        id: "",
        title: "",
        uri: "",
        lang: "NaN",
        new: "",
    }
}

wikiPage.NotFound = function () {
    return {
        "WriteTime": 0,  // int64
        "AppInstanceID": "",
        "UserConnectionID": "",

        "OrgID": cookie.GetByName(HTTPCookieNameDelegateUserID),
        "URI": "urn:gs1:ean13:",
        "Title": "Not Found",
        "Text": "Not Found",
        "Language": users.active.ContentPreferences.Language.id,
        "Pictures": [],
    }
}

wikiPage.showByID = async function () {
    try {
        // GetWikiByIDReq is the request structure of GetWikiByID()
        const GetWikiByIDReq = {
            "ID": this.Conditions.id,
            "Language": this.Conditions.lang,
        }
        const GetWikiByIDRes = await GetWikiByID(GetWikiByIDReq)
        GetWikiByIDRes.Language = this.Conditions.lang
        window.document.body.innerHTML = this.HTML(GetWikiByIDRes)
    } catch (err) {
        errors.HandleError(err)
        return
    }
}

wikiPage.EnableNew = function () {
    const newWikiElement = document.getElementById("newWiki")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")

    const wikiTitleInputElement = document.getElementById("wikiTitleInput")
    const wikiLanguageInputElement = document.getElementById("wikiLanguageInput")
    const wikiURIInputInputElement = document.getElementById("wikiURIInput")
    // TODO::: pictures
    const wikiTextElement = document.getElementById("wikiText")
    const wikiTextInputContainerElement = document.getElementById("wikiTextInputContainer")

    newWikiElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false

    wikiTitleInputElement.hidden = false
    wikiLanguageInputElement.hidden = false
    wikiURIInputInputElement.hidden = false
    wikiTextElement.hidden = true
    wikiTextInputContainerElement.hidden = false
}

wikiPage.EnableEdit = function () {
    const editWikiElement = document.getElementById("editWiki")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")

    const wikiTitleInputElement = document.getElementById("wikiTitleInput")
    const wikiURIInputInputElement = document.getElementById("wikiURIInput")
    // TODO::: pictures
    const wikiTextElement = document.getElementById("wikiText")
    const wikiTextInputContainerElement = document.getElementById("wikiTextInputContainer")
    const wikiTextInputElement = document.getElementById("wikiTextInput")

    editWikiElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false

    wikiTitleInputElement.hidden = false
    wikiTextElement.hidden = true
    wikiURIInputInputElement.hidden = false
    wikiTextInputContainerElement.hidden = false
    wikiTextInputElement.innerText = wikiTextElement.innerText
}

wikiPage.AddNewLanguage = function () {
    const addNewLanguageElement = document.getElementById("addNewLanguage")
    const saveChangesElement = document.getElementById("saveChanges")
    const discardChangesElement = document.getElementById("discardChanges")

    const wikiTitleInputElement = document.getElementById("wikiTitleInput")
    const wikiLanguageInputElement = document.getElementById("wikiLanguageInput")
    // TODO::: pictures
    const wikiTextElement = document.getElementById("wikiText")
    const wikiTextInputContainerElement = document.getElementById("wikiTextInputContainer")
    const wikiTextInputElement = document.getElementById("wikiTextInput")

    addNewLanguageElement.disabled = true
    saveChangesElement.disabled = false
    discardChangesElement.disabled = false

    wikiTitleInputElement.hidden = false
    wikiLanguageInputElement.hidden = false
    wikiTextElement.hidden = true
    wikiTextInputContainerElement.hidden = false
    wikiTextInputElement.innerText = wikiTextElement.innerText
}

wikiPage.SaveEdit = async function () {
    const newWikiElement = document.getElementById("newWiki")
    const editWikiElement = document.getElementById("editWiki")
    const addNewLanguageElement = document.getElementById("addNewLanguage")
    // const saveChangesElement = document.getElementById("saveChanges")
    // const discardChangesElement = document.getElementById("discardChanges")

    // const wikiTitleElement = document.getElementById("wikiTitle")
    const wikiTitleInputElement = document.getElementById("wikiTitleInput")
    // const wikiLanguageElement = document.getElementById("wikiLanguage")
    const wikiLanguageInputElement = document.getElementById("wikiLanguageInput")
    // const wikiURIElement = document.getElementById("wikiURI")
    const wikiURIInputInputElement = document.getElementById("wikiURIInput")
    // TODO::: pictures
    // const wikiTextElement = document.getElementById("wikiText")
    // const wikiTextInputContainerElement = document.getElementById("wikiTextInputContainer")
    const wikiTextInputElement = document.getElementById("wikiTextInput")

    const lang = language.GetSupportedByNativeName(wikiLanguageInputElement.value)
    if (!lang) {
        // TODO::: fix error
        errors.HandleError(errors.poolByID[473425575])
        return
        return
    }

    if (newWikiElement.disabled) {
        try {
            // RegisterNewWikiReq is the request structure of RegisterNewWiki()
            const RegisterNewWikiReq = {
                "Language": lang.id,
                "URI": wikiURIInputInputElement.value,
                "Title": wikiTitleInputElement.value,
                "Text": wikiTextInputElement.value,
                // "Pictures": [],
            }
            let res = await RegisterNewWiki(RegisterNewWikiReq)
            pages.Router({}, "/wiki?id=" + res.ID + "&lang="+lang.id)
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else if (editWikiElement.disabled) {
        try {
            // UpdateWikiReq is the request structure of UpdateWiki()
            const UpdateWikiReq = {
                "ID": this.Conditions.id,
                "Language": Number(wikiLanguageInputElement.value),
                "URI": wikiURIInputInputElement.value,
                "Title": wikiTitleInputElement.value,
                "Text": wikiTextInputElement.value,
                // "Pictures": []
            }
            await UpdateWiki(UpdateWikiReq)
            pages.Router({}, "/wiki?id=" + this.Conditions.id + "&lang="+lang.id)
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else if (addNewLanguageElement.disabled) {
        try {
            // RegisterWikiNewLanguageReq is the request structure of RegisterWikiNewLanguage()
            const RegisterWikiNewLanguageReq = {
                "ID": this.Conditions.id,
                "Language": Number(wikiLanguageInputElement.value),
                "URI": wikiURIInputInputElement.value,
                "Title": wikiTitleInputElement.value,
                "Text": wikiTextInputElement.value,
                // "Pictures": []
            }
            await RegisterWikiNewLanguage(RegisterWikiNewLanguageReq)
            pages.Router({}, "/wiki?id=" + this.Conditions.id + "&lang="+lang.id)
            return
        } catch (err) {
            errors.HandleError(err)
            return
        }
    } else {
        // TODO::: warn user can't do anything!
    }
}

wikiPage.DiscardChanges = function () {
    alert("Sorry! Not implemented yet!")
}
