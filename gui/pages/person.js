/* For license and copyright information please see LEGAL file in repository */

import '../libjs/time.js'
import '../../sdk-js/get-person-number.js'

const personPage = {
    ID: "person",
    Conditions: {
        id: "",
    },
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
    Related: [], // "groups", "sessions", "settings", "security"
    HTML: () => ``,
    CSS: '',
    Templates: {},
}
pages.poolByID[personPage.ID] = personPage

personPage.ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    personPage.GetPhoneNumber()
}

personPage.DisconnectedCallback = function () {
}

personPage.GetPhoneNumber = async function() {
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
        errors.HandleError(err)
    }
}
