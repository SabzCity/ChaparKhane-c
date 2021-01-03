/* For license and copyright information please see LEGAL file in ChaparKhane repository */

package services

import (
	er "../libgo/error"
	lang "../libgo/language"
)

const errorEnglishDomain = "Sabz.City"
const errorPersianDomain = "شهرسبز"

// Errors
var (
	// Common
	ErrNotImplemented = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Not Implemented",
		"Requested service or part of it registered but not implemented yet. Please try agin later.").Save()

	ErrBadSituation = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Bad Situation",
		"Platform occur bad situation! Developers know about this error and fix it as soon as possible. Please try agin later.").Save()

	ErrBlockedByJustice = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Blocked By Justice",
		"Given request to register||update||delete a data on a record that blocked by justice department for some reason").Save()

	ErrBlockedPerson = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Blocked Person",
		"Given request from a person that blocked for some reason! So can't proccess the request").Save()

	// Authentication
	ErrBadPasswordOrOTP = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Bad Password or OTP",
		"Given person password or OTP is not valid and can't use for requested service").Save()

	// PersonNumber
	ErrPersonNumberRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Person Number Registered",
		"Given person number to register new person on platform already registered").Save()

	// OrganizationAuthentication
	ErrOrgNameRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Org Name Registered",
		"Given organization name to register new organization or update exiting organization already registered").Save()

	ErrOrgDomainRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Org Domain Registered",
		"Given organization domain to register new organization or update exiting organization already registered").Save()

	// Quiddity
	ErrQuiddityTitleRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Quiddity Title Registered",
		"Given quiddity title to register already registered and active for other one!").Save()

	ErrQuiddityURIRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Quiddity URI Registered",
		"Given quiddity URI to register already registered and active for other one!").Save()

	// ProductAuction
	ErrProductAuctionRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Product Auction Registered",
		"Product auction already registered and active! Please edit it for any changes").Save()

	ErrProductAuctionDefaultNotRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Product Auction Default Not Registered",
		"Default product auction not register yet! So you can't register custom auction!").Save()

	ErrProductAuctionNotRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Product Auction Not Registered",
		"Desire product auction not register yet! So you can't update it!").Save()

	// ProductPrice
	ErrProductPriceNotRegistered = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Product Price Not Registered",
		"Product price not register yet! So you can't update it!").Save()

	// FinancialTransaction
	ErrFinancialTransactionBadSociety = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Financial Transaction Bad Society",
		"Can't proccess request that from and to other different societies!").Save()

	ErrFinancialTransactionBadUser = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Financial Transaction Bad User",
		"Can't proccess request that from or to user ID is not same with active user!").Save()

	ErrFinancialTransactionBalance = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Financial Transaction Balance",
		"Financial transaction canceled due to user don't has enough balance").Save()

	// Product
	ErrProductInvoiceDelegate = er.New().SetDetail(lang.LanguageEnglish, errorEnglishDomain, "Delegate Product Invoice",
		"User of the connection can't register delegate invoice without send valid OTP or Transaction ID of desire user").Save()
)
