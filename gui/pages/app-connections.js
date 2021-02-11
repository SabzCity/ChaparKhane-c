/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/error/error.js'
import '../libjs/sdk/sabz.city/get-user-app-connection.js'
import '../libjs/sdk/sabz.city/find-user-app-connection.js'
import '../libjs/sdk/sabz.city/find-user-app-connection-by-given-delegate.js'
import '../libjs/sdk/sabz.city/find-user-app-connection-by-gotten-delegate.js'

// Platform have multi user features! It is very easy just change active userID & UserToken state!
// Org membership show here!

const appConnectionsPage = {
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
    Related: ["login", "register", "person"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "short-detail": (conn) => ``,
    },
}
pages.RegisterPage(appConnectionsPage)

appConnectionsPage.ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    // widgets["hamburger-menu"].ConnectedCallback("leftTopHeader")
    // widgets["user-menu"].ConnectedCallback("rightTopHeader")
    // serviceMenuWidget.ConnectedCallback("rightTopHeader")

    appConnectionsPage.SetConnections(1844674407370955161, 10)
    appConnectionsPage.SetGottenDelegate(1844674407370955161, 100)
    appConnectionsPage.SetGivenDelegate(1844674407370955161, 10)
}

appConnectionsPage.DisconnectedCallback = function () {
}

appConnectionsPage.SetConnections = async function (offset, limit) {
    const userConnections = window.document.getElementById("user-connections")

    try {
        const FindUserAppConnectionReq = {
            "Offset": offset,
            "Limit": limit,
        }
        const res = await FindUserAppConnection(FindUserAppConnectionReq)
        for (let ID of res.IDs) {
            try {
                const req = {
                    "ID": ID,
                }
                const conn = await GetUserAppConnection(req)
                conn.UserName = users.active.UserName
                userConnections.insertAdjacentHTML('afterbegin', this.Templates["short-detail"](conn))
            } catch (err) {
                return PersiaError.NotifyError(err)
            }
        }
    } catch (err) {
        return PersiaError.NotifyError(err)
    }
}

appConnectionsPage.SetGottenDelegate = async function (offset, limit) {
    const gottenDelegate = window.document.getElementById("gotten-delegate")

    try {
        const req = {
            "Offset": offset,
            "Limit": limit,
        }
        const res = await FindUserAppConnectionByGottenDelegate(req)
        for (let ID of res.IDs) {
            try {
                const req = {
                    "ID": ID,
                }
                const conn = await GetUserAppConnection(req)

                try {
                    const GetOrganizationReq = {
                        "ID": conn.UserID,
                    }
                    const org = await GetOrganization(GetOrganizationReq)

                    try {
                        // GetQuiddityReq is the request structure of GetQuiddity()
                        const GetQuiddityReq = {
                            "ID": org.QuiddityID,
                            "Language": users.active.ContentPreferences.Language.ID,
                        }
                        const GetQuiddityRes = await GetQuiddity(GetQuiddityReq)

                        conn.UserName = GetQuiddityRes.Title
                        gottenDelegate.insertAdjacentHTML('afterbegin', this.Templates["short-detail"](conn))
                    } catch (err) {
                        PersiaError.NotifyError(err)
                        continue
                    }
                } catch (err) {
                    PersiaError.NotifyError(err)
                    continue
                }
            } catch (err) {
                PersiaError.NotifyError(err)
                continue
            }
        }
    } catch (err) {
        return PersiaError.NotifyError(err)
    }
}

appConnectionsPage.SetGivenDelegate = async function (offset, limit) {
    const givenDelegate = window.document.getElementById("given-delegate")

    try {
        const req = {
            "Offset": offset,
            "Limit": limit,
        }
        const res = await FindUserAppConnectionByGivenDelegate(req)
        for (let ID of res.IDs) {
            let conn = {
                ID: ID,
            }
            givenDelegate.insertAdjacentHTML('afterbegin', this.Templates["short-detail"](conn))
        }
    } catch (err) {
        return PersiaError.NotifyError(err)
    }
}