/* For license and copyright information please see LEGAL file in repository */

Application.Pages["recover"] = {
    ID: "recover",
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
    Icon: "history",
    Related: [],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["recover"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

Application.Pages["recover"].next = function () {
    const card = window.document.getElementById('container')
    const animateOut = [{
        transform: ' translate(0) '
    },
    {
        transform: ' translateX(-100vw)'
    },
    ]
    const animateIn = [{
        transform: ' translate(100vw) '
    },
    {
        transform: ' translateX(0)'
    },
    ]
    const duration = 400
    const animation = card.animate(animateOut, duration)

    animation.onfinish = () => {
        card.animate(animateIn, duration)
        this.step = 'validate'
    }
}

Application.Pages["recover"].back = function () {
    const card = window.document.getElementById('container')
    const animateOut = [{
        transform: ' translate(0) '
    },
    {
        transform: ' translateX(100vw)'
    },
    ]
    const animateIn = [{
        transform: ' translate(-100vw) '
    },
    {
        transform: ' translateX(0)'
    },
    ]
    const duration = 400
    const animation = card.animate(animateOut, duration)

    animation.onfinish = () => {
        card.animate(animateIn, duration)
        this.step = 'select'
    }
}
