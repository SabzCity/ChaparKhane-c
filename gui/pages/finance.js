/* For license and copyright information please see LEGAL file in repository */

Application.Pages["finance"] = {
    ID: "finance",
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
    Icon: "account_balance",
    Related: [], // "invoices", "wallet", "stock-exchange", "foreign-exchange"
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["finance"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["finance"].DisconnectedCallback = function () {
}
