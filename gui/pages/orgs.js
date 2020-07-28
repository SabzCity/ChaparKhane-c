/* For license and copyright information please see LEGAL file in repository */

Application.Pages["orgs"] = {
    ID: "orgs",
    RecordID: "",
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
    Icon: "business",
    Related: ["login"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "org": () => ``,
        "user-org": () => ``,
    },
}

Application.Pages["orgs"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["orgs"].DisconnectedCallback = function () {
}

Application.Pages["orgs"].addUserOrgs = function (query) {
}

Application.Pages["orgs"].addUserTimeLine = function () {
}

Application.Pages["orgs"].toggleRegisterOrgDialog = function () {
    if (!Application.UserPreferences.UsersState.ActiveUserID) {
        alert("Login first")
        return
    }
    const registerOrgDialog = window.document.getElementById("registerOrgDialog")
    registerOrgDialog.showModal()
}

Application.Pages["orgs"].registerOrg = function () {
    const registerOrgDialog = window.document.getElementById("registerOrgDialog")
    registerOrgDialog.close()
}

Application.Pages["orgs"].cancelRegisterOrg = function () {
    const registerOrgDialog = window.document.getElementById("registerOrgDialog")
    registerOrgDialog.close()
}


Application.Pages["orgs"].TestData = {
    userJoinedOrgs: ["1", "2"],
    orgs: {
        "1": {
            OrganizationID: "1",
            StartDate: "2019/05/12",
            RegisterDate: "2019/05/12",
        },
        "2": {
            OrganizationID: "2",
            StartDate: "2019/08/25",
            RegisterDate: "2019/08/25",

        }
    },
}