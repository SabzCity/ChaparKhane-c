/* For license and copyright information please see LEGAL file in repository */

Application.Widgets["product-search"] = {
    ID: "product-search",
    HTML: () => ``,
    CSS: '',
    Templates: {}
}

Application.Widgets["product-search"].ConnectedCallback = function () {
    pageStylesElement.insertAdjacentHTML("beforeend", this.CSS)
    return this.HTML()
}

Application.Widgets["product-search"].DisconnectedCallback = function () {
}

function toggleSearchDialog() {
    if (document.getElementById("searchDialog").open == true) {
        document.getElementById("searchDialog").open = false
        document.getElementById("searchDisabledBackground").setAttribute('hidden', '')
    } else {
        document.getElementById("searchDisabledBackground").removeAttribute('hidden')
        document.getElementById("searchDialog").open = true
        // document.getElementById("searchDialog").showModal()
        // Input must get focus on physical keyboard || not have small screen!
        if (window.navigator.maxTouchPoints === 0) document.getElementById("textSearch").focus()
    }
}

function applySearchFilter() {
    this.toggleSearchDialog()

    let url = '/store?'
    const q = this.shadowRoot.getElementById("textSearch").value
    if (q) url = url + 'q=' + q + '&'
    const tags = this.shadowRoot.getElementById("tags").value
    if (tags) url = url + 'tags=' + tags + '&'
    const sort = this.shadowRoot.getElementById("sort").value
    if (sort) url = url + 'sort=' + sort + '&'

    if (q || tags || sort) {
        history.pushState(history.state, "", url)
        window.dispatchEvent(new Event('pushState'))
    }
}

function clearSearchFilter() {
    this.toggleSearchDialog()
    this.shadowRoot.getElementById("textSearch").value = ""
    history.pushState(history.state, "", '/store')
    window.dispatchEvent(new Event('pushState'))
}
