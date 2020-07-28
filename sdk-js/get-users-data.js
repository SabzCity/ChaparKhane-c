/* For license and copyright information please see LEGAL file in repository */

/**
 * GetUserData Request
 * @Typedef {Object} GetUserDataReq
 * @property {string} ID User UUID
 */

/**
 * GetUserData Response
 * @Typedef {Object} GetUserDataRes
 * @property {string} ID User UUID
 * @property {string} Name User UUID
 * @property {string} Picture Picture URL || Object UUID
 * @property {Number[]} ServiceMenu Service menu order
 */

/**
 * GetUserData use to retrieve user basic info like name & picture.
 * @param {GetUserDataReq} req
 * @returns {GetUserDataRes} res
 */
function GetUserData(req) {
    // just for test SDK
    return {
        ID: "",
        Name: "",
        Picture: "",
    }
}