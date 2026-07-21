package usecase

import (
	"idas-video/internal/usecase/inbound"
	"idas-video/internal/usecase/outbound"
)

var _ inbound.LoginUsecase = (*AuthUsecase)(nil)
var _ inbound.UserAuthenticator = (*AuthUsecase)(nil)
var _ inbound.ActiveSubscriptionUsecase = (*SubscriptionUsecase)(nil)
var _ inbound.SubscriptionPurchaseUsecase = (*SubscriptionUsecase)(nil)
var _ inbound.TransactionReaderUsecase = (*TransactionUsecase)(nil)
var _ inbound.VideoAccessUsecase = (*VideoUsecase)(nil)
var _ outbound.SubscriptionActivator = (*SubscriptionUsecase)(nil)
