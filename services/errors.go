/* For license and copyright information please see LEGAL file in ChaparKhane repository */

package services

import (
	er "../libgo/error"
	lang "../libgo/language"
)

// Errors
var (
	// Common
	ErrBadSituation = er.New().SetDetail(lang.EnglishLanguage, "Platform - Bad Situation",
		"Platform occur bad situation! Developers know about this error and fix it as soon as possible. Please try agin later.").Save()

	ErrBlockedByJustice = er.New().SetDetail(lang.EnglishLanguage, "Platform - Blocked By Justice",
		"Given request to register||update||delete a data on a record that blocked by justice department for some reason").Save()

	ErrBlockedPerson = er.New().SetDetail(lang.EnglishLanguage, "Platform - Blocked Person",
		"Given request from a person that blocked for some reason! So can't proccess the request").Save()

	// Authentication
	ErrBadPasswordOrOTP = er.New().SetDetail(lang.EnglishLanguage, "Platform - Bad Password or OTP",
		"Given person password or OTP is not valid and can't use for requested service").Save()

	// PersonNumber
	ErrPersonNumberRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Person Number Registered",
		"Given person number to register new person on platform already registered").Save()

	// OrganizationAuthentication
	ErrOrgNameRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Org Name Registered",
		"Given organization name to register new organization or update exiting organization already registered").Save()

	ErrOrgDomainRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Org Domain Registered",
		"Given organization domain to register new organization or update exiting organization already registered").Save()

	// Wiki
	ErrWikiTitleRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Wiki Title Registered",
		"Given wiki title to register already registered and active for other one!").Save()

	ErrWikiURIRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Wiki URI Registered",
		"Given wiki URI to register already registered and active for other one!").Save()

	// ProductAuction
	ErrProductAuctionRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Product Auction Registered",
		"Product auction in desire currency to register already registered and active! Please edit it for any changes").Save()

	ErrProductAuctionDefaultNotRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Product Auction Default Not Registered",
		"Default product auction in desire currency not register yet! So you can't register custom auction!").Save()

	ErrProductAuctionNotRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Product Auction Not Registered",
		"Desire product auction not register yet! So you can't update it!").Save()
)
