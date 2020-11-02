/* For license and copyright information please see LEGAL file in repository */

import '../libjs/time.js'
import '../../sdk-js/get-person-number.js'

Application.Pages["my"] = {
    ID: "my",
    Conditions: {},
    State: "",
    Robots: "all",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "person",
    Related: ["aboutme"], // "groups", "sessions", "settings", "security"
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["my"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    myPageGetPhoneNumber()
}

Application.Pages["my"].DisconnectedCallback = function () {
}

async function myPageGetPhoneNumber() {
    const userNumberElement = document.getElementById("userNumber")
    const userNumberWriteTimeElement = document.getElementById("userNumberWriteTime")
    const userNumberStatusElement = document.getElementById("userNumberStatus")

    try {
        let res = await GetPersonNumber()
        userNumberElement.innerText = res.Number
        userNumberWriteTimeElement.innerText = time.unix.String(res.WriteTime)
        userNumberStatusElement.innerText = res.Status

        users.active.UserNumber = res.Number
    } catch (err) {
        // TODO::: toast related dialog to warn user about server 500 situation!
        // console.log(err)
    }
}
