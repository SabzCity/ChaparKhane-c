/* For license and copyright information please see LEGAL file in repository */


Application.Pages["login"] = {
    ID: "login",
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
    Icon: "fingerprint",
    Related: ["register", "recover"],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

// https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest

Application.Pages["login"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["login"].DisconnectedCallback = function () {
}

// userID: String,
// userName: String,
// twoStep: Boolean,

Application.Pages["login"].next = function () {
    // Validate input data

    // Warn about privacy changes and store needed cookie or local storage

    const usernamePhase = window.document.getElementById('usernamePhase')
    const passwordPhase = window.document.getElementById('passwordPhase')

    const animation = usernamePhase.animate(animateSwapLeftOut, animateDuration)
    animation.onfinish = () => {
        usernamePhase.hidden = true
        passwordPhase.hidden = false
        passwordPhase.animate(animateSwapRightIn, animateDuration)
    }
}

Application.Pages["login"].back = function () {
    const usernamePhase = window.document.getElementById('usernamePhase')
    const passwordPhase = window.document.getElementById('passwordPhase')

    const animation = passwordPhase.animate(animateSwapRightOut, animateDuration)
    animation.onfinish = () => {
        passwordPhase.hidden = true
        usernamePhase.hidden = false
        usernamePhase.animate(animateSwapLeftIn, animateDuration)
    }
}

Application.Pages["login"].login = function () {
    // Send login data to apis.sabz.city

    // add usersID to UsersState
    Application.UserPreferences.UsersState.ActiveUserID = ""
    Application.UserPreferences.UsersState.UsersID.push(Application.UserPreferences.UsersState.ActiveUserID)

    // send user to last page if exist or my home page as default login page
    if (Application.PreviousPage && Application.PreviousPage.ActiveURI) {
        window.location.replace(Application.PreviousPage.ActiveURI)
    } else {
        window.location.replace("/my")
    }

}

// Animation const helper
const animateDuration = 400
const animateSwapRightIn = [
    {
        transform: ' translate(100vw) '
    }, {
        transform: ' translateX(0)'
    },
]
const animateSwapRightOut = [
    {
        transform: ' translate(0) '
    }, {
        transform: ' translateX(100vw)'
    },
]
const animateSwapLeftIn = [
    {
        transform: ' translate(-100vw) '
    }, {
        transform: ' translateX(0)'
    },
]
const animateSwapLeftOut = [
    {
        transform: ' translate(0) '
    }, {
        transform: ' translateX(-100vw)'
    },
]
