import { type IPFamilyFormValue, type PingConfigFormState, type TCPConfigFormState, type TracerouteConfigFormState, type TracerouteProtocolFormValue } from "@/features/checks/data/checkConfig";
import type { CheckType } from "@/features/checks/data/checks";
import { FieldLabel, SelectField, TextField } from "@netstamp/ui";
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
	tcpConfig: TCPConfigFormState;
	tracerouteConfig: TracerouteConfigFormState;
	onPingConfigChange: (patch: Partial<PingConfigFormState>) => void;
	onTCPConfigChange: (patch: Partial<TCPConfigFormState>) => void;
	onTracerouteConfigChange: (patch: Partial<TracerouteConfigFormState>) => void;
}

export function CheckConfigFields({ checkType, disabled, pingConfig, tcpConfig, tracerouteConfig, onPingConfigChange, onTCPConfigChange, onTracerouteConfigChange }: CheckConfigFieldsProps) {
	return (
		<div className={styles.checkConfigSection}>
			<FieldLabel>{checkType} config</FieldLabel>
			{checkType === "Traceroute" ? (
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
