/* For license and copyright information please see LEGAL file in repository */

Application.Pages[""] = {
    ID: "",
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
    Icon: "home",
    Related: ["login", "register"],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages[""].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    // Application.Widgets["hamburger-menu"].ConnectedCallback("leftTopHeader")
    // Application.Widgets["user-menu"].ConnectedCallback("rightTopHeader")
    // Application.Widgets["service-menu"].ConnectedCallback("rightTopHeader")
}

Application.Pages[""].DisconnectedCallback = function () {
}
