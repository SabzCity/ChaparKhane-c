/* For license and copyright information please see LEGAL file in repository */

import '../widgets/widget-product-search.js'

Application.Pages["store"] = {
    ID: "store",
    RecordID: null,
    Condition: {
        "q": "", // query
        "tags": [],
        "sort": "",
        "orgID": "",
        "wareHouseDistance": 0, // 0 means just near available product
        "lat": 0.0, // latitude
        "long": 0.0, // longitude
        "ownOrder": false, // all products user buy before
    },
    State: "",
    Robots: "noindex, follow, noarchive, noimageindex",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "store", Related: [], // "invoices", "order-list", "compare-products"
    HTML: () => ``,
    CSS: '',
    Templates: {
        "product": (p, w) => ``,
    },
    Widgets: ["hamburger-menu", "user-menu", "service-menu", "product-search"],
}

Application.Pages["store"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    const productsMain = window.document.getElementById("productsMain")
    let products = []
    if (Application.Pages["store"].Condition["q"] !== "" ||
        Application.Pages["store"].Condition["tags"].length !== 0 ||
        Application.Pages["store"].Condition["sort"] !== "") {
        products = ["12345"]
    } else {
        // TODO : Persistence state in time line even if route occur!
        products = ["12345", "5453", "5454", "8547", "8889"]
    }
    for (let id of products) {
        productsMain.insertAdjacentHTML('beforeend', Application.Pages["store"].getProduct(id))
    }
}

Application.Pages["store"].DisconnectedCallback = function () {
    // reset last page state and conditions
    Application.Pages["store"].Condition["q"] = ""
    Application.Pages["store"].Condition["tags"] = []
    Application.Pages["store"].Condition["sort"] = ""
}

/**
 * getProduct use to get product detail and return details in html object
 * @param {string} uuid Product UUID
 */
Application.Pages["store"].getProduct = function (uuid) {
    // Get product details by given ID
    const w = Application.Pages["store"].TestData.wiki[uuid]
    const p = Application.Pages["store"].TestData.product[w.ID]
    if (p && w) return Application.Pages["store"].Templates["product"](p, w)
}

Application.Pages["store"].TestData = {
    wiki: {
        "12345": {
            ID: "12345",
            Name: "Where the Crawdads Sing",
            Pictures: [
                "https://images-na.ssl-images-amazon.com/images/I/81WWiiLgEyL._AC_UL480_SR318,480_.jpg"
            ],
            Tags: ["Book", "Novel"],
        },
        "5453": {
            ID: "5453",
            Name: "Fire TV Stick 4K with Alexa Voice Remote, streaming media player",
            Pictures: [
                "https://images-na.ssl-images-amazon.com/images/I/51CgKGfMelL._AC_UL320_SR320,320_.jpg"
            ],
            Tags: ["TV", "TVBox"],
        },
        "5454": {
            ID: "5454",
            Name: 'Intex River Run I Sport Lounge, Inflatable Water Float, 53" Diameter',
            Pictures: [
                "https://images-na.ssl-images-amazon.com/images/I/61KBtaWa%2B-L._AC_UL320_SR320,320_.jpg"
            ],
            Tags: ["River", "Swim"],
        },
        "8547": {
            ID: "8547",
            Name: "KORSIS Women's Summer Casual T Shirt Dresses Short Sleeve Swing Dress Pockets",
            Pictures: [
                "https://images-na.ssl-images-amazon.com/images/I/51iKmLOFhjL._AC_UL320_SR258,320_.jpg"
            ],
            Tags: ["Woman", "Shirt"],
        },
        "8889": {
            ID: "8889",
            Name: 'Womens and Mens Kids Water Shoes Barefoot Quick-Dry Aqua Socks for Beach Swim Surf Yoga Exercise',
            Pictures: [
                "https://images-na.ssl-images-amazon.com/images/I/71pVY69VM0L._AC_UL320_SR320,320_.jpg"
            ],
            Tags: ["Woman", "Shoes"],
        },
    },
    product: {
        "12345": {
            ProductID: "12345",
            OrganizationID: "",
            WikiID: "12345",
            WarehouseID: "",
            Currency: 0,
            RealPrice: 14,
            DiscountPercent: 10,
            PayablePrice: 12.6,
            TTL: "",
        },
        "5453": {
            ProductID: "5453",
            OrganizationID: "",
            WikiID: "5453",
            WarehouseID: "",
            Currency: 0,
            RealPrice: 49.99,
            DiscountPercent: 10,
            PayablePrice: 44.99,
            TTL: "",
        },
        "5454": {
            ProductID: "5454",
            OrganizationID: "",
            WikiID: "5454",
            WarehouseID: "",
            Currency: 0,
            RealPrice: 22.99,
            DiscountPercent: 0,
            PayablePrice: 22.99,
            TTL: "",
        },
        "8547": {
            ProductID: "8547",
            OrganizationID: "",
            WikiID: "8547",
            WarehouseID: "",
            Currency: 0,
            RealPrice: 25.99,
            DiscountPercent: 0,
            PayablePrice: 25.99,
            TTL: "",
        },
        "8889": {
            ProductID: "8889",
            OrganizationID: "",
            WikiID: "8889",
            WarehouseID: "",
            Currency: 0,
            RealPrice: 13.58,
            DiscountPercent: 0,
            PayablePrice: 13.58,
            TTL: "",
        },
    }
}
