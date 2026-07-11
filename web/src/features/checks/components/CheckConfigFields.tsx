import {
	type HTTPConfigFormState,
	type HTTPConfigValidation,
	type HTTPStatusClass,
	type IPFamilyFormValue,
	type PingConfigFormState,
	type PingConfigValidation,
	type TCPConfigFormState,
	type TCPConfigValidation,
	type TracerouteConfigFormState,
	type TracerouteConfigValidation,
	type TracerouteProtocolFormValue
} from "@/features/checks/data/checkConfig";
import type { CheckType } from "@/features/checks/data/checks";
import { ActionRow, Button, Checkbox, FieldLabel, SelectField, TextAreaField, TextField } from "@netstamp/ui";
import styles from "./CheckConfigFields.module.css";

const ipFamilyOptions: Array<{ value: IPFamilyFormValue; label: string }> = [
	{ value: "", label: "Auto" },
	{ value: "inet", label: "IPv4" },
	{ value: "inet6", label: "IPv6" }
];

const tracerouteProtocolOptions: Array<{ value: TracerouteProtocolFormValue; label: string }> = [
	{ value: "icmp", label: "ICMP" },
	{ value: "udp", label: "UDP" }
];

interface CheckConfigFieldsProps {
	checkType: CheckType;
	disabled: boolean;
	pingConfig: PingConfigFormState;
	pingValidation: PingConfigValidation;
	tcpConfig: TCPConfigFormState;
	tcpValidation: TCPConfigValidation;
	tracerouteConfig: TracerouteConfigFormState;
	tracerouteValidation: TracerouteConfigValidation;
	httpConfig: HTTPConfigFormState;
	httpValidation: HTTPConfigValidation;
	onPingConfigChange: (patch: Partial<PingConfigFormState>) => void;
	onTCPConfigChange: (patch: Partial<TCPConfigFormState>) => void;
	onTracerouteConfigChange: (patch: Partial<TracerouteConfigFormState>) => void;
	onHTTPConfigChange: (patch: Partial<HTTPConfigFormState>) => void;
}

