/* For license and copyright information please see LEGAL file in ChaparKhane repository */
// Auto-generated, edits will be overwritten

package services

import (
	"crypto/rand"

	"../libgo/achaemenid"
	"../libgo/captcha"
	er "../libgo/error"
	"../libgo/log"
	"../libgo/sdk/asanak.com"
	sep "../libgo/sdk/sep.ir"
)

var (
	phraseCaptchas = captcha.NewDefaultPhraseCaptchas()

	smsProvider       asanak.Asanak
	smsOTPSecurityKey = make([]byte, 32)

	sepPOS sep.POS

	adminUserID = [32]byte{128}
)

func init() {
	var err *er.Error
	var goErr error

	var asanakJSON = achaemenid.Server.Assets.Secret.GetFile("asanak.com.json")
	if asanakJSON == nil {
		log.Fatal("Can't find 'asanak.com.json' file in 'secret' folder in top of repository")
	}
	smsProvider.Init(asanakJSON.Data)

	_, goErr = rand.Read(smsOTPSecurityKey)
	// Note that goErr == nil only if we read len(SecurityKey) bytes.
	if goErr != nil {
		log.Fatal(goErr)
	}

	var sepJSON = achaemenid.Server.Assets.Secret.GetFile("sep.ir-pos.json")
	err = sepPOS.Init(sepJSON.Data)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// PersonAuthentication
	achaemenid.Server.Services.RegisterService(&registerPersonService)
	achaemenid.Server.Services.RegisterService(&changePersonPasswordService)
	achaemenid.Server.Services.RegisterService(&blockPersonService)
	achaemenid.Server.Services.RegisterService(&recoverPersonAccountService)
	achaemenid.Server.Services.RegisterService(&revokePersonPublicKeyService)
	achaemenid.Server.Services.RegisterService(&unblockPersonService)
	achaemenid.Server.Services.RegisterService(&getPersonStatusService)

	// PersonNumber
	achaemenid.Server.Services.RegisterService(&registerPersonNumberService)
	achaemenid.Server.Services.RegisterService(&getPersonNumberStatusService)
	achaemenid.Server.Services.RegisterService(&getPersonNumberService)
	// achaemenid.Server.Services.RegisterService(&)

	// UserAppConnection
	achaemenid.Server.Services.RegisterService(&authenticateAppConnectionService)
	achaemenid.Server.Services.RegisterService(&findUserAppConnectionService)
	achaemenid.Server.Services.RegisterService(&getUserAppConnectionService)
	achaemenid.Server.Services.RegisterService(&findUserAppConnectionByGivenDelegateService)
	achaemenid.Server.Services.RegisterService(&findUserAppConnectionByGottenDelegateService)
	achaemenid.Server.Services.RegisterService(&registerUserAppConnectionService)
	achaemenid.Server.Services.RegisterService(&updateUserAppConnectionService)
	// achaemenid.Server.Services.RegisterService(&)

	// PersonPublicKey
	achaemenid.Server.Services.RegisterService(&approvePersonPublicKeyService)
	achaemenid.Server.Services.RegisterService(&authenticatePersonPublicKeyService)

	// OrganizationAuthentication
	achaemenid.Server.Services.RegisterService(&registerNewOrganizationService)
	achaemenid.Server.Services.RegisterService(&updateOrganizationService)
	achaemenid.Server.Services.RegisterService(&getOrganizationService)
	achaemenid.Server.Services.RegisterService(&getLastOrganizationsIDService)
	// achaemenid.Server.Services.RegisterService(&blockOrgService)
	// achaemenid.Server.Services.RegisterService(&)

	// Quiddity
	achaemenid.Server.Services.RegisterService(&registerQuiddityService)
	achaemenid.Server.Services.RegisterService(&registerQuiddityNewLanguageService)
	achaemenid.Server.Services.RegisterService(&updateQuiddityService)
	achaemenid.Server.Services.RegisterService(&getQuiddityService)
	achaemenid.Server.Services.RegisterService(&findQuiddityByTitleService)
	achaemenid.Server.Services.RegisterService(&findQuiddityByURIService)
	achaemenid.Server.Services.RegisterService(&findQuiddityByOrgIDService)
	achaemenid.Server.Services.RegisterService(&getQuiddityLanguagesService)

	// ProductAuction
	achaemenid.Server.Services.RegisterService(&registerDefaultProductAuctionService)
	achaemenid.Server.Services.RegisterService(&registerCustomProductAuctionService)
	achaemenid.Server.Services.RegisterService(&updateProductAuctionService)
	achaemenid.Server.Services.RegisterService(&getProductAuctionService)
	achaemenid.Server.Services.RegisterService(&findProductAuctionByDistributionCenterIDService)
	achaemenid.Server.Services.RegisterService(&findProductAuctionByGroupIDService)
	achaemenid.Server.Services.RegisterService(&findProductAuctionByOrgIDService)
	achaemenid.Server.Services.RegisterService(&findProductAuctionByQuiddityIDDistributionCenterIDService)
	achaemenid.Server.Services.RegisterService(&findProductAuctionByQuiddityIDGroupIDService)
	achaemenid.Server.Services.RegisterService(&findProductAuctionByQuiddityIDService)

	// ProductPrice
	achaemenid.Server.Services.RegisterService(&registerProductPriceService)
	achaemenid.Server.Services.RegisterService(&updateProductPriceService)
	achaemenid.Server.Services.RegisterService(&getProductPriceService)
	achaemenid.Server.Services.RegisterService(&findProductPriceByOrgIDService)
	// achaemenid.Server.Services.RegisterService(&)

	// Product
	achaemenid.Server.Services.RegisterService(&registerProductService)
	// achaemenid.Server.Services.RegisterService(&approveProductAuctionByWarehouseService)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)

	// FinancialTransaction
	achaemenid.Server.Services.RegisterService(&registerFinancialTransactionService)
	achaemenid.Server.Services.RegisterService(&getFinancialTransactionService)
	achaemenid.Server.Services.RegisterService(&findFinancialTransactionByDayService)

	// ForeignDetail
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)

	// OrganizationStaff
	// achaemenid.Server.Services.RegisterService(&approveOrgPositionByPersonService)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)
	// achaemenid.Server.Services.RegisterService(&)

	// Common Services
	achaemenid.Server.Services.RegisterService(&getNewPhraseCaptchaService)
	achaemenid.Server.Services.RegisterService(&getPhraseCaptchaAudioService)
	achaemenid.Server.Services.RegisterService(&solvePhraseCaptchaService)
	achaemenid.Server.Services.RegisterService(&sendOtpService)
}
