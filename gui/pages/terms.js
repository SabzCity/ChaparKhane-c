/* For license and copyright information please see LEGAL file in repository */

Application.Pages["terms"] = {
    ID: "terms",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "",
    Related: [],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["terms"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["terms"].DisconnectedCallback = function () {
}
