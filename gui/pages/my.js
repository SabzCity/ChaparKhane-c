/* For license and copyright information please see LEGAL file in repository */

Application.Pages["my"] = {
    ID: "my",
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
    Icon: "person",
    Related: ["aboutme", "security"], // "groups", "sessions", "settings"
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["my"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["my"].DisconnectedCallback = function () {
}
