/* For license and copyright information please see LEGAL file in repository */

import '../../sdk-js/get-user-app-connections-id.js'
import '../../sdk-js/get-user-app-given-delegate-connections-id.js'
import '../../sdk-js/get-user-app-gotten-delegate-connections-id.js'

// Platform have multi user features! It is very easy just change active userID & UserToken state!
// Org membership show here!

Application.Pages["app-connections"] = {
    ID: "app-connections",
    Conditions: {
        id: "",
    },
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
    Icon: "sync",
    Related: ["register", "my"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "short-detail": (conn) => ``,
    },
}

Application.Pages["app-connections"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    // widgets["hamburger-menu"].ConnectedCallback("leftTopHeader")
    // widgets["user-menu"].ConnectedCallback("rightTopHeader")
    // widgets["service-menu"].ConnectedCallback("rightTopHeader")

    appConnectionsPageSetConnections(1844674407370955161, 10)
    appConnectionsPageSetGottenDelegate(1844674407370955161, 10)
    appConnectionsPageSetGivenDelegate(1844674407370955161, 10)
}

Application.Pages["app-connections"].DisconnectedCallback = function () {
}

async function appConnectionsPageSetConnections(offset, limit) {
    const userConnections = window.document.getElementById("user-connections")

    try {
        const GetUserAppConnectionsIDReq = {
            "Offset": offset,
            "Limit": limit,
        }
        const res = await GetUserAppConnectionsID(GetUserAppConnectionsIDReq)
        for (let ID of res.IDs) {
            let conn = {
                ID: ID,
            }
            userConnections.insertAdjacentHTML('afterbegin', Application.Pages["app-connections"].Templates["short-detail"](conn))
        }
    } catch (err) {
        switch (err) {
            case 2605698703: // Authorization - User Not Allow
            case 1685872164: // Ganjine - Record Not Found
        }
    }
}

async function appConnectionsPageSetGottenDelegate(offset, limit) {
    const gottenDelegate = window.document.getElementById("gotten-delegate")

    try {
        const req = {
            "Offset": offset,
            "Limit": limit,
        }
        const res = await GetUserAppGottenDelegateConnectionsID(req)
        for (let ID of res.IDs) {
            let conn = {
                ID: ID,
            }
            gottenDelegate.insertAdjacentHTML('afterbegin', Application.Pages["app-connections"].Templates["short-detail"](conn))
        }
    } catch (err) {
        switch (err) {
            case 2605698703: // Authorization - User Not Allow
            case 1685872164: // Ganjine - Record Not Found
        }
    }
}

async function appConnectionsPageSetGivenDelegate(offset, limit) {
    const givenDelegate = window.document.getElementById("given-delegate")

    try {
        const req = {
            "Offset": offset,
            "Limit": limit,
        }
        const res = await GetUserAppGivenDelegateConnectionsID(req)
        for (let ID of res.IDs) {
            let conn = {
                ID: ID,
            }
            givenDelegate.insertAdjacentHTML('afterbegin', Application.Pages["app-connections"].Templates["short-detail"](conn))
        }
    } catch (err) {
        switch (err) {
            case 2605698703: // Authorization - User Not Allow
            case 1685872164: // Ganjine - Record Not Found
        }
    }
}