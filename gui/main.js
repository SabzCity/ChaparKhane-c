/* For license and copyright information please see LEGAL file in repository */

import './libjs/application.js'

// function init() {
Application.Icon = "app-icon-512x512.png"
Application.ContentPreferences = {
    Languages: ["en", "fa"],
    Language: {
        englishName: "English",
        nativeName: "English",
        iso639_1: "en",
        iso639_2T: "eng",
        iso639_2B: "eng",
        iso639_3: "eng",
        dir: "ltr"
    },
    Region: {},
    Currency: {
        englishName: "Persia Derik",
        nativeName: "Persia Derik",
        iso4217: "PRD",
        iso4217_num: 0,
        symbol: "D",
    },
    Charset: "utf-8"
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
Application.MostUsedPages = ["my", "orgs", "finance", "store"]
// }
// Call init function to application work on not supported browsers!! that now there is no browser!! ;)
// init()

// function main() {
Application.Start()
// }
// Call main function to application work on not supported browsers!! that now there is no browser!! ;)
// main()
