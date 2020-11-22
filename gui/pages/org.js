/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/errors.js'
import '../libjs/cookie.js'
import '../libjs/price/currency.js'
import '../../sdk-js/datastore-organization-authentication.js'
import '../../sdk-js/register-new-organization.js'
import '../../sdk-js/update-organization.js'
import '../../sdk-js/get-organization-by-id.js'

const orgPage = {
    ID: "org",
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
    Icon: "business",
    Related: [],
    HTML: (org) => ``,
    CSS: '',
    Templates: {},
}
pages.poolByID["org"] = orgPage

orgPage.ConnectedCallback = async function () {
    let org
    if (this.Conditions.id) {
        try {
            // GetOrganizationByIDReq is the request structure of GetOrganizationByID()
            const GetOrganizationByIDReq = {
                "ID": this.Conditions.id,
            }
            org = await GetOrganizationByID(GetOrganizationByIDReq)
        } catch (err) {
            org = this.NotFoundPage()
            errors.HandleError(err)
        }
    } else {
        org = this.NotFoundPage()
        errors.HandleError(errors.poolByID[1685872164])
    }

    window.document.body.innerHTML = this.HTML(org)
}

orgPage.DisconnectedCallback = function () { }

orgPage.NotFoundPage = function () {
    return org = {
        Name: "LocaleText[25]",
        Domain: "www.google.com/search?q=LocaleText[0]",
        FinancialCreditAmount: 0,
        ServicesType: 0,
        Status: 0
    }
}

orgPage.EnableNew = function () {
    if (cookie.GetByName(HTTPCookieNameBaseUserID) !== AdminUserID) {
        this.ToggleTermDialog()
        return
    }

    const newOrgElement = document.getElementById("newOrg")
    const editOrgElement = document.getElementById("editOrg")
    const saveChangesButtonElement = document.getElementById("saveChanges")
    const discardChangesButtonElement = document.getElementById("discardChanges")

    const orgNameValueElement = document.getElementById('orgNameValue')
    const orgNameInputElement = document.getElementById('orgNameInput')

    const orgDomainValueElement = document.getElementById('orgDomainValue')
    const orgDomainInputElement = document.getElementById('orgDomainInput')

    const orgFinancialCreditAmountValueElement = document.getElementById('orgFinancialCreditAmountValue')
    const orgFinancialCreditAmountInputElement = document.getElementById('orgFinancialCreditAmountInput')

    const orgShortServicesTypeValueElement = document.getElementById('orgShortServicesTypeValue')
    const orgLongServicesTypeValueElement = document.getElementById('orgLongServicesTypeValue')
    const orgServicesTypeSelectElement = document.getElementById('orgServicesTypeSelect')

    const orgLeaderPersonIDElement = document.getElementById('orgLeaderPersonID')
    
    discardChangesButtonElement.disabled = false
    saveChangesButtonElement.disabled = false
    editOrgElement.disabled = true
    newOrgElement.disabled = true
    
    orgNameValueElement.hidden = true
    orgNameInputElement.hidden = false

    orgDomainInputElement.hidden = false
    orgDomainValueElement.hidden = true

    orgFinancialCreditAmountInputElement.hidden = false
    orgFinancialCreditAmountValueElement.hidden = true

    orgShortServicesTypeValueElement.hidden = true
    orgLongServicesTypeValueElement.hidden = true
    orgServicesTypeSelectElement.hidden = false

    orgLeaderPersonIDElement.hidden = false
}

