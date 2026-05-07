import type { Material, Object3D } from "three";

interface DisposableObject extends Object3D {
	geometry?: { dispose: () => void };
	material?: Material | Material[];
}

export function disposeSceneObjects(root: Object3D) {
	root.traverse(object => {
		const disposable = object as DisposableObject;
		disposable.geometry?.dispose();

		if (Array.isArray(disposable.material)) {
			disposable.material.forEach(material => material.dispose());
			return;
		}

		disposable.material?.dispose();
	});
}
