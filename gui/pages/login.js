/* For license and copyright information please see LEGAL file in repository */

import '../libjs/application.js'
import '../libjs/base64.js'
import '../../sdk-js/authenticate-app-connection.js'
import '../../sdk-js/get-new-phrase-captcha.js'
import '../../sdk-js/solve-phrase-captcha.js'
import '../../sdk-js/get-person-number-status.js'

Application.Pages["login"] = {
    ID: "login",
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
    Icon: "fingerprint",
    Related: ["register", "recover"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "logged-in-user": (UserPreferences) => ``,
    },
}

// https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest

Application.Pages["login"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    const LoggedInUserElement = window.document.getElementById('LoggedInUser')
    for (id in users.poolByID) {
        LoggedInUserElement.insertAdjacentHTML('beforeend', Application.Pages["login"].Templates["logged-in-user"](users.poolByID[id]))
    }
}

Application.Pages["login"].DisconnectedCallback = function () {
}

function loginPageGoToNewLoginPhase() {
    const userIdentifierPhaseElement = window.document.getElementById('userIdentifierPhase')
    userIdentifierPhaseElement.hidden = true

    const newLoginPhaseElement = window.document.getElementById('newLoginPhase')
    newLoginPhaseElement.hidden = false

    loginPageRefreshCaptcha()
}

function loginPageGoToUserIdentifierPhase() {
    const userIdentifierPhaseElement = window.document.getElementById('userIdentifierPhase')
    userIdentifierPhaseElement.hidden = false

    const newLoginPhaseElement = window.document.getElementById('newLoginPhase')
    newLoginPhaseElement.hidden = true
}

const AuthenticateAppConnectionReq = {
    CaptchaID: "", // [16]byte
    PersonID: "", // [16]byte
    PasswordHash: "", // [32]byte
    OTP: 0,  // uint32
}

// Get captchaID and set related img tag
async function loginPageRefreshCaptcha() {
    const GetNewPhraseCaptchaReq = {
        Language: 0, // EN
        ImageFormat: 0, // PNG
    }
    try {
        const res = await GetNewPhraseCaptcha(GetNewPhraseCaptchaReq)
        window.document.getElementById('captchaImg').src = "data:image/png;base64," + res.Image
        AuthenticateAppConnectionReq.CaptchaID = res.CaptchaID
    } catch (err) {
        switch (err) {
            case 0:
            default:
        }
    }
}

async function loginPageLogin() {
    const userIdentifierElement = window.document.getElementById('userIdentifier')
    if (users.active.UserID !== GuestUserID) {
        userIdentifierElement.setCustomValidity("LocaleText[40]")
        userIdentifierElement.reportValidity()
        return
    }

    // TODO::: Warn about privacy changes and store needed cookie or local storage

    // Solve captcha
    const captchaTextElement = window.document.getElementById('captchaText')
    const captchaTextElementValue = captchaTextElement.value
    if (captchaTextElementValue === "") {
        // warn user about required captcha text
        captchaTextElement.setCustomValidity("LocaleText[35]")
        captchaTextElement.reportValidity()
        return
    }

    try {
        const SolvePhraseCaptchaReq = {
            "CaptchaID": AuthenticateAppConnectionReq.CaptchaID,
            "Answer": captchaTextElementValue
        }
        await SolvePhraseCaptcha(SolvePhraseCaptchaReq)
        // Captcha solved, Nothing to do!
    } catch (err) {
        switch (err) {
            case 666099564: // Expired
            case 563598327: // NotFound
                // Last captcha Expired-NotFound, Solved new one
                loginPageRefreshCaptcha()
                captchaTextElement.setCustomValidity("LocaleText[37]")
                captchaTextElement.reportValidity()
                return
            case 3952537117:
                // Last given captcha answer not valid! Warn user to re-enter captcha
                captchaTextElement.setCustomValidity("LocaleText[38]")
                captchaTextElement.reportValidity()
                return
            default:
                // TODO::: warn user about no network!
                console.log(err)
                return
        }
    }

    // TODO::: Need to validate phone number here?
    let userIdentifierElementValue = userIdentifierElement.value
    if (userIdentifierElementValue === "") {
        userIdentifierElement.setCustomValidity("LocaleText[35]")
        userIdentifierElement.reportValidity()
        return
    }

    let GetPersonNumberStatusRes
    if (userIdentifierElementValue.startsWith("0")) {
        userIdentifierElementValue = userIdentifierElementValue.substring(1)
        const phoneNumber = "98" + userIdentifierElementValue

        // check phone number registered before this request!
        const GetPersonNumberStatusReq = {
            CaptchaID: AuthenticateAppConnectionReq.CaptchaID,
            PhoneNumber: Number(phoneNumber)
        }
        try {
            let GetPersonNumberStatusRes = await GetPersonNumberStatus(GetPersonNumberStatusReq)
            if (GetPersonNumberStatusRes.Status === 0 || GetPersonNumberStatusRes.Status === 2 || GetPersonNumberStatusRes.PersonID === GuestUserID) {
                userIdentifierElement.setCustomValidity("LocaleText[36]")
                userIdentifierElement.reportValidity()
                return
            }
            AuthenticateAppConnectionReq.PersonID = GetPersonNumberStatusRes.PersonID
        } catch (err) {
            // TODO::: toast related dialog to warn user about server 500 situation!
            // console.log(err)
        }
    } else if (userIdentifierElementValue.includes("@")) {
        // email
    } else {
        // username
    }

    // TODO::: validate password strength
    // https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest
    const passwordElement = window.document.getElementById('password')
    const passwordElementValue = passwordElement.value
    if (passwordElementValue === "") {
        passwordElement.setCustomValidity("LocaleText[35]")
        passwordElement.reportValidity()
        return
    }
    AuthenticateAppConnectionReq.PasswordHash = await base64.stdWithoutPadding.PasswordHash(passwordElementValue)

    const otpElement = window.document.getElementById('OTP')
    AuthenticateAppConnectionReq.OTP = Number(otpElement.value)

    let AuthenticateAppConnectionRes
    try {
        AuthenticateAppConnectionRes = await AuthenticateAppConnection(AuthenticateAppConnectionReq)
        // TODO::: If user register before login her||him
    } catch (err) {
        switch (err) {
            case 2605698703: // Authorization - User Not Allow.
                passwordElement.setCustomValidity("LocaleText[40]")
                passwordElement.reportValidity()
                otpElement.setCustomValidity("LocaleText[40]")
                otpElement.reportValidity()
                return
            case 3552753310: // Bad password or OTP
                passwordElement.setCustomValidity("LocaleText[39]")
                passwordElement.reportValidity()
                otpElement.setCustomValidity("LocaleText[39]")
                otpElement.reportValidity()
                return
            default:
                // TODO::: warn user about no network!
                console.log(err)
                return
        }
    }

    users.ChangeActiveUser(AuthenticateAppConnectionReq.PersonID)
    users.active.UserNumber = GetPersonNumberStatusRes.PhoneNumber

    // if (Application.PreviousPage) {
    //     window.history.back()
    //     window.location.reload()
    // } else {
    window.location.replace("/" + users.active.HomePage)
    // }

}

async function loginPageLogout() {
    alert("Sorry! Not implemented yet!")
}