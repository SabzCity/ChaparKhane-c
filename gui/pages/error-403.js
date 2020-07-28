/* For license and copyright information please see LEGAL file in repository */

Application.Pages["error-403"] = {
    ID: "error-403",
    RecordID: "",
    Condition: {},
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
    Icon: "error",
    Related: [""],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["error-403"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["error-403"].DisconnectedCallback = function () {
}
