/* For license and copyright information please see LEGAL file in repository */

Application.Widgets["hamburger-menu"] = {
    ID: "hamburger-menu",
    HTML: () => ``,
    CSS: '',
    Templates: {}
}

Application.Widgets["hamburger-menu"].ConnectedCallback = function () {
    pageStylesElement.insertAdjacentHTML("beforeend", this.CSS)
    return this.HTML()
}

Application.Widgets["hamburger-menu"].DisconnectedCallback = function () {
}

function toggleHamMenu() {
    if (document.getElementById("hamMenu").open == true) {
        document.getElementById("hamMenu").open = false
        document.getElementById("hamDisabledBackground").setAttribute('hidden', '')
    } else {
        document.getElementById("hamDisabledBackground").removeAttribute('hidden')
        document.getElementById("hamMenu").open = true
        // document.getElementById("hamMenu").showModal()
    }
}
