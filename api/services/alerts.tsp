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

	@operationId("listProjectNotifications")
	@summary("List project notifications")
	@get
	@route("/notifications")
	listProjectNotifications(...ProjectRefParam): OkJson<NotificationListResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("createProjectNotification")
	@summary("Create project notification")
	@post
	@route("/notifications")
	createProjectNotification(...ProjectRefParam, @body body: CreateNotificationRequest):
		| CreatedJson<NotificationResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("getProjectNotification")
	@summary("Get project notification")
	@get
	@route("/notifications/{notification_id}")
	getProjectNotification(...ProjectRefParam, ...NotificationIdPathParam): OkJson<NotificationResponse> | UnauthorizedProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;

	@operationId("updateProjectNotification")
	@summary("Update project notification")
	@patch
	@route("/notifications/{notification_id}")
	updateProjectNotification(...ProjectRefParam, ...NotificationIdPathParam, @body body: UpdateNotificationRequest):
		| OkJson<NotificationResponse>
		| BadRequestProblem
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("testProjectNotification")
	@summary("Send a test notification through a project notification")
	@post
	@route("/notifications/{notification_id}/test")
	testProjectNotification(...ProjectRefParam, ...NotificationIdPathParam):
		| OkJson<NotificationTestResponse>
		| UnauthorizedProblem
		| ForbiddenProblem
		| NotFoundProblem
		| UnprocessableEntityProblem
		| InternalServerProblem;

	@operationId("deleteProjectNotification")
	@summary("Delete project notification")
	@delete
	@route("/notifications/{notification_id}")
	deleteProjectNotification(...ProjectRefParam, ...NotificationIdPathParam): NoContent | UnauthorizedProblem | ForbiddenProblem | NotFoundProblem | UnprocessableEntityProblem | InternalServerProblem;
}
