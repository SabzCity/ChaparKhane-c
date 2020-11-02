/* For license and copyright information please see LEGAL file in repository */

import '../libjs/application.js'
import '../libjs/base64.js'
import '../../sdk-js/register-new-person.js'
import '../../sdk-js/get-new-phrase-captcha.js'
import '../../sdk-js/solve-phrase-captcha.js'
import '../../sdk-js/send-otp.js'
import '../../sdk-js/get-person-number-status.js'

Application.Pages["register"] = {
    ID: "register",
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
    Icon: "person_add",
    Related: ["login", "terms"],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["register"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()

    registerPageRefreshCaptcha()
}

Application.Pages["register"].DisconnectedCallback = function () {
}

// RegisterNewPersonReq is the request structure of RegisterNewPerson()
const RegisterNewPersonReq = {
    "PhoneNumber": 0, // uint64
    "PhoneOTP": 0, // uint64
    "PasswordHash": "", // [32]byte
    "CaptchaID": "", // [16]byte
}

// Get captchaID and set related img tag
async function registerPageRefreshCaptcha() {
    const GetNewPhraseCaptchaReq = {
        Language: 0, // EN
        ImageFormat: 0, // PNG
    }
    try {
        const res = await GetNewPhraseCaptcha(GetNewPhraseCaptchaReq)
        window.document.getElementById('captchaImg').src = "data:image/png;base64," + res.Image
        RegisterNewPersonReq.CaptchaID = res.CaptchaID
    } catch (err) {
        switch (err) {
            case 0:
            default:
        }
    }
}

let registerPageResendOTPTimerDelay = 1800000 // 30 minutes

// Send OTP to person phone number
async function registerPageSendOTP() {
    // TODO::: Need to validate phone number here?
    const phoneNumberElement = window.document.getElementById('phoneNumber')
    let phoneNumberElementValue = phoneNumberElement.value
    if (phoneNumberElementValue === "") {
        phoneNumberElement.setCustomValidity("LocaleText[43]")
        phoneNumberElement.reportValidity()
        return
    }
    if (phoneNumberElementValue.startsWith("0")) phoneNumberElementValue = phoneNumberElementValue.substring(1)
    const phoneNumber = "98" + phoneNumberElementValue
    RegisterNewPersonReq.PhoneNumber = Number(phoneNumber)

    // Solve captcha
    const captchaTextElement = window.document.getElementById('captchaText')
    const captchaTextElementValue = captchaTextElement.value
    if (captchaTextElementValue === "") {
        // warn user about required captcha text
        captchaTextElement.setCustomValidity("LocaleText[43]")
        captchaTextElement.reportValidity()
        return
    }

    try {
        const SolvePhraseCaptchaReq = {
            "CaptchaID": RegisterNewPersonReq.CaptchaID,
            "Answer": captchaTextElementValue
        }
        await SolvePhraseCaptcha(SolvePhraseCaptchaReq)
        // Captcha solved, Nothing to do!
    } catch (err) {
        switch (err) {
            case 666099564: // Expired
            case 563598327: // NotFound
                // Last captcha Expired-NotFound, Solved new one
                registerPageRefreshCaptcha()
                captchaTextElement.setCustomValidity("LocaleText[39]")
                captchaTextElement.reportValidity()
                return
            case 3952537117:
                // Last given captcha answer not valid! Warn user to re-enter captcha
                captchaTextElement.setCustomValidity("LocaleText[40]")
                captchaTextElement.reportValidity()
                return
            default:
                // TODO::: warn user about no network!
                console.log(err)
                return
        }
    }

    // check phone number registered before this request!
    try {
        const GetPersonNumberStatusReq = {
            CaptchaID: RegisterNewPersonReq.CaptchaID,
            PhoneNumber: RegisterNewPersonReq.PhoneNumber
        }
        let res = await GetPersonNumberStatus(GetPersonNumberStatusReq)
        if (res.Status == 1 || res.Status == 3) {
            phoneNumberElement.setCustomValidity("LocaleText[44]")
            phoneNumberElement.reportValidity()
            return
        }
        if (users.active.UserID === GuestUserID) {
            users.active.UserNumber = RegisterNewPersonReq.PhoneNumber
        }
    } catch (err) {
        switch (err) {
            case 1685872164:
            // Ganjine - Record Not Found
        }
        // TODO::: toast related dialog to warn user about server 500 situation!
        // console.log(err)
    }

    // disabled related button
    const sendOTPButtonElement = document.getElementById("sendOTPButton")
    sendOTPButtonElement.disabled = true

    try {
        // SendOtpReq is the request structure of SendOtp()
        const SendOtpReq = {
            "CaptchaID": RegisterNewPersonReq.CaptchaID, // [16]byte
            "PhoneNumber": RegisterNewPersonReq.PhoneNumber, // uint64
            "PhoneType": 0, // uint8
            "Language": 1, // TODO::: get user language to send
        }
        await SendOtp(SendOtpReq)
    } catch (err) {
        switch (err) {
            case 3050345768:
                console.log("Asanak.com - Bad Request")
                return
            case 847328815:
                console.log("Asanak.com - Bad Response")
                return
            case 4161691025:
                // SMS provider can't reach
                return
        }
    }

    const otpTimerElement = document.getElementById("otpTimer")
    otpTimerElement.hidden = false
    const otpInputElement = document.getElementById("otpInput")
    otpInputElement.hidden = false

    // Set the time we're counting down to
    let end = Date.now() + registerPageResendOTPTimerDelay

    // Update the count down every 60 second
    let timer = setInterval(function () {
        // Find the distance between now and the end
        let distance = end - Date.now()
        // If the count down is finished, let user use sendOTPButtonElement again
        if (distance <= 0) {
            clearInterval(timer)
            sendOTPButtonElement.disabled = false
            document.getElementById("resendOTPTimer").innerText = ""
            return
        }

        // Time calculations for minutes and seconds
        let minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60))
        // let seconds = Math.floor((distance % (1000 * 60)) / 1000)
        // Display the result in the element
        document.getElementById("resendOTPTimer").innerText = minutes + " LocaleText[46]" // + seconds + "s "

    }, 60000)
}

