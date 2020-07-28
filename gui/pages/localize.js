/* For license and copyright information please see LEGAL file in repository */

import '../libjs/widget-localize/widget-content-preferences.js'
import '../libjs/widget-localize/widget-presentation-preferences.js'

Application.Pages["localize"] = {
    ID: "localize",
    RecordID: null,
    Condition: {},
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
    Icon: "language",
    Related: [],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["localize"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["localize"].DisconnectedCallback = function () {
}

function localizePageSave() {
    // Save user choose in related SabzCity API
    // SetUserInfo(UserState.ID) = Application.UserPreferences.ContentPreferences.Language.iso639_1 + "-" + Application.UserPreferences.ContentPreferences.Region.iso3166_1_a2

    if (Application.PreviousPage && Application.PreviousPage.ActiveURI) {
        Application.PreviousPage.ActiveURI.searchParams.set('hl', Application.UserPreferences.ContentPreferences.Language.iso639_1 + "-" + Application.UserPreferences.ContentPreferences.Region.iso3166_1_a2)
        window.location.replace(Application.PreviousPage.ActiveURI)
    } else {
        window.location.replace("/?hl=" + Application.UserPreferences.ContentPreferences.Language.iso639_1 + "-" + Application.UserPreferences.ContentPreferences.Region.iso3166_1_a2)
    }

    // TODO!!??
    // Router(url)
}