orgPage.EnableEdit = function () {
    if (cookie.GetByName(HTTPCookieNameBaseUserID) !== AdminUserID) {
        this.ToggleTermDialog()
        return
    }

    const editOrgElement = document.getElementById("editOrg")
    const saveChangesButtonElement = document.getElementById("saveChanges")
    const discardChangesButtonElement = document.getElementById("discardChanges")

    const orgNameValueElement = document.getElementById('orgNameValue')
    const orgNameInputElement = document.getElementById('orgNameInput')

    const orgDomainValueElement = document.getElementById('orgDomainValue')
    const orgDomainInputElement = document.getElementById('orgDomainInput')

    const orgFinancialCreditAmountValueElement = document.getElementById('orgFinancialCreditAmountValue')
    const orgFinancialCreditAmountInputElement = document.getElementById('orgFinancialCreditAmountInput')

    const orgThingIDValueElement = document.getElementById('orgThingIDValue')
    const orgThingIDInputElement = document.getElementById('orgThingIDInput')

    const orgCoordinateIDValueElement = document.getElementById('orgCoordinateIDValue')
    const orgCoordinateIDInputElement = document.getElementById('orgCoordinateIDInput')

    const orgShortServicesTypeValueElement = document.getElementById('orgShortServicesTypeValue')
    const orgLongServicesTypeValueElement = document.getElementById('orgLongServicesTypeValue')
    const orgServicesTypeSelectElement = document.getElementById('orgServicesTypeSelect')

    const orgShortStatusValueElement = document.getElementById('orgShortStatusValue')
    const orgStatusSelectElement = document.getElementById('orgStatusSelect')
    const orgLongStatusValueElement = document.getElementById('orgLongStatusValue')

    editOrgElement.disabled = true
    saveChangesButtonElement.disabled = false
    discardChangesButtonElement.disabled = false

    orgNameValueElement.hidden = true
    orgNameInputElement.hidden = false

    orgDomainValueElement.hidden = true
    orgDomainInputElement.hidden = false

    orgFinancialCreditAmountValueElement.hidden = true
    orgFinancialCreditAmountInputElement.hidden = false

    orgThingIDValueElement.hidden = true
    orgThingIDInputElement.hidden = false

    orgCoordinateIDValueElement.hidden = true
    orgCoordinateIDInputElement.hidden = false

    orgShortServicesTypeValueElement.hidden = true
    orgLongServicesTypeValueElement.hidden = true
    orgServicesTypeSelectElement.hidden = false

    orgShortStatusValueElement.hidden = true
    orgLongStatusValueElement.hidden = true
    orgStatusSelectElement.hidden = false
}

