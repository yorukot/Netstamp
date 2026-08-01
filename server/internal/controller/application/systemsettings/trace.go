package systemsettings

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var systemSettingsTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/systemsettings")

var (
	attrSettingsAction   = attribute.Key("system_settings.action")
	attrSettingsResource = attribute.Key("system_settings.resource")
	attrSettingsActorID  = attribute.Key("user.id")
)
