import "@typespec/http";
import "@typespec/openapi";

import "./models/common.tsp";
import "./models/auth.tsp";
import "./models/user.tsp";
import "./models/project.tsp";
import "./models/label.tsp";
import "./models/check.tsp";
import "./models/probe.tsp";
import "./models/assignment.tsp";
import "./models/result.tsp";

import "./services/system.tsp";
import "./services/auth.tsp";
import "./services/users.tsp";
import "./services/projects.tsp";
import "./services/labels.tsp";
import "./services/checks.tsp";
import "./services/assignments.tsp";
import "./services/probes.tsp";
import "./services/results.tsp";
import "./services/probe-runtime.tsp";

using Http;
using OpenAPI;

@service(#{ title: "Netstamp API" })
@server("/api/v1", "Default v1 API base path")
@doc("HTTP API contract for the Netstamp controller and probe runtime.")
@info(#{ version: "0.1.0" })
@tagMetadata("System", #{ description: "Controller status and readiness endpoints." })
@tagMetadata("Auth", #{ description: "User registration, login, logout, and current-session access." })
@tagMetadata("Users", #{ description: "Authenticated user profile, email, and password management." })
@tagMetadata("Projects", #{ description: "Authenticated project lifecycle operations." })
@tagMetadata("Project Members", #{ description: "Project membership and role management." })
@tagMetadata("Labels", #{ description: "Project-scoped key/value labels used by selectors." })
@tagMetadata("Checks", #{ description: "Ping and traceroute check definitions." })
@tagMetadata("Assignments", #{ description: "Selector previews and effective probe-check assignments for console users." })
@tagMetadata("Probes", #{ description: "Project-scoped probe registration and secret management." })
@tagMetadata("Results", #{ description: "Authenticated result query endpoints." })
@tagMetadata("Measurements", #{ description: "Raw-ish measurement rows for tables, logs, and recent activity views." })
@tagMetadata("Probe Runtime", #{ description: "Probe-agent runtime handshake, polling, and result ingestion." })
namespace Netstamp;
