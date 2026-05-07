import { classNames } from "@/shared/utils/classNames";
import { Button, FieldLabel, TextField } from "@netstamp/ui";
import styles from "./NewProbeDrawer.module.css";

interface TagPickerProps {
	tagOptions: string[];
	selectedTags: string[];
	newTag: string;
	disabled?: boolean;
	onToggleTag: (tag: string) => void;
	onNewTagChange: (value: string) => void;
	onAddTag: () => void;
}

export function TagPicker({ tagOptions, selectedTags, newTag, disabled, onToggleTag, onNewTagChange, onAddTag }: TagPickerProps) {
	return (
		<div className={styles.tagPicker}>
			<FieldLabel>Tags</FieldLabel>
			<div className={styles.tagCloud}>
				{tagOptions.map(tag => (
					<Button variant="plain" className={classNames(styles.tagButton, selectedTags.includes(tag) && styles.tagSelected)} key={tag} type="button" disabled={disabled} onClick={() => onToggleTag(tag)}>
						{tag}
					</Button>
				))}
			</div>
			<div className={styles.tagCreate}>
				<TextField
					label="Create tag"
					value={newTag}
					placeholder="backbone"
					disabled={disabled}
					onChange={event => onNewTagChange(event.currentTarget.value)}
					onKeyDown={event => {
						if (event.key === "Enter") {
							event.preventDefault();
							onAddTag();
						}
					}}
				/>
				<Button type="button" variant="outline" disabled={disabled} onClick={onAddTag}>
					Add tag
				</Button>
			</div>
		</div>
	);
}
