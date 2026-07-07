import "@typespec/http";
import "@typespec/openapi";

import "./models/common.tsp";
import "./models/admin.tsp";
import "./models/auth.tsp";
import "./models/user.tsp";
import "./models/project.tsp";
import "./models/label.tsp";
import "./models/check.tsp";
import "./models/probe.tsp";
import "./models/assignment.tsp";
import "./models/result.tsp";
import "./models/alert.tsp";
import "./models/public-status.tsp";

import "./services/system.tsp";
import "./services/admin.tsp";
import "./services/auth.tsp";
import "./services/users.tsp";
import "./services/projects.tsp";
import "./services/labels.tsp";
import "./services/checks.tsp";
import "./services/assignments.tsp";
import "./services/probes.tsp";
import "./services/results.tsp";
import "./services/alerts.tsp";
import "./services/public-status.tsp";
import "./services/probe-runtime.tsp";
import "./services/install.tsp";

using Http;
using OpenAPI;

@service(#{ title: "Netstamp API" })
@server("/api/v1", "Default v1 API base path")
@doc("HTTP API contract for the Netstamp controller and probe runtime.")
@info(#{ version: "0.1.0" })
@tagMetadata("System", #{ description: "Controller status and readiness endpoints." })
@tagMetadata("Admin", #{ description: "Instance-level administrator settings and bootstrap state." })
@tagMetadata("Auth", #{ description: "User registration, login, logout, and current-session access." })
@tagMetadata("Users", #{ description: "Authenticated user profile, email, and password management." })
@tagMetadata("Projects", #{ description: "Authenticated project lifecycle operations." })
@tagMetadata("Project Members", #{ description: "Project membership and role management." })
@tagMetadata("Project Invites", #{ description: "Pending project invitations and invite responses." })
@tagMetadata("Labels", #{ description: "Project-scoped key/value labels used by selectors." })
@tagMetadata("Checks", #{ description: "Ping, TCP, and traceroute check definitions." })
@tagMetadata("Assignments", #{ description: "Selector previews and effective probe-check assignments for console users." })
@tagMetadata("Probes", #{ description: "Project-scoped probe registration and secret management." })
@tagMetadata("Results", #{ description: "Authenticated result query endpoints." })
@tagMetadata("Alerts", #{ description: "Project-scoped alert rules, alert incidents, and notifications." })
@tagMetadata("Public Status Pages", #{ description: "Public, shareable project status pages and their ordered elements." })
@tagMetadata("Probe Runtime", #{ description: "Probe-agent runtime handshake, polling, and result ingestion." })
@tagMetadata("Install", #{ description: "Probe-agent install assets for Linux systemd hosts." })
namespace Netstamp;
