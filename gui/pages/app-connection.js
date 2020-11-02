/* For license and copyright information please see LEGAL file in repository */

import '../libjs/connection.js'
import '../libjs/base64.js'
import '../libjs/ip.js'
import '../libjs/time.js'
import '../libjs/authorization.js'
import '../libjs/connection.js'
import '../libjs/services.js'
import '../../sdk-js/get-user-app-connection.js'
import '../../sdk-js/register-user-app-connection.js'

Application.Pages["app-connection"] = {
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
    Related: [],
    HTML: (conn) => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["app-connection"].ConnectedCallback = async function () {
    let conn = {
        Status: UserAppsConnectionIssued,
        AccessControl: {}
    }

    if (Application.ActivePage.Conditions.new === "true") {
        appConnectionPageEnableNew()
        return
    }

    if (!Application.ActivePage.Conditions.id) {
        // TODO:::
    } else {
        try {
            const req = {
                "ID": Application.ActivePage.Conditions.id,
            }
            conn = await GetUserAppConnection(req)
        } catch (err) {
            switch (err) {
                case 2605698703: // Authorization - User Not Allow
                case 1685872164: // Ganjine - Record Not Found
            }
        }
    }

    document.body.innerHTML = this.HTML(conn)

    if (Application.ActivePage.Conditions.edit === "true") {
        appConnectionPageEnableEdit()
    }

    // widgets["hamburger-menu"].ConnectedCallback()
    // widgets["user-menu"].ConnectedCallback()
    // widgets["service-menu"].ConnectedCallback()
}

Application.Pages["app-connection"].DisconnectedCallback = function () {
    const saveChangesButtonElement = document.getElementById("saveChanges")
    if (saveChangesButtonElement.disabled === false) {
        alert("Changes not saved yet! All changes will lost if you leave page")
        return false
    }
}

function appConnectionPageEnableNew() {
    let conn = {
        Status: UserAppsConnectionIssued,
        Description: "",
        Weight: 0,
        ThingID: "",
        DelegateUserID: "",
        DelegateUserType: 0,
        PeerPublicKey: "",
        AccessControl: {}
    }
    document.body.innerHTML = Application.Pages["app-connection"].HTML(conn)

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

function appConnectionPageEnableEdit() {
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

async function appConnectionPageSaveEdit() {
    const saveChangesButtonElement = document.getElementById("saveChanges")
    saveChangesButtonElement.disabled = true
    const discardChangesButtonElement = document.getElementById("discardChanges")
    discardChangesButtonElement.disabled = true

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
        delegateUserTypeValueElement.innerHTML = connection.type.GetNameByID(delegateUserTypeSelectElement.value)
    
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
            switch (err) {
                case 2605698703: // Authorization - User Not Allow
                alert("Authorization - User Not Allow")
                case 2605698703: // Authorization - Not Allow To Delegate
                alert("Authorization - Not Allow To Delegate")
                case 1685872164: // Ganjine - Record Not Found
                case 1657764712: // JSON - Encoded String Corrupted
            }
        }
    } else {
        const statusSelectElement = document.getElementById('statusSelect')
        statusSelectElement.hidden = true
        const statusValueElement = document.getElementById('statusValue')
        statusValueElement.hidden = false
        statusValueElement.innerHTML = connection.status.GetNameByID(statusSelectElement.value)

        alert("Sorry! Not implemented yet!")
    }
}

function appConnectionPageDiscardChanges() {
    // TODO::: warn user first

    const saveChangesButtonElement = document.getElementById("saveChanges")
    saveChangesButtonElement.disabled = true
    const discardChangesButtonElement = document.getElementById("discardChanges")
    discardChangesButtonElement.disabled = true
}

function appConnectionPageEnableExpireTheConnection() {
    alert("Sorry! Not implemented yet!")
}

function appConnectionPageEnableRevokeTheConnection() {
    alert("Sorry! Not implemented yet!")
}
