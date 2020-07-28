/* For license and copyright information please see LEGAL file in repository */

import '../../sdk-js/register-new-person.js'
import '../../sdk-js/get-new-phrase-captcha.js'
import '../../sdk-js/solve-phrase-captcha.js'
import '../../sdk-js/send-otp.js'

Application.Pages["register"] = {
    ID: "register",
    RecordID: null,
    Condition: {},
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

// Get captchaID and set related img tag
async function registerPageRefreshCaptcha() {
    const GetNewPhraseCaptchaReq = {
        Language: 0, // EN
        ImageFormat: 0, // PNG
    }
    try {
        const res = await GetNewPhraseCaptcha(GetNewPhraseCaptchaReq)
        window.document.getElementById('captchaImg').src = "data:image/png;base64," + atob(res.Image)
        RegisterNewPersonReq.CaptchaID = res.CaptchaID
    } catch (err) {

    }
}

let registerPageResendOTPTimerDelay = 120000

// Send OTP to person phone number
async function registerPageSendOTP() {
    // disabled related button
    document.getElementById("resendOTPButton").disabled = true

    // SendOtpReq is the request structure of SendOtp()
    const SendOtpReq = {
        "PhoneNumber": RegisterNewPersonReq.PhoneNumber, // uint64
        "PhoneType": 0, // uint8
        "CaptchaID": RegisterNewPersonReq.CaptchaID, // [16]byte
    }
    SendOtp(SendOtpReq)

    // Set the time we're counting down to
    let end = Date.now() + registerPageResendOTPTimerDelay

    // Update the count down every 10 second
    let timer = setInterval(function () {
        // Find the distance between now and the end
        let distance = end - Date.now()
        // If the count down is finished, let user use resendOTPButton
        if (distance <= 0) {
            clearInterval(timer)
            document.getElementById("resendOTPButton").disabled = false
            document.getElementById("resendOTPTimer").innerText = ""
            return
        }

        // Time calculations for minutes and seconds
        let minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60))
        let seconds = Math.floor((distance % (1000 * 60)) / 1000)
        // Display the result in the element
        document.getElementById("resendOTPTimer").innerText = minutes + "m : " + seconds + "s "

    }, 10000)

    registerPageResendOTPTimerDelay *= 2
}

// RegisterNewPersonReq is the request structure of RegisterNewPerson()
const RegisterNewPersonReq = {
    "PhoneNumber": 0, // uint64
    "PhoneOTP": 0, // uint64
    "PasswordHash": [], // [32]byte
    "CaptchaID": [], // [16]byte
}

async function registerPageNextStep() {
    // TODO::: Need to validate phone number here?
    const phoneNumber = window.document.getElementById('phoneNumber').value
    RegisterNewPersonReq.PhoneNumber = Number(phoneNumber)

    // Solve captcha
    const captchaTextElement = window.document.getElementById('captchaText')
    if (captchaTextElement.value === "") {
        // warn user about required captcha text
        captchaTextElement.setCustomValidity("LocaleText[43]")
        captchaTextElement.reportValidity()
        return
    }
    const SolvePhraseCaptchaReq = {
        "CaptchaID": RegisterNewPersonReq.CaptchaID,
        "Answer": captchaTextElement.value
    }
    try {
        const SolvePhraseCaptchaRes = await SolvePhraseCaptcha(SolvePhraseCaptchaReq)
        switch (SolvePhraseCaptchaRes.CaptchaState) {
            case 0:
            case 1, 2:
                registerPageRefreshCaptcha()
                captchaTextElement.setCustomValidity("LocaleText[39]")
                captchaTextElement.reportValidity()
                throw "Last captcha expired or not exist, Solved new one"
            case 3:
                captchaTextElement.setCustomValidity("LocaleText[40]")
                captchaTextElement.reportValidity()
                throw "Last given captcha answer not valid! Warn user to re-enter captcha"
            case 4:
            // Captcha solved, Nothing to do!
        }
    } catch (err) {
        // TODO::: warn user about no network!
        // console.log(err)
        return
    }

    const phonePhase = window.document.getElementById('phonePhase')
    const otpPhase = window.document.getElementById('otpPhase')

    // TODO::: change to CSS animate
    const animation = phonePhase.animate(animateSwapLeftOut, animateDuration)
    animation.onfinish = () => {
        phonePhase.hidden = true
        otpPhase.hidden = false
        otpPhase.animate(animateSwapRightIn, animateDuration)
    }

    // request OTP for given phone number
    registerPageSendOTP()

    // TODO::: validate password strength
    // https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest
    const encoder = new TextEncoder()
    const pass = encoder.encode(window.document.getElementById('password').value)
    const hash = await crypto.subtle.digest('SHA-256', pass)
    RegisterNewPersonReq.PasswordHash = Array.from(new Uint8Array(hash)) // [... new Uint8Array(hash)]
}

async function registerPageRegister() {
    const receivedOTP = window.document.getElementById('receivedOTP')
    const userEnterOPT = Number(receivedOTP.value)
    if (userEnterOPT === 0) {
        // warn user about required OTP
        receivedOTP.setCustomValidity("LocaleText[43]")
        receivedOTP.reportValidity()
        return
    }
    RegisterNewPersonReq.PhoneOTP = userEnterOPT

    // send user data to register service by SDK
    try {
        var res = await RegisterNewPerson(RegisterNewPersonReq)
        // TODO::: If user register before login her||him
    } catch (err) {
        // TODO::: Back user to previous step to register again
    }

    // TODO::: login user by data
    // console.log(res)

    // send user to last page if exist or my home page as default login page
    if (Application.PreviousPage && Application.PreviousPage.ActiveURI) {
        window.location.replace(Application.PreviousPage.ActiveURI)
    } else {
        window.location.replace("/my")
    }
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
