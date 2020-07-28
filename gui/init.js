/* For license and copyright information please see LEGAL file in repository */

// function init() {
Application.Initialize(
    {
        Icon: "app-icon-512x512.png",
        Info: {
            Name: "LocaleText[0]",
            ShortName: "LocaleText[1]",
            Tagline: "LocaleText[2]",
            Slogan: "LocaleText[3]",
            Description: "LocaleText[4]",
            Tags: ["LocaleText[5]"]
        },
        ContentPreferences: {
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
            Currency: {
                englishName: "Persia Derik",
                nativeName: "Persia Derik",
                iso4217: "PRD",
                iso4217_num: 0,
                symbol: "D",
            },
            Charset: "utf-8",
        },
        PresentationPreferences: {
            DesignLanguage: "material",
            ColorScheme: "no-preference",
            ThemeColor: "#66ff55",
            PrimaryFontFamily: "Roboto",
            Display: "standalone",
            Orientation: "portrait",
        },
        HomePage: "store", // start with store page!
        MostUsedPages: [
            "my", "orgs", "finance", "store",
        ],
    }
)
// }
// Call init function to application work on not supported browsers!! that now there is no browser!! ;)
// init()
