/* For license and copyright information please see LEGAL file in repository */

import '../libjs/pages.js'
import '../libjs/errors.js'
import '../../sdk-js/find-product-auction-by-distribution-center-id.js'
import '../../sdk-js/find-product-auction-by-group-id.js'
import '../../sdk-js/find-product-auction-by-org-id.js'
import '../../sdk-js/find-product-auction-by-wiki-id.js'
import '../../sdk-js/find-product-auction-by-wiki-id-distribution-center-id.js'
import '../../sdk-js/find-product-auction-by-wiki-id-group-id.js'
import '../../sdk-js/get-product-auction.js'

const productAuctionsPage = {
    ID: "product-auctions",
    Conditions: {
        org: "",
        dis: "",
        group: "",
        wiki: "",
        currency: "NaN",
        offset: "NaN",
        limit: "NaN",
    },
    State: "",
    Robots: "noindex, follow, noarchive, noimageindex",
    Info: {
        Name: "LocaleText[0]",
        ShortName: "LocaleText[1]",
        Tagline: "LocaleText[2]",
        Slogan: "LocaleText[3]",
        Description: "LocaleText[4]",
        Tags: ["LocaleText[5]"]
    },
    Icon: "leaderboard",
    Related: ["wiki", "orgs"],
    HTML: () => ``,
    CSS: '',
    Templates: {
        "auction": (productAuction) => ``
    },
}
pages.RegisterPage(productAuctionsPage)

productAuctionsPage.resetCondition = function () {
    this.Conditions = {
        org: "",
        dis: "",
        group: "",
        wiki: "",
        currency: "NaN",
        offset: "NaN",
        limit: "NaN",
    }
}

productAuctionsPage.ConnectedCallback = async function () {
    window.document.body.innerHTML = this.HTML()

    if (isNaN(this.Conditions.offset)) {
        this.Conditions.offset = 1844674407370955161
    } else {
        this.Conditions.offset = Number(this.Conditions.offset)
    }
    if (isNaN(this.Conditions.limit)) {
        this.Conditions.limit = 9
    } else {
        this.Conditions.limit = Number(this.Conditions.limit)
    }

    if (this.Conditions.wiki && this.Conditions.group) {
        try {
            const FindProductAuctionByWikiIDGroupIDReq = {
                "WikiID": this.Conditions.wiki,
                "GroupID": this.Conditions.group,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const res = await FindProductAuctionByWikiIDGroupID(FindProductAuctionByWikiIDGroupIDReq)
            this.showByIDs(res.IDs)
        } catch (err) {
            errors.HandleError(err)
        }
    } else if (this.Conditions.wiki && this.Conditions.dis) {
        try {
            const FindProductAuctionByWikiIDDistributionCenterIDReq = {
                "WikiID": this.Conditions.wiki,
                "DistributionCenterID": this.Conditions.dis,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const res = await FindProductAuctionByWikiIDDistributionCenterID(FindProductAuctionByWikiIDDistributionCenterIDReq)
            this.showByIDs(res.IDs)
        } catch (err) {
            errors.HandleError(err)
        }
    } else if (this.Conditions.org) {
        try {
            const FindProductAuctionByOrgIDReq = {
                "OrgID": this.Conditions.org,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const res = await FindProductAuctionByOrgID(FindProductAuctionByOrgIDReq)
            this.showByIDs(res.IDs)
        } catch (err) {
            errors.HandleError(err)
        }
    } else if (this.Conditions.group) {
        try {
            const FindProductAuctionByGroupIDReq = {
                "GroupID": this.Conditions.group,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const res = await FindProductAuctionByGroupID(FindProductAuctionByGroupIDReq)
            this.showByIDs(res.IDs)
        } catch (err) {
            errors.HandleError(err)
        }
    } else if (this.Conditions.dis) {
        try {
            const FindProductAuctionByDistributionCenterIDReq = {
                "DistributionCenterID": this.Conditions.dis,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const res = await FindProductAuctionByDistributionCenterID(FindProductAuctionByDistributionCenterIDReq)
            this.showByIDs(res.IDs)
        } catch (err) {
            errors.HandleError(err)
        }
    } else if (this.Conditions.wiki) {
        try {
            const FindProductAuctionByWikiIDReq = {
                "WikiID": this.Conditions.wiki,
                "Offset": this.Conditions.offset,
                "Limit": this.Conditions.limit,
            }
            const res = await FindProductAuctionByWikiID(FindProductAuctionByWikiIDReq)
            this.showByIDs(res.IDs)
        } catch (err) {
            errors.HandleError(err)
        }
    }
}

productAuctionsPage.DisconnectedCallback = function () {
    this.resetCondition()
}

productAuctionsPage.showByIDs = async function (IDs) {
    const foundedContainerElement = document.getElementById('foundedContainer')
    for (let id of IDs) {
        try {
            // GetProductAuctionReq is the request structure of GetProductAuction()
            const GetProductAuctionReq = {
                "ID": id,
            }
            productAuction = await GetProductAuction(GetProductAuctionReq)
            productAuction.ID = id
            foundedContainerElement.insertAdjacentHTML('afterbegin', this.Templates["auction"](productAuction))
        } catch (err) {
            errors.HandleError(err)
        }
    }
}
