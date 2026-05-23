import "@typespec/http";
import "@typespec/openapi";

using Http;
using OpenAPI;

namespace Netstamp;

alias ShellScriptFile = File<"text/x-shellscript", string>;
alias AgentBinaryFile = File<"application/octet-stream", bytes>;

@tag("Install")
@route("/install")
interface InstallService {
	@operationId("getAgentInstallerScript")
	@summary("Get Linux probe agent installer script")
	@get
	@route("/agent.sh")
	getAgentInstallerScript(): ShellScriptFile | InternalServerProblem;

	@operationId("getAgentUninstallerScript")
	@summary("Get Linux probe agent uninstaller script")
	@get
	@route("/uninstall-agent.sh")
	getAgentUninstallerScript(): ShellScriptFile | InternalServerProblem;

	@operationId("downloadAgentLinuxAmd64")
	@summary("Download Linux amd64 probe agent binary")
	@get
	@route("/netstamp-agent-linux-amd64")
	downloadAgentLinuxAmd64(): AgentBinaryFile | NotFoundProblem | ServiceUnavailableProblem;

	@operationId("downloadAgentLinuxArm64")
	@summary("Download Linux arm64 probe agent binary")
	@get
	@route("/netstamp-agent-linux-arm64")
	downloadAgentLinuxArm64(): AgentBinaryFile | NotFoundProblem | ServiceUnavailableProblem;
}
