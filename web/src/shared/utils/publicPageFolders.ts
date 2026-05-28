import type { ApiPublicPageFolder } from "@/shared/api/types";

export function publicPageFolderLabel(folder: ApiPublicPageFolder, folders: ApiPublicPageFolder[]) {
	const names: string[] = [folder.name];
	let current = folder;
	const guard = new Set<string>([folder.id]);

	while (current.parentId) {
		const parent = folders.find(candidate => candidate.id === current.parentId);
		if (!parent || guard.has(parent.id)) {
			break;
		}

		names.unshift(parent.name);
		guard.add(parent.id);
		current = parent;
	}

	return names.join(" / ");
}

export function isPublicPageDescendantFolder(folder: ApiPublicPageFolder, ancestorID: string, folders: ApiPublicPageFolder[]) {
	let current = folder;
	const guard = new Set<string>([folder.id]);

	while (current.parentId) {
		if (current.parentId === ancestorID) {
			return true;
		}

		const parent = folders.find(candidate => candidate.id === current.parentId);
		if (!parent || guard.has(parent.id)) {
			return false;
		}

		guard.add(parent.id);
		current = parent;
	}

	return false;
}
