/* For license and copyright information please see LEGAL file in repository */

Application.Pages["aboutme"] = {
    ID: "aboutme",
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
    Icon: "person",
    Related: [],
    HTML: () => ``,
    CSS: '',
    Templates: {},
}

Application.Pages["aboutme"].ConnectedCallback = function () {
    window.document.body.innerHTML = this.HTML()
}

// avatarContainers: {
//     type: Array,
//         value() {
//         return [
//             { firstName: 'cristopher', lastName: 'nolan', nickName: 'the magician' },
//         ];
//     }
// },
// infoSections: {
//     type: Array,
//         value() {
//         return [
//             { name: 'gender', value: 'male' },
//         ];
//     }
// },
// customCards: {
//     type: Array,
//         value() {
//         return [
//             { title: 'address', id: '1', value: '' },
//         ];
//     }
// }