orgPage.SaveEdit = async function () {
    const orgNameInputElement = document.getElementById('orgNameInput')
    const orgDomainInputElement = document.getElementById('orgDomainInput')
    const orgFinancialCreditAmountInputElement = document.getElementById('orgFinancialCreditAmountInput')
    const orgServicesTypeSelectElement = document.getElementById('orgServicesTypeSelect')
    const ordIDValueElement = document.getElementById('ordIDValue')
    const newOrgElement = document.getElementById("newOrg")

    if (newOrgElement.disabled === true) {
        const orgLeaderPersonIDInputElement = document.getElementById('orgLeaderPersonIDInput')
        const orgLeaderPersonIDElement = document.getElementById('orgLeaderPersonID')

        try {
            // RegisterNewOrganizationReq is the request structure of RegisterNewOrganization()
            const RegisterNewOrganizationReq = {
                "LeaderPersonID": orgLeaderPersonIDInputElement.value,
                "Name": orgNameInputElement.value,
                "Domain": orgDomainInputElement.value,
                "FinancialCreditAmount": Number(orgFinancialCreditAmountInputElement.value),
                "ServicesType": Number(orgServicesTypeSelectElement.value),
            }
            let res = await RegisterNewOrganization(RegisterNewOrganizationReq)
            ordIDValueElement.innerText = res.ID
        } catch (err) {
            return errors.HandleError(err)
        }

        orgLeaderPersonIDElement.hidden = true
        newOrgElement.disabled = false
    } else {
        const orgThingIDInputElement = document.getElementById('orgThingIDInput')
        const orgCoordinateIDInputElement = document.getElementById('orgCoordinateIDInput')
        const orgStatusSelectElement = document.getElementById('orgStatusSelect')

        try {
            // UpdateOrganizationReq is the request structure of UpdateOrganization()
            const UpdateOrganizationReq = {
                "ID": ordIDValueElement.innerText,
                "Name": orgNameInputElement.value,
                "Domain": orgDomainInputElement.value,
                "FinancialCreditAmount": Number(orgFinancialCreditAmountInputElement.value),
                "ThingID": orgThingIDInputElement.value,
                "CoordinateID": orgCoordinateIDInputElement.value,
                "ServicesType": Number(orgServicesTypeSelectElement.value),
                "Status": Number(orgStatusSelectElement.value),
            }
            await UpdateOrganization(UpdateOrganizationReq)
        } catch (err) {
            return errors.HandleError(err)
        }

        const orgThingIDValueElement = document.getElementById('orgThingIDValue')
        const orgCoordinateIDValueElement = document.getElementById('orgCoordinateIDValue')
        const orgShortStatusValueElement = document.getElementById('orgShortStatusValue')
        const orgLongStatusValueElement = document.getElementById('orgLongStatusValue')
        const editOrgElement = document.getElementById("editOrg")

        orgThingIDInputElement.hidden = true
        orgThingIDValueElement.hidden = false
        orgThingIDValueElement.innerText = orgThingIDInputElement.value

        orgCoordinateIDInputElement.hidden = true
        orgCoordinateIDValueElement.hidden = false
        orgCoordinateIDValueElement.innerText = orgCoordinateIDInputElement.value

        orgStatusSelectElement.hidden = true
        orgShortStatusValueElement.hidden = false
        orgShortStatusValueElement.innerText = OrganizationAuthenticationStatus.GetShortDetailByID(Number(orgStatusSelectElement.value))
        orgLongStatusValueElement.hidden = false
        orgLongStatusValueElement.innerText = OrganizationAuthenticationStatus.GetLongDetailByID(Number(orgStatusSelectElement.value))

        editOrgElement.disabled = false
    }

    const orgNameValueElement = document.getElementById('orgNameValue')
    const orgDomainValueElement = document.getElementById('orgDomainValue')
    const orgFinancialCreditAmountValueElement = document.getElementById('orgFinancialCreditAmountValue')
    const orgShortServicesTypeValueElement = document.getElementById('orgShortServicesTypeValue')
    const orgLongServicesTypeValueElement = document.getElementById('orgLongServicesTypeValue')
    const saveChangesButtonElement = document.getElementById("saveChanges")
    const discardChangesButtonElement = document.getElementById("discardChanges")

    orgNameInputElement.hidden = true
    orgNameValueElement.hidden = false
    orgNameValueElement.innerText = orgNameInputElement.value

    orgDomainInputElement.hidden = true
    orgDomainValueElement.hidden = false
    orgDomainValueElement.innerText = orgDomainInputElement.value
    orgDomainValueElement.href = "https://" + orgDomainInputElement.value + "/"

    orgFinancialCreditAmountInputElement.hidden = true
    orgFinancialCreditAmountValueElement.hidden = false
    orgFinancialCreditAmountValueElement.innerText = currency.String(orgFinancialCreditAmountInputElement.value)

    orgServicesTypeSelectElement.hidden = true
    orgShortServicesTypeValueElement.hidden = false
    orgShortServicesTypeValueElement.innerText = OrganizationAuthenticationType.GetShortDetailByID(Number(orgServicesTypeSelectElement.value))
    orgLongServicesTypeValueElement.hidden = false
    orgLongServicesTypeValueElement.innerText = OrganizationAuthenticationType.GetShortDetailByID(Number(orgServicesTypeSelectElement.value))

    saveChangesButtonElement.disabled = true
    discardChangesButtonElement.disabled = true
}

orgPage.DiscardChanges = function () {
    alert("Sorry! Not implemented yet!")
}

orgPage.ToggleTermDialog = function () {
    const termDialogElement = document.getElementById("termDialog")
    if (termDialogElement.open === true) {
        termDialogElement.close()
    } else {
        termDialogElement.showModal()
    }
}
