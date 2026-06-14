import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

@tag("Alerts")
@useAuth(SessionCookieAuth)
@route("/projects/{ref}/alerts")
interface AlertService {
	@operationId("listProjectAlertRules")
	@summary("List project alert rules")
	@get
	@route("/rules")
	listProjectAlertRules(...ProjectRefParam, ...AlertRuleListQuery):
		| OkJson<AlertRuleListResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("createProjectAlertRule")
	@summary("Create project alert rule")
	@post
	@route("/rules")
	createProjectAlertRule(...ProjectRefParam, @body body: CreateAlertRuleRequest):
		| CreatedJson<AlertRuleResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectAlertRule")
	@summary("Get project alert rule")
	@get
	@route("/rules/{rule_id}")
	getProjectAlertRule(...ProjectRefParam, ...AlertRuleIdPathParam): OkJson<AlertRuleResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("updateProjectAlertRule")
	@summary("Update project alert rule")
	@patch
	@route("/rules/{rule_id}")
	updateProjectAlertRule(...ProjectRefParam, ...AlertRuleIdPathParam, @body body: UpdateAlertRuleRequest):
		| OkJson<AlertRuleResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectAlertRule")
	@summary("Delete project alert rule")
	@delete
	@route("/rules/{rule_id}")
	deleteProjectAlertRule(...ProjectRefParam, ...AlertRuleIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("listProjectAlertIncidents")
	@summary("List project alert incidents")
	@get
	@route("/incidents")
	listProjectAlertIncidents(...ProjectRefParam, ...AlertIncidentListQuery):
		| OkJson<AlertIncidentListResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectAlertIncident")
	@summary("Get project alert incident")
	@get
	@route("/incidents/{incident_id}")
	getProjectAlertIncident(...ProjectRefParam, ...AlertIncidentIdPathParam): OkJson<AlertIncidentResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("listProjectNotificationChannels")
	@summary("List project notification channels")
	@get
	@route("/channels")
	listProjectNotificationChannels(...ProjectRefParam): OkJson<NotificationChannelListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("createProjectNotificationChannel")
	@summary("Create project notification channel")
	@post
	@route("/channels")
	createProjectNotificationChannel(...ProjectRefParam, @body body: CreateNotificationChannelRequest):
		| CreatedJson<NotificationChannelResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectNotificationChannel")
	@summary("Get project notification channel")
	@get
	@route("/channels/{channel_id}")
	getProjectNotificationChannel(...ProjectRefParam, ...NotificationChannelIdPathParam):
		| OkJson<NotificationChannelResponse>
		| UnauthorizedProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("updateProjectNotificationChannel")
	@summary("Update project notification channel")
	@patch
	@route("/channels/{channel_id}")
	updateProjectNotificationChannel(...ProjectRefParam, ...NotificationChannelIdPathParam, @body body: UpdateNotificationChannelRequest):
		| OkJson<NotificationChannelResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectNotificationChannel")
	@summary("Delete project notification channel")
	@delete
	@route("/channels/{channel_id}")
	deleteProjectNotificationChannel(...ProjectRefParam, ...NotificationChannelIdPathParam):
		| NoContent
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;
}
