/* For license and copyright information please see LEGAL file in repository */

import './libjs/gui-engine/application.js'

// function init() {

// }
// Call init function to application work on not supported browsers!! that now there is no browser!! ;)
// init()

// function main() {
const lang = Application.UserPreferences.ContentPreferences.Language.iso639_1 || Polyfill.GetLangRegFromURL().l || "en"

const initScript = document.createElement('script')
initScript.src = "/init-" + lang + ".js"
document.head.appendChild(initScript)
initScript.onload = function () {
    // First check user preference in PWA version
    if (!window.matchMedia('(display-mode: browser)').matches && window.location.pathname === "/") {
        window.history.replaceState({}, "", "/" + Application.UserPreferences.HomePage)
        Application.Router(Application.UserPreferences.HomePage, "")
    } else {
        // Do normal routing!
        Application.Router("", window.location.href)
    }
}
// }
// Call main function to application work on not supported browsers!! that now there is no browser!! ;)
// main()
