/* For license and copyright information please see LEGAL file in repository */

import '../libjs/base64.js'
import '../libjs/ip.js'
import '../libjs/authorization/authorization.js'
import '../libjs/authorization/user.js'
import '../libjs/services.js'
import '../libjs/error/error.js'
import '../libjs/cookie.js'
import '../libjs/widget-notification/center.js'
import '../libjs/widget-notification/pop-up.js'
import '../libjs/sdk/sabz.city/get-user-app-connection.js'
import '../libjs/sdk/sabz.city/register-user-app-connection.js'
import '../libjs/sdk/sabz.city/datastore-user-apps-connection.js'

const appConnectionPage = {
    ID: "app-connection",
    Conditions: {
        id: "",
        edit: "false",
        new: "false",
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
    Icon: "sync_alt",
    Related: ["login", "person"],
    HTML: (conn) => ``,
    CSS: '',
    Templates: {},
}
pages.RegisterPage(appConnectionPage)

appConnectionPage.ConnectedCallback = async function () {
    let conn = {
        Status: UserAppConnectionIssued,
        AccessControl: {}
    }

    if (pages.ActivePage.Conditions.new === "true") {
        appConnectionPageEnableNew()
        return
    }

    if (!pages.ActivePage.Conditions.id) {
        // TODO:::
    } else {
        try {
            const req = {
                "ID": pages.ActivePage.Conditions.id,
            }
            conn = await GetUserAppConnection(req)
        } catch (err) {
            return PersiaError.NotifyError(err)
        }
    }

    document.body.innerHTML = this.HTML(conn)

    if (pages.ActivePage.Conditions.edit === "true") {
        appConnectionPageEnableEdit()
    }

    // widgets["hamburger-menu"].ConnectedCallback()
    // widgets["user-menu"].ConnectedCallback()
    // serviceMenuWidget.ConnectedCallback()
}

appConnectionPage.DisconnectedCallback = function () {
    const saveChangesButtonElement = document.getElementById("saveChanges")
    if (saveChangesButtonElement && saveChangesButtonElement.disabled === false) {
        alert("Changes not saved yet! All changes will lost if you leave page")
        return false
    }
}

appConnectionPage.EnableNew = function () {
    let conn = {
        Status: UserAppConnectionIssued,
        Description: "",
        Weight: 0,
        ThingID: "",
        DelegateUserID: "",
        DelegateUserType: 0,
        PeerPublicKey: "",
        AccessControl: {}
    }
    document.body.innerHTML = this.HTML(conn)

    const newConnElement = document.getElementById("newConn")
    newConnElement.disabled = true
    const editConnElement = document.getElementById("editConn")
    editConnElement.disabled = true
    const saveChangesButtonElement = document.getElementById("saveChanges")
    saveChangesButtonElement.disabled = false
    const discardChangesButtonElement = document.getElementById("discardChanges")
    discardChangesButtonElement.disabled = false

    const descriptionValueElement = document.getElementById('descriptionValue')
    descriptionValueElement.hidden = true
    const descriptionInputElement = document.getElementById('descriptionInput')
    descriptionInputElement.hidden = false

    const weightInputElement = document.getElementById('weightInput')
    weightInputElement.hidden = false

    const thingIDValueElement = document.getElementById('thingIDValue')
    thingIDValueElement.hidden = true
    const thingIDInputElement = document.getElementById('thingIDInput')
    thingIDInputElement.hidden = false

    const delegateUserIDValueElement = document.getElementById('delegateUserIDValue')
    delegateUserIDValueElement.hidden = true
    const delegateUserIDInputElement = document.getElementById('delegateUserIDInput')
    delegateUserIDInputElement.hidden = false

    const delegateUserTypeValueElement = document.getElementById('delegateUserTypeValue')
    delegateUserTypeValueElement.hidden = true
    const delegateUserTypeSelectElement = document.getElementById('delegateUserTypeSelect')
    delegateUserTypeSelectElement.hidden = false

    const publicKeyValueElement = document.getElementById('publicKeyValue')
    publicKeyValueElement.hidden = true
    const publicKeyInputElement = document.getElementById('publicKeyInput')
    publicKeyInputElement.hidden = false

    // TODO::: AccessControl
}

appConnectionPage.EnableEdit = function () {
    const editConnElement = document.getElementById("editConn")
    editConnElement.disabled = true
    const saveChangesButtonElement = document.getElementById("saveChanges")
    saveChangesButtonElement.disabled = false
    const discardChangesButtonElement = document.getElementById("discardChanges")
    discardChangesButtonElement.disabled = false

    const statusValueElement = document.getElementById('statusValue')
    statusValueElement.hidden = true
    const statusSelectElement = document.getElementById('statusSelect')
    statusSelectElement.hidden = false

    const descriptionValueElement = document.getElementById('descriptionValue')
    descriptionValueElement.hidden = true
    const descriptionInputElement = document.getElementById('descriptionInput')
    descriptionInputElement.hidden = false

    const weightInputElement = document.getElementById('weightInput')
    weightInputElement.hidden = false

    const publicKeyValueElement = document.getElementById('publicKeyValue')
    publicKeyValueElement.hidden = true
    const publicKeyInputElement = document.getElementById('publicKeyInput')
    publicKeyInputElement.hidden = false

    // TODO::: AccessControl
}

appConnectionPage.SaveEdit = async function () {
    const descriptionInputElement = document.getElementById('descriptionInput')
    descriptionInputElement.hidden = true
    const descriptionValueElement = document.getElementById('descriptionValue')
    descriptionValueElement.hidden = false
    descriptionValueElement.innerHTML = descriptionInputElement.value

    const weightInputElement = document.getElementById('weightInput')
    weightInputElement.hidden = true

    const publicKeyInputElement = document.getElementById('publicKeyInput')
    publicKeyInputElement.hidden = true
    const publicKeyValueElement = document.getElementById('publicKeyValue')
    publicKeyValueElement.hidden = false
    publicKeyValueElement.innerHTML = publicKeyInputElement.value

    // TODO::: AccessControl

    const newConnElement = document.getElementById("newConn")
    if (newConnElement.disabled === true) {
        const thingIDInputElement = document.getElementById('thingIDInput')
        thingIDInputElement.hidden = true
        const thingIDValueElement = document.getElementById('thingIDValue')
        thingIDValueElement.hidden = false
        thingIDValueElement.innerHTML = thingIDInputElement.value

        const delegateUserIDInputElement = document.getElementById('delegateUserIDInput')
        delegateUserIDInputElement.hidden = true
        const delegateUserIDValueElement = document.getElementById('delegateUserIDValue')
        delegateUserIDValueElement.hidden = false
        delegateUserIDValueElement.innerHTML = delegateUserIDInputElement.value

        const delegateUserTypeSelectElement = document.getElementById('delegateUserTypeSelect')
        delegateUserTypeSelectElement.hidden = true
        const delegateUserTypeValueElement = document.getElementById('delegateUserTypeValue')
        delegateUserTypeValueElement.hidden = false
        delegateUserTypeValueElement.innerHTML = authorization.UserType.GetDetailsByID(delegateUserTypeSelectElement.value)

        try {
            const req = {
                "Description": descriptionInputElement.value,
                "Weight": Number(weightInputElement.value),

                "ThingID": thingIDInputElement.value,
                "DelegateUserID": delegateUserIDInputElement.value,
                "DelegateUserType": Number(delegateUserTypeSelectElement.value),

                "PublicKey": publicKeyInputElement.value,
                "AccessControl": {}  // authorization.AccessControl
            }
            res = await RegisterUserAppConnection(req)
            document.getElementById('connID').innerHTML = res.ID
        } catch (err) {
            return PersiaError.NotifyError(err)
        }

        newConnElement.disabled = false
    } else {
        const statusSelectElement = document.getElementById('statusSelect')
        statusSelectElement.hidden = true
        const statusValueElement = document.getElementById('statusValue')
        statusValueElement.hidden = false
        statusValueElement.innerHTML = UserAppConnectionStatus.GetShortDetailByID(statusSelectElement.value)

        alert("Sorry! Not implemented yet!")

        const editConnElement = document.getElementById("editConn")
        editConnElement.disabled = false
    }

    const saveChangesButtonElement = document.getElementById("saveChanges")
    saveChangesButtonElement.disabled = true
    const discardChangesButtonElement = document.getElementById("discardChanges")
    discardChangesButtonElement.disabled = true
}

appConnectionPage.DiscardChanges = function () {
    // TODO::: warn user first

    const saveChangesButtonElement = document.getElementById("saveChanges")
    saveChangesButtonElement.disabled = true
    const discardChangesButtonElement = document.getElementById("discardChanges")
    discardChangesButtonElement.disabled = true
}

appConnectionPage.EnableExpireTheConnection = function () {
    alert("Sorry! Not implemented yet!")
}

appConnectionPage.EnableRevokeTheConnection = function () {
    alert("Sorry! Not implemented yet!")
}

appConnectionPage.ActiveOrgConnection = function (personID, OrgID, UserType) {
    if (UserType !== UserTypeOrg) {
        // TODO::: remove below and active person gotten delegate connection. don't let active given delegate!
        PersiaError.NotifyError(PersiaError.GetByID(2657416029)) // "Not Allow To Delegate"
        return
    }

    users.ChangeActiveUser(OrgID, personID)

    cookie.SetCookie({
        Name: HTTPCookieNameDelegateUserID,
        Value: OrgID,
        MaxAge: "630720000",
        // Secure: true,
    })

    cookie.SetCookie({
        Name: HTTPCookieNameDelegateConnID,
        Value: personID,
        MaxAge: "630720000",
        // Secure: true,
    })

    popUpNotificationWidget.New("LocaleText[54]", "LocaleText[55]", "Success")
    // centerNotificationWidget.New("LocaleText[54]", "LocaleText[55]", "Success")
}
