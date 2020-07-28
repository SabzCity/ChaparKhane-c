/* For license and copyright information please see LEGAL file in repository */

import { GetUserData } from '../../sdk-js/get-users-data.js'

Application.Widgets["user-menu"] = {
    ID: "user-menu",
    HTML: (state) => ``,
    CSS: '',
    Templates: {}
}

Application.Widgets["user-menu"].ConnectedCallback = function () {
    let userMenuState
    if (Application.UserPreferences.UsersState.ActiveUserID !== "") {
        userMenuState.aHref = "/sessions"
        userMenuState.aTitle = "LocaleText[1]"
        let userData = GetUserData(Application.UserPreferences.UsersState.ActiveUserID)
        if (userData.Picture !== "") userMenuState.imgSrc = userData.Picture
    } else {
        userMenuState = {
            aHref: "/login",
            aTitle: "LocaleText[0]",
            imgSrc: "/images/not-login-user.svg",
        }
    }
    pageStylesElement.insertAdjacentHTML("beforeend", this.CSS)
    return this.HTML(userMenuState)
}

Application.Widgets["user-menu"].DisconnectedCallback = function () {
}
