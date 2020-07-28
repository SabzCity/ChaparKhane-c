/* For license and copyright information please see LEGAL file in repository */

Application.Widgets["service-menu"] = {
    ID: "service-menu",
    HTML: (services) => ``,
    CSS: '',
    Templates: {
        "service": (page, checked) => ``,
    }
}

Application.Widgets["service-menu"].ConnectedCallback = function () {
    let activeMostUsedPages = Application.UserPreferences.MostUsedPages || Application.MostUsedPages
    let initServices = ""
    for (s of activeMostUsedPages) {
        let page = Application.Pages[s]
        initServices += Application.Widgets['service-menu'].Templates["service"](page, Application.ActivePage.ID === page.ID)
    }

    pageStylesElement.insertAdjacentHTML("beforeend", this.CSS)
    return this.HTML(initServices)
}

Application.Widgets["service-menu"].DisconnectedCallback = function () {
    document.getElementById("service-" + Application.ActivePage.Name).removeAttribute("checked")
}

function toggleServiceMenu() {
    if (document.getElementById("serviceMenu").open == true) {
        document.getElementById("serviceMenu").open = false
        document.getElementById("serDisabledBackground").setAttribute('hidden', '')
    } else {
        document.getElementById("serDisabledBackground").removeAttribute('hidden')
        document.getElementById("serviceMenu").open = true
        // document.getElementById("serviceMenu").showModal()
        // Input must get focus on physical keyboard || not have small screen!
        if (window.navigator.maxTouchPoints === 0) document.getElementById("findInput").focus()
    }
    // ??TODO?? reset menu order to default when closing dialog??
    // activeMostUsedPages = Application.UserPreferences.MostUsedPages || Application.MostUsedPages
}

function findServiceInServiceMenu(findInput) {
    let activeMostUsedPages
    if (findInput === "") {
        activeMostUsedPages = Application.UserPreferences.MostUsedPages || Application.MostUsedPages
    } else {
        activeMostUsedPages = Object.keys(Application.Pages)
            .filter(s => {
                let page = Application.Pages[s]
                return page.Info.Name.toLowerCase().includes(findInput.toLowerCase()) ||
                    page.Info.ShortName.toLowerCase().includes(findInput.toLowerCase())
            })
    }
    let services = ""
    for (s of activeMostUsedPages) {
        let page = Application.Pages[s]
        services += Application.Widgets['service-menu'].Templates["service"](page, Application.ActivePage.ID === page.ID)
    }
    document.getElementById("servicesContainer").innerHTML = services
}
