package notify

import appnotification "github.com/yorukot/netstamp/internal/controller/application/notification"

func retryable(kind, code, message string) appnotification.DeliveryResult {
	return appnotification.DeliveryResult{Retryable: true, Kind: kind, Code: code, Message: message}
}

func permanent(kind, code, message string) appnotification.DeliveryResult {
	return appnotification.DeliveryResult{Retryable: false, Kind: kind, Code: code, Message: message}
}
