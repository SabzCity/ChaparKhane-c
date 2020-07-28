/* For license and copyright information please see LEGAL file in repository */

Application.Pages["product"] = {
    ID: "product",
    RecordID: "",
    Condition: {},
    State: "",
    Robots: "all, index, follow",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "shopping_cart",
    Related: ["store", ], // "invoices", "order-list", "compare-products"
    HTML: (w, p) => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["product"].ConnectedCallback = function () {
    // Get product details
    const w = Application.Pages["store"].TestData.wiki[Application.Pages["product"].RecordID]
    if (!w) {
        Application.Router("error-404", "")
        return
    }
    const p = Application.Pages["store"].TestData.product[w.ID]

    window.document.body.innerHTML = this.HTML(w, p)
    window.document.title += " - " + w.Name

    // get user liked product before and set it to page
    // checked product add to cart before and set it to page

    Application.Pages["product"].getSuggestedProducts(Application.Pages["product"].RecordID)
}

Application.Pages["product"].DisconnectedCallback = function () {
}

Application.Pages["product"].getSuggestedProducts = function (UUID) {
    let suggestedProductsID = ["12345", "5453", "5454", "8547", "8889"]
    const suggestedProducts = window.document.getElementById("suggestedProducts")
    // listen to scrolling and if user go to end of page load related products and load until user requested
    for (let id of suggestedProductsID) {
        suggestedProducts.insertAdjacentHTML('beforeend', Application.Pages["store"].getProduct(id))
    }
}

Application.Pages["product"].addToCard = function (ID) {
}

Application.Pages["product"].toggleLikeProduct = function () {
    // first send like request to server
    const likeButton = window.document.getElementById("likeButton")
    likeButton.setAttribute("checked", "")
    // in server request error case we must revert button checked attribute again
}
