import { formatDateTime, publicStatusPath } from "@/features/status-pages/api/statusPageAdapters";
import { pathForStatusPageEditor } from "@/routes/routePaths";
import { projectQueries } from "@/shared/api/queries";
import type { ApiPublicStatusPage } from "@/shared/api/types";
import { useCurrentProject } from "@/shared/api/useCurrentProject";
import { PageStack } from "@/shared/components/PageStack";
import { ScreenHeader } from "@/shared/components/ScreenHeader";
import { pushErrorToast, pushToast } from "@/shared/toast/toastStore";
import { Badge, Button, DataTable, Spinner, type DataColumn } from "@netstamp/ui";
import { CopyIcon } from "@phosphor-icons/react/dist/csr/Copy";
import { PlusIcon } from "@phosphor-icons/react/dist/csr/Plus";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import styles from "./StatusPagesPage.module.css";

const emptyPages: ApiPublicStatusPage[] = [];

function absolutePublicStatusURL(slug: string) {
	return new URL(publicStatusPath(slug), window.location.origin).toString();
}

export function StatusPagesPage() {
	const { t } = useTranslation("status");
	const { projectRef } = useCurrentProject();
	const pagesQuery = useQuery({
		...projectQueries.statusPages(projectRef || ""),
		enabled: Boolean(projectRef)
	});
	const pages = pagesQuery.data?.pages ?? emptyPages;

	async function copyPageLink(page: ApiPublicStatusPage) {
		try {
			await navigator.clipboard.writeText(absolutePublicStatusURL(page.slug));
			pushToast({ title: t("list.copied"), message: page.title, tone: "success" });
		} catch {
			pushErrorToast(t("list.copyError"));
		}
	}

	const columns: DataColumn<ApiPublicStatusPage>[] = [
		{
			key: "title",
			label: t("list.pageTitle"),
			sortable: true,
			sortValue: row => row.title,
			render: row => <strong className={styles.titleCell}>{row.title}</strong>
		},
		{
			key: "slug",
			label: t("list.slug"),
			sortable: true,
			sortValue: row => row.slug,
			render: row => <span className={styles.slugCell}>/status/{row.slug}</span>
		},
		{
			key: "visibility",
			label: t("list.status"),
			sortable: true,
			sortValue: row => (row.enabled ? 1 : 0),
			render: row => <Badge tone={row.enabled ? "success" : "neutral"}>{row.enabled ? t("list.public") : t("list.private")}</Badge>
		},
		{
			key: "updatedAt",
			label: t("list.modified"),
			sortable: true,
			sortValue: row => Date.parse(row.updatedAt),
			render: row => <time className={styles.timeCell}>{formatDateTime(row.updatedAt)}</time>
		}
	];

	return (
		<PageStack>
			<ScreenHeader
				title={t("list.title")}
				actions={
					<Button asChild>
						<Link to={pathForStatusPageEditor(projectRef)}>
							<PlusIcon aria-hidden="true" focusable="false" />
							{t("list.new")}
						</Link>
					</Button>
				}
			/>

			{pagesQuery.isPending ? (
				<Spinner label={t("list.loading")} layout="panel" size="lg" />
			) : (
				<DataTable
					columns={columns}
					rows={pages}
					density="compact"
					minWidth="52rem"
					ariaLabel={t("list.aria")}
					getRowKey={row => row.id}
					emptyLabel={t("list.empty")}
					rowActions={page => (
						<div className={styles.rowActions}>
							<Button asChild variant="outline" size="sm">
								<Link to={pathForStatusPageEditor(projectRef, page.id)}>{t("list.edit")}</Link>
							</Button>
							<Button type="button" variant="ghost" size="sm" onClick={() => void copyPageLink(page)}>
								<CopyIcon aria-hidden="true" focusable="false" />
								{t("list.copy")}
							</Button>
							<Button asChild variant="secondary" size="sm">
								<a href={publicStatusPath(page.slug)} target="_blank" rel="noreferrer">
									{t("list.view")}
								</a>
							</Button>
						</div>
					)}
				/>
			)}
		</PageStack>
	);
}
