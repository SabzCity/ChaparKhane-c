/* For license and copyright information please see LEGAL file in repository */

import './libjs/application.js'
import './libjs/polyfill.js'
import './libjs/widget-notification/center.js'
import './libjs/widget-notification/pop-up.js'

// function init() {
Application.Icon = "app-icon-512x512.png"
Application.ContentPreferences = {
    Languages: ["fa", "en"],
    Regions: ["IRN"],
    Currencies: ["IRR"],
}
Application.PresentationPreferences = {
    DesignLanguage: "material",
    ColorScheme: "no-preference",
    ThemeColor: "#66ff55",
    PrimaryFontFamily: "Roboto",
    Display: "standalone",
    Orientation: "portrait",
}
Application.HomePage = "store" // start with store page!
Application.MostUsedPages = ["person", "orgs", "finance", "store", "wikis"]
// }
// Call init function to application work on not supported browsers!! that now there is no browser!! ;)
// init()

// function main() {
Application.Start()

// centerNotificationWidget.ConnectedCallback({})
popUpNotificationWidget.ConnectedCallback({})
Polyfill.SuggestSupportedBrowser()
// }
// Call main function to application work on not supported browsers!! that now there is no browser!! ;)
// main()