export function CheckConfigFields({
	checkType,
	disabled,
	pingConfig,
	pingValidation,
	tcpConfig,
	tcpValidation,
	tracerouteConfig,
	tracerouteValidation,
	httpConfig,
	httpValidation,
	onPingConfigChange,
	onTCPConfigChange,
	onTracerouteConfigChange,
	onHTTPConfigChange
}: CheckConfigFieldsProps) {
	const statusClasses: HTTPStatusClass[] = ["1xx", "2xx", "3xx", "4xx", "5xx"];
	function toggleStatusClass(value: HTTPStatusClass, checked: boolean) {
		onHTTPConfigChange({ statusClasses: checked ? [...httpConfig.statusClasses, value] : httpConfig.statusClasses.filter(item => item !== value) });
	}
	return (
		<div className={styles.checkConfigSection}>
			<FieldLabel>{checkType} config</FieldLabel>
			{checkType === "HTTP" ? (
				<div className={styles.checkConfigGrid}>
					<SelectField
						label="Method"
						value={httpConfig.method}
						disabled={disabled}
						onChange={event =>
							onHTTPConfigChange({ method: event.currentTarget.value as HTTPConfigFormState["method"], body: ["GET", "HEAD"].includes(event.currentTarget.value) ? "" : httpConfig.body })
						}
						options={["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"].map(value => ({ value, label: value }))}
					/>
					<TextField
						label="Timeout ms"
						type="number"
						min={1}
						max={60000}
						value={httpConfig.timeoutMs}
						disabled={disabled}
						error={httpValidation.timeout.error || undefined}
						onChange={event => onHTTPConfigChange({ timeoutMs: event.currentTarget.value })}
					/>
					<SelectField
						label="IP family"
						value={httpConfig.ipFamily}
						disabled={disabled}
						onChange={event => onHTTPConfigChange({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
						options={ipFamilyOptions}
					/>
					<div>
						<FieldLabel>Expected status classes</FieldLabel>
						<div className={styles.statusClassGrid}>
							{statusClasses.map(value => (
								<label key={value} className={styles.checkboxLabel}>
									<Checkbox checked={httpConfig.statusClasses.includes(value)} disabled={disabled} onChange={event => toggleStatusClass(value, event.currentTarget.checked)} />
									<span>{value}</span>
								</label>
							))}
						</div>
					</div>
					<TextField
						label="Exact status codes"
						value={httpConfig.statusCodes}
						disabled={disabled}
						placeholder="200, 202, 404"
						error={httpValidation.statusError || undefined}
						onChange={event => onHTTPConfigChange({ statusCodes: event.currentTarget.value })}
					/>
					<TextField
						label="Response contains"
						value={httpConfig.bodyContains}
						disabled={disabled}
						maxLength={1024}
						onChange={event => onHTTPConfigChange({ bodyContains: event.currentTarget.value })}
					/>
					<label className={styles.checkboxLabel}>
						<Checkbox checked={httpConfig.followRedirects} disabled={disabled} onChange={event => onHTTPConfigChange({ followRedirects: event.currentTarget.checked })} />
						<span>Follow redirects</span>
					</label>
					<label className={styles.checkboxLabel}>
						<Checkbox checked={httpConfig.skipTlsVerify} disabled={disabled} onChange={event => onHTTPConfigChange({ skipTlsVerify: event.currentTarget.checked })} />
						<span>Skip TLS verification</span>
					</label>
					<div className={styles.fullWidth}>
						<FieldLabel>Request headers</FieldLabel>
						<div className={styles.headerList}>
							{httpConfig.headers.map(header => (
								<div className={styles.headerRow} key={header.id}>
									<TextField
										label="Header name"
										aria-label="Header name"
										value={header.name}
										disabled={disabled}
										onChange={event => onHTTPConfigChange({ headers: httpConfig.headers.map(item => (item.id === header.id ? { ...item, name: event.currentTarget.value } : item)) })}
									/>
									<TextField
										label="Header value"
										aria-label="Header value"
										value={header.value}
										disabled={disabled}
										onChange={event => onHTTPConfigChange({ headers: httpConfig.headers.map(item => (item.id === header.id ? { ...item, value: event.currentTarget.value } : item)) })}
									/>
									<Button type="button" variant="secondary" disabled={disabled} onClick={() => onHTTPConfigChange({ headers: httpConfig.headers.filter(item => item.id !== header.id) })}>
										Remove
									</Button>
								</div>
							))}
						</div>
						<ActionRow>
							<Button
								type="button"
								variant="secondary"
								disabled={disabled || httpConfig.headers.length >= 50}
								onClick={() => onHTTPConfigChange({ headers: [...httpConfig.headers, { id: crypto.randomUUID(), name: "", value: "" }] })}
							>
								Add header
							</Button>
						</ActionRow>
					</div>
					{httpConfig.method !== "GET" && httpConfig.method !== "HEAD" ? (
						<TextAreaField
							className={styles.fullWidth}
							label="Request body"
							rows={6}
							value={httpConfig.body}
							disabled={disabled}
							maxLength={65536}
							error={httpValidation.bodyError || undefined}
							onChange={event => onHTTPConfigChange({ body: event.currentTarget.value })}
						/>
					) : null}
				</div>
			) : checkType === "Traceroute" ? (
				<div className={styles.checkConfigGrid}>
					<SelectField
						label="Protocol"
						value={tracerouteConfig.protocol}
						disabled={disabled}
						onChange={event => onTracerouteConfigChange({ protocol: event.currentTarget.value as TracerouteProtocolFormValue })}
						options={tracerouteProtocolOptions}
					/>
					<TextField
						label="Max hops"
						type="number"
						min={1}
						max={64}
						step={1}
						inputMode="numeric"
						value={tracerouteConfig.maxHops}
						disabled={disabled}
						error={tracerouteValidation.maxHops.error || undefined}
						onChange={event => onTracerouteConfigChange({ maxHops: event.currentTarget.value })}
					/>
					<TextField
						label="Timeout ms"
						type="number"
						min={1}
						max={60000}
						step={1}
						inputMode="numeric"
						value={tracerouteConfig.timeoutMs}
						disabled={disabled}
						error={tracerouteValidation.timeoutMs.error || undefined}
						onChange={event => onTracerouteConfigChange({ timeoutMs: event.currentTarget.value })}
					/>
					<TextField
						label="Queries per hop"
						type="number"
						min={1}
						max={10}
						step={1}
						inputMode="numeric"
						value={tracerouteConfig.queriesPerHop}
						disabled={disabled}
						error={tracerouteValidation.queriesPerHop.error || undefined}
						onChange={event => onTracerouteConfigChange({ queriesPerHop: event.currentTarget.value })}
					/>
					<TextField
						label="Packet size bytes"
						type="number"
						min={1}
						max={65507}
						step={1}
						inputMode="numeric"
						value={tracerouteConfig.packetSizeBytes}
						disabled={disabled}
						error={tracerouteValidation.packetSizeBytes.error || undefined}
						onChange={event => onTracerouteConfigChange({ packetSizeBytes: event.currentTarget.value })}
					/>
					{tracerouteConfig.protocol === "udp" ? (
						<TextField
							label="Port"
							type="number"
							min={1}
							max={65535}
							step={1}
							inputMode="numeric"
							value={tracerouteConfig.port}
							disabled={disabled}
							error={tracerouteValidation.port.error || undefined}
							onChange={event => onTracerouteConfigChange({ port: event.currentTarget.value })}
						/>
					) : null}
					<SelectField
						label="IP family"
						value={tracerouteConfig.ipFamily}
						disabled={disabled}
						onChange={event => onTracerouteConfigChange({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
						options={ipFamilyOptions}
					/>
				</div>
			) : checkType === "TCP" ? (
				<div className={styles.checkConfigGrid}>
					<TextField
						label="Port"
						type="number"
						min={1}
						max={65535}
						step={1}
						inputMode="numeric"
						value={tcpConfig.port}
						disabled={disabled}
						error={tcpValidation.port.error || undefined}
						onChange={event => onTCPConfigChange({ port: event.currentTarget.value })}
					/>
					<TextField
						label="Timeout ms"
						type="number"
						min={1}
						step={1}
						inputMode="numeric"
						value={tcpConfig.timeoutMs}
						disabled={disabled}
						error={tcpValidation.timeoutMs.error || undefined}
						onChange={event => onTCPConfigChange({ timeoutMs: event.currentTarget.value })}
					/>
					<SelectField
						label="IP family"
						value={tcpConfig.ipFamily}
						disabled={disabled}
						onChange={event => onTCPConfigChange({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
						options={ipFamilyOptions}
					/>
				</div>
			) : (
				<div className={styles.checkConfigGrid}>
					<TextField
						label="Packet count"
						type="number"
						min={1}
						step={1}
						inputMode="numeric"
						value={pingConfig.packetCount}
						disabled={disabled}
						error={pingValidation.packetCount.error || undefined}
						onChange={event => onPingConfigChange({ packetCount: event.currentTarget.value })}
					/>
					<TextField
						label="Packet size bytes"
						type="number"
						min={1}
						max={65507}
						step={1}
						inputMode="numeric"
						value={pingConfig.packetSizeBytes}
						disabled={disabled}
						error={pingValidation.packetSizeBytes.error || undefined}
						onChange={event => onPingConfigChange({ packetSizeBytes: event.currentTarget.value })}
					/>
					<TextField
						label="Timeout ms"
						type="number"
						min={1}
						step={1}
						inputMode="numeric"
						value={pingConfig.timeoutMs}
						disabled={disabled}
						error={pingValidation.timeoutMs.error || undefined}
						onChange={event => onPingConfigChange({ timeoutMs: event.currentTarget.value })}
					/>
					<SelectField
						label="IP family"
						value={pingConfig.ipFamily}
						disabled={disabled}
						onChange={event => onPingConfigChange({ ipFamily: event.currentTarget.value as IPFamilyFormValue })}
						options={ipFamilyOptions}
					/>
				</div>
			)}
		</div>
	);
}