async function registerPageRegister() {
    if (RegisterNewPersonReq.PhoneNumber === 0) {
        registerPageSendOTP()
        return
    }

    // TODO::: validate password strength
    // https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest
    const passwordElement = window.document.getElementById('password')
    const passwordElementValue = passwordElement.value
    if (passwordElementValue === "") {
        passwordElement.setCustomValidity("LocaleText[43]")
        passwordElement.reportValidity()
        return
    }
    RegisterNewPersonReq.PasswordHash = await base64.stdWithoutPadding.PasswordHash(passwordElementValue)

    const receivedOTP = window.document.getElementById('receivedOTP')
    const receivedOTPValue = Number(receivedOTP.value)
    if (receivedOTPValue === 0) {
        // warn user about required OTP
        receivedOTP.setCustomValidity("LocaleText[43]")
        receivedOTP.reportValidity()
        return
    }
    RegisterNewPersonReq.PhoneOTP = receivedOTPValue

    // send user data to register service by SDK
    let RegisterNewPersonRes
    try {
        RegisterNewPersonRes = await RegisterNewPerson(RegisterNewPersonReq)
        // TODO::: If user register before login her||him
    } catch (err) {
        switch (err) {
            case 4094100234: // Bad OTP number
                receivedOTP.setCustomValidity("LocaleText[45]")
                receivedOTP.reportValidity()
                return
            case 1019445371:
                console.log("Phone number registered already, But no reachable code")
                return
            default:
                // TODO::: warn user about no network!
                // console.log(err)
                return
        }
    }

    if (users.active.UserID === GuestUserID) {
        users.ChangeActiveUser(RegisterNewPersonRes.PersonID)
    } else {
        // TODO::: toast a dialog and ask user about add new user to exiting users
    }

    // send user to last page if exist or my home page as default login page
    // if (Application.PreviousPage) {
    //     window.history.back()
    //     window.location.reload()
    // } else {
    window.location.replace("/" + users.active.HomePage)
    // }
}

function registerPageToggleTermDialog() {
    if (document.getElementById("termDialog").open === true) {
        document.getElementById("termDialog").open = false
        document.getElementById("termDialogBack").setAttribute('hidden', '')
        // window.document.getElementById("termDialog").close()
    } else {
        document.getElementById("termDialogBack").removeAttribute('hidden')
        document.getElementById("termDialog").open = true
        // document.getElementById("termDialog").showModal()
    }
}

function registerPagePlayCaptchaAudio() {
    // TODO:::
}
