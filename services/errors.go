/* For license and copyright information please see LEGAL file in ChaparKhane repository */

package services

import (
	er "../libgo/error"
	lang "../libgo/language"
)

// Errors
var (
	// Common
	ErrPlatformBadSituation = er.New().SetDetail(lang.EnglishLanguage, "Platform - Bad Situation",
		"Platform occur bad situation! Developers know about this error and fix it as soon as possible. Please try agin later.").Save()

	ErrPlatformBlockedByJustice = er.New().SetDetail(lang.EnglishLanguage, "Platform - Blocked By Justice",
		"Given request to register||update||delete a data on a record that blocked by justice department for some reason").Save()

	// Authentication
	ErrPlatformBadPasswordOrOTP = er.New().SetDetail(lang.EnglishLanguage, "Platform - Bad Password or OTP",
		"Given person password or OTP is not valid and can't use for requested service").Save()

	// PersonNumber
	ErrPlatformPersonNumberRegistered = er.New().SetDetail(lang.EnglishLanguage, "Platform - Person Number Registered",
		"Given person number to register new person on platform already registered").Save()
)
