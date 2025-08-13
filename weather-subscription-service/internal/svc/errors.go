package svc

import "errors"

var (
	ErrorNotFound      = errors.New("resource not found")
	ErrorAlreadyExists = errors.New("resource already exists")
	ErrorTokenNotFound = errors.New("token not found")

	ErrorCacheMiss = errors.New("key not found")

	ErrorGetCityWeather = errors.New("unable to get weather")

	ErrorSubscriptionCreate      = errors.New("subscription create")
	ErrorSubscriptionConfirm     = errors.New("subscription confirm")
	ErrorSubscriptionUnsubscribe = errors.New("subscription unsubscribe")
	ErrorSendEmail               = errors.New("send email")
)